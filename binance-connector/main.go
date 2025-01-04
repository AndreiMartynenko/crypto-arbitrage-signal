package main

import (
	"encoding/json" // We need this to convert JSON data to/from Go structs
	"fmt"           // For string formatting
	"log"           // For logging errors and info
	"net/http"      // For making requests and creating an HTTP server
	"os"            // To read environment variables
	"strconv"       // Converting string to float for bid/ask prices
	"sync"          // Provides sync.RWMutex for concurrency-safe data access
	"time"          // For time.Duration, sleeping, and timestamps
)

// BinanceBookTicker represents the structure of the JSON response
// from Binance's /api/v3/ticker/bookTicker endpoint.
// We only care about these specific fields for now.
type BinanceBookTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
	// There are additional fields in the real API response, but we only
	// define what we need to parse here.
}

// TickerData holds the internal representation of the latest bid/ask
// we fetched from Binance. This is what we'll return to other services.
type TickerData struct {
	Symbol     string    `json:"symbol"`      // e.g. "BTCUSDT"
	Bid        float64   `json:"bid"`         // Parsed bid price as float
	Ask        float64   `json:"ask"`         // Parsed ask price as float
	LastUpdate time.Time `json:"last_update"` // Timestamp of the most recent fetch
}

// We define global variables here for simplicity. In production, you might
// want a more sophisticated approach, but this is fine for a connector service.
var (
	// tickerMutex ensures that reading/writing tickerInfo is thread-safe
	// when multiple goroutines are involved (e.g., polling + HTTP handler).
	tickerMutex sync.RWMutex

	// tickerInfo holds the most recent TickerData. We update it periodically
	// in a background loop, and return it to users via the HTTP endpoint.
	tickerInfo = TickerData{}
)

// main is the entry point of our binance-connector service.
func main() {
	// 1. Read environment variables for configuration (keys, pair, poll interval).
	binanceAPIKey := os.Getenv("BINANCE_API_KEY") // Used if we need private endpoints
	binanceAPISecret := os.Getenv("BINANCE_API_SECRET")
	pairs := os.Getenv("PAIRS") // Typically "BTC/USDT" or "ETH/USDT", etc.
	if pairs == "" {
		pairs = "BTC/USDT" // Default if none provided
	}

	// This tells how frequently we'll poll Binance for new data.
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = "5s" // Default to 5 seconds
	}
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		log.Printf("Invalid POLL_INTERVAL format, defaulting to 5s: %v", err)
		pollInterval = 5 * time.Second
	}

	// formatSymbol converts "BTC/USDT" -> "BTCUSDT" for Binance's API format.
	binanceSymbol := formatSymbol(pairs)

	log.Printf("Starting binance-connector for symbol=%s, pollInterval=%v\n",
		binanceSymbol, pollInterval)

	// 2. Start a background goroutine that continuously fetches data from Binance.
	//    This goroutine runs independently of the main thread.
	go func() {
		for {
			// Attempt to fetch the latest bid/ask from Binance
			err := fetchBinanceData(binanceSymbol, binanceAPIKey, binanceAPISecret)
			if err != nil {
				log.Printf("Error fetching data from Binance: %v", err)
			}
			// Sleep for pollInterval before fetching again
			time.Sleep(pollInterval)
		}
	}()

	// 3. Expose an HTTP endpoint (/latest-price) to retrieve the current data.
	http.HandleFunc("/latest-price", handleLatestPrice)

	// 4. Start the HTTP server on port 8001. This will block until an error or shutdown.
	port := "8001"
	log.Printf("HTTP server running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

// formatSymbol is a helper that just ensures "BTC/USDT" -> "BTCUSDT"
// This is what Binance expects in its 'symbol' query parameter.
func formatSymbol(pair string) string {
	return fmt.Sprintf("%s%s", stripSlash(pair))
}

// stripSlash removes "/" from a string (e.g., "BTC/USDT" -> "BTCUSDT").
func stripSlash(pair string) string {
	return stringReplace(pair, "/", "")
}

// stringReplace is a basic function that replaces all occurrences of `old` with `new`.
// We could also use strings.ReplaceAll(pair, "/", "") for brevity.
func stringReplace(str, old, new string) string {
	result := ""
	for _, ch := range str {
		if string(ch) == old {
			result += new
		} else {
			result += string(ch)
		}
	}
	return result
}

// fetchBinanceData fetches the latest bid/ask data from the Binance API,
// parses the JSON response, and updates our global tickerInfo in a thread-safe way.
func fetchBinanceData(symbol, apiKey, apiSecret string) error {
	// Construct the URL for the public bookTicker endpoint, e.g.:
	// https://api.binance.com/api/v3/ticker/bookTicker?symbol=BTCUSDT
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/bookTicker?symbol=%s", symbol)

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// If we needed authentication for private endpoints, we would do it here, e.g.:
	// req.Header.Set("X-MBX-APIKEY", apiKey)
	// And sign the request with secret, etc.
	// But for this public endpoint, it's unnecessary.

	// Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response returned HTTP 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response from Binance: %d", resp.StatusCode)
	}

	// Decode JSON into our BinanceBookTicker struct
	var bTicker BinanceBookTicker
	if err := json.NewDecoder(resp.Body).Decode(&bTicker); err != nil {
		return err
	}

	// Convert the string-based BidPrice and AskPrice into float64
	bid, err := strconv.ParseFloat(bTicker.BidPrice, 64)
	if err != nil {
		return fmt.Errorf("failed to parse bid price: %v", err)
	}
	ask, err := strconv.ParseFloat(bTicker.AskPrice, 64)
	if err != nil {
		return fmt.Errorf("failed to parse ask price: %v", err)
	}

	// Lock the mutex before writing to tickerInfo to ensure thread safety.
	tickerMutex.Lock()
	tickerInfo.Symbol = bTicker.Symbol
	tickerInfo.Bid = bid
	tickerInfo.Ask = ask
	tickerInfo.LastUpdate = time.Now()
	tickerMutex.Unlock()

	return nil
}

// handleLatestPrice is the HTTP handler for the /latest-price endpoint.
// It returns the most recent bid/ask data in JSON format.
func handleLatestPrice(w http.ResponseWriter, r *http.Request) {
	// Acquire a read lock (RLock) since we're only reading tickerInfo
	tickerMutex.RLock()
	defer tickerMutex.RUnlock()

	// Convert tickerInfo to JSON
	data, err := json.Marshal(tickerInfo)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	// Send back a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
