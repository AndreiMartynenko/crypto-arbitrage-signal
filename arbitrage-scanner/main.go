package main

import "os"

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

}
