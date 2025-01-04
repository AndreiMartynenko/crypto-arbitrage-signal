package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
) // Provides sync.RWMutex for concurrency-safe data access

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

// fetchBinanceData fetches the latest bid/ask data from the Binance API,
// parses the JSON response, and updates our global tickerInfo in a thread-safe way.
func fetchBinanceData() {

	// Construct the URL for the public bookTicker endpoint, e.g.:
	// https://api.binance.com/api/v3/ticker/bookTicker?symbol=BTCUSDT
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/bookTicker?symbol=%s", symbol)

}
