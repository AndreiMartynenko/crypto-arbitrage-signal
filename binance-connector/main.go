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

// BinanceBookTicker represents the JSON response from Binance
type BinanceBookTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
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
	binanceAPIKey := os.Getenv("BINANCE_API_KEY")
	binanceAPISecret := os.Getenv("BINANCE_API_SECRET")
	pairs := os.Getenv("PAIRS") // Example: "BTC/USDT,ETH/USDT"
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

	// Split pairs and format them for Binance API
	symbols := formatSymbols(pairs) // Converts "BTC/USDT,ETH/USDT" -> ["BTCUSDT", "ETHUSDT"]

	log.Printf("Starting binance-connector for symbols=%v, pollInterval=%v", symbols, pollInterval)

	// Start fetching data in a background goroutine
	go func() {
		for {
			fetchBinanceData(symbols, binanceAPIKey, binanceAPISecret)
			time.Sleep(pollInterval)
		}
	}()

	// Start HTTP server
	http.HandleFunc("/latest-price", handleLatestPrice)
	port := "8001"
	log.Printf("HTTP server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// formatSymbols converts "BTC/USDT,ETH/USDT" to ["BTCUSDT", "ETHUSDT"]
func formatSymbols(pairs string) []string {
	rawPairs := strings.Split(pairs, ",")
	formatted := make([]string, len(rawPairs))
	for i, pair := range rawPairs {
		formatted[i] = strings.ReplaceAll(pair, "/", "")
	}
	return formatted
}

// fetchBinanceData fetches the bid/ask data for all symbols
func fetchBinanceData(symbols []string, apiKey, apiSecret string) {
	for _, symbol := range symbols {
		url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/bookTicker?symbol=%s", symbol)

		log.Printf("Request URL: %s", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching data for %s: %v", symbol, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Non-200 response for %s: %d", symbol, resp.StatusCode)
			continue
		}

		var bTicker BinanceBookTicker
		if err := json.NewDecoder(resp.Body).Decode(&bTicker); err != nil {
			log.Printf("Error decoding response for %s: %v", symbol, err)
			continue
		}

		updateTickerData(bTicker)
	}
}

// updateTickerData safely updates the global tickerInfo map
func updateTickerData(bTicker BinanceBookTicker) {
	tickerMutex.Lock()
	defer tickerMutex.Unlock()

	// Parse bid and ask prices
	bid, _ := strconv.ParseFloat(bTicker.BidPrice, 64)
	ask, _ := strconv.ParseFloat(bTicker.AskPrice, 64)

	tickerInfo[bTicker.Symbol] = TickerData{
		Symbol:     bTicker.Symbol,
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
