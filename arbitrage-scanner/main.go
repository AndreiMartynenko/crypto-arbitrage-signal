package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

// TickerData struct matches the JSON response from the binance-connector
type TickerData struct {
	Symbol     string  `json:"symbol"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	LastUpdate string  `json:"last_update"`
}

func main() {
	// Read environment variables
	connectorURL := os.Getenv("CONNECTOR_URL")
	if connectorURL == "" {
		connectorURL = "http://binance-connector:8001/latest-price" // Default value
	}

	thresholdStr := os.Getenv("PRICE_THRESHOLD")
	if thresholdStr == "" {
		thresholdStr = "20000" // Default threshold
	}
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		log.Fatalf("Invalid PRICE_THRESHOLD: %v", err)
	}

	checkIntervalStr := os.Getenv("CHECK_INTERVAL")
	if checkIntervalStr == "" {
		checkIntervalStr = "5s" // Default check interval
	}
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		log.Fatalf("Invalid CHECK_INTERVAL: %v", err)
	}

	log.Printf("Starting arbitrage scanner: connectorURL=%s, threshold=%.2f, interval=%v",
		connectorURL, threshold, checkInterval)

	// Main loop
	for {
		// Fetch the latest price from the connector
		data, err := fetchTickerData(connectorURL)
		if err != nil {
			log.Printf("Error fetching data from connector: %v", err)
		} else {
			// Compare the ask price to the threshold
			if data.Ask < threshold {
				log.Printf("Opportunity found! Symbol=%s, Ask=%.2f < Threshold=%.2f", data.Symbol, data.Ask, threshold)
				// Here, you could trigger a Telegram alert or other action
			} else {
				log.Printf("No opportunity: Symbol=%s, Ask=%.2f", data.Symbol, data.Ask)
			}
		}

		// Wait for the next interval
		time.Sleep(checkInterval)
	}

}
