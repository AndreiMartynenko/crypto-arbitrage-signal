package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// KrakenTickerResponse represents the JSON response from Kraken's Ticker API
type KrakenTickerResponse struct {
	Result map[string]KrakenTickerInfo `json:"result"`
	Error  []string                    `json:"error"`
}

// KrakenTickerInfo contains bid/ask data for a specific trading pair
type KrakenTickerInfo struct {
	Bid []string `json:"b"` // Bid prices
	Ask []string `json:"a"` // Ask prices
}

// TickerData holds the internal representation of bid/ask data
type TickerData struct {
	Symbol     string    `json:"symbol"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	LastUpdate time.Time `json:"last_update"`
}

// Mutex to protect access to tickerInfo
var (
	tickerMutex sync.RWMutex
	tickerInfo  = map[string]TickerData{}
)

func main() {
	// Load environment variables
	apiKey := os.Getenv("KRAKEN_API_KEY")
	apiSecret := os.Getenv("KRAKEN_API_SECRET")
	pairs := os.Getenv("PAIRS") // Example: "BTC/USD,ETH/USD"
	if pairs == "" {
		log.Fatalf("No PAIRS specified in environment variables")
	}
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = "5s"
	}
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		log.Printf("Invalid POLL_INTERVAL format, defaulting to 5s: %v", err)
		pollInterval = 5 * time.Second
	}

	// Split pairs and format them for Kraken API
	symbols := formatSymbols(pairs) // Converts "BTC/USD,ETH/USD" -> ["BTCUSD", "ETHUSD"]

	log.Printf("Starting kraken-connector for symbols=%v, pollInterval=%v", symbols, pollInterval)

	// Start fetching data in a background goroutine
	go func() {
		for {
			fetchKrakenData(symbols, apiKey, apiSecret)
			time.Sleep(pollInterval)
		}
	}()

	// Start HTTP server
	http.HandleFunc("/latest-price", handleLatestPrice)
	port := "8002"
	log.Printf("HTTP server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// formatSymbols converts "BTC/USD,ETH/USD" to ["BTCUSD", "ETHUSD"]
func formatSymbols(pairs string) []string {
	rawPairs := strings.Split(pairs, ",")
	formatted := make([]string, len(rawPairs))
	for i, pair := range rawPairs {
		formatted[i] = strings.ReplaceAll(pair, "/", "")
	}
	return formatted
}

// fetchKrakenData fetches the bid/ask data for all symbols from Kraken API
func fetchKrakenData(symbols []string, apiKey, apiSecret string) {
	// Kraken API URL for Ticker data
	url := fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", strings.Join(symbols, ","))

	log.Printf("Request URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-200 response: %d", resp.StatusCode)
		return
	}

	var response KrakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	if len(response.Error) > 0 {
		log.Printf("Kraken API returned errors: %v", response.Error)
		return
	}

	// Update ticker data
	for pair, ticker := range response.Result {
		updateTickerData(pair, ticker)
	}
}

// updateTickerData safely updates the global tickerInfo map
func updateTickerData(symbol string, ticker KrakenTickerInfo) {
	tickerMutex.Lock()
	defer tickerMutex.Unlock()

	// Parse bid and ask prices
	bid, _ := strconv.ParseFloat(ticker.Bid[0], 64)
	ask, _ := strconv.ParseFloat(ticker.Ask[0], 64)

	tickerInfo[symbol] = TickerData{
		Symbol:     symbol,
		Bid:        bid,
		Ask:        ask,
		LastUpdate: time.Now(),
	}
}

// handleLatestPrice serves the latest bid/ask data as JSON
func handleLatestPrice(w http.ResponseWriter, r *http.Request) {
	tickerMutex.RLock()
	defer tickerMutex.RUnlock()

	data, err := json.Marshal(tickerInfo)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("Sending response: %s", string(data))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
