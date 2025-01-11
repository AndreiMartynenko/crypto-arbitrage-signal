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

// KrakenTickerResponse represents the Kraken API response structure
type KrakenTickerResponse struct {
	Error  []string                          `json:"error"`
	Result map[string]KrakenTickerResultData `json:"result"`
}

// KrakenTickerResultData holds bid/ask data for a symbol
type KrakenTickerResultData struct {
	Ask []string `json:"a"` // Ask price and volume
	Bid []string `json:"b"` // Bid price and volume
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

	// Convert pairs to Kraken format
	symbols := mapToKrakenSymbols(strings.Split(pairs, ","))

	log.Printf("Starting kraken-connector for symbols=%v, pollInterval=%v", symbols, pollInterval)

	// Start fetching data in a background goroutine
	go func() {
		for {
			fetchKrakenData(symbols)
			time.Sleep(pollInterval)
		}
	}()

	// Start HTTP server
	http.HandleFunc("/latest-price", handleLatestPrice)
	port := "8002"
	log.Printf("HTTP server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// mapToKrakenSymbols converts pairs to Kraken-specific symbols
func mapToKrakenSymbols(pairs []string) []string {
	mappings := map[string]string{
		"BTC/USDT": "XXBTZUSD",
		"ETH/USDT": "XETHZUSD",
	}
	var krakenSymbols []string
	for _, pair := range pairs {
		if symbol, exists := mappings[pair]; exists {
			krakenSymbols = append(krakenSymbols, symbol)
		}
	}
	return krakenSymbols
}

// fetchKrakenData fetches the bid/ask data for all symbols from Kraken
func fetchKrakenData(symbols []string) {
	baseURL := "https://api.kraken.com/0/public/Ticker"
	symbolsQuery := strings.Join(symbols, ",")

	url := fmt.Sprintf("%s?pair=%s", baseURL, symbolsQuery)
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

	// Parse Kraken API response
	var kResponse KrakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&kResponse); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	// Check for API errors
	if len(kResponse.Error) > 0 {
		log.Printf("Kraken API error: %v", kResponse.Error)
		return
	}

	// Update ticker data
	for symbol, kData := range kResponse.Result {
		updateTickerData(symbol, kData)
	}
}

// updateTickerData safely updates the global tickerInfo map
func updateTickerData(symbol string, kData KrakenTickerResultData) {
	tickerMutex.Lock()
	defer tickerMutex.Unlock()

	// Parse bid and ask prices
	ask, _ := strconv.ParseFloat(kData.Ask[0], 64)
	bid, _ := strconv.ParseFloat(kData.Bid[0], 64)

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
