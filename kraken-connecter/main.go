package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// KrakenBookTicker represents the JSON response from Binance
type KrakenBookTicker struct {
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
	krakenAPIKey := os.Getenv("KRAKEN_API_KEY")
	krakenAPISecret := os.Getenv("KRAKEN_API_SECRET")
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

	// Split pairs and format them for Kraken API
	symbols := formatSymbols(pairs) // Converts "BTC/USDT,ETH/USDT" -> ["BTCUSDT", "ETHUSDT"]

	log.Printf("Starting binance-connector for symbols=%v, pollInterval=%v", symbols, pollInterval)

	// Start fetching data in a background goroutine
	go func() {
		for {
			fetchBinanceData(symbols, krakenAPIKey, krakenAPISecret)
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
