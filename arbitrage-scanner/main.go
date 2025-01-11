package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// TickerData represents the structure of the JSON response from the binance-connector
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
		// Fetch the latest prices from the connector
		tickers, err := fetchTickerData(connectorURL)
		if err != nil {
			log.Printf("Error fetching data from connector: %v", err)
		} else {
			// Process each ticker and find opportunities
			for _, data := range tickers {
				log.Printf("Fetched data: Symbol=%s, Bid=%.2f, Ask=%.2f, LastUpdate=%s",
					data.Symbol, data.Bid, data.Ask, data.LastUpdate)

				if data.Ask < threshold {
					message := fmt.Sprintf("Opportunity found! Symbol=%s, Ask=%.2f < Threshold=%.2f", data.Symbol, data.Ask, threshold)
					log.Println(message)

					// Notify Telegram
					if err := notifyTelegram(message); err != nil {
						log.Printf("Failed to send Telegram notification: %v", err)
					} else {
						log.Printf("Telegram notification sent: %s", message)
					}
				} else {
					log.Printf("No opportunity: Symbol=%s, Ask=%.2f", data.Symbol, data.Ask)
				}
			}
		}

		// Wait for the next interval
		time.Sleep(checkInterval)
	}
}

// fetchTickerData fetches and parses data from the binance-connector
func fetchTickerData(url string) ([]TickerData, error) {
	var tickerDataMap map[string]TickerData
	var tickerDataList []TickerData

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from connector: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode JSON response into a map
	if err := json.NewDecoder(resp.Body).Decode(&tickerDataMap); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	// Convert map values to a slice
	for _, data := range tickerDataMap {
		tickerDataList = append(tickerDataList, data)
	}

	return tickerDataList, nil
}

// notifyTelegram sends a notification message to the Telegram Notifier service
func notifyTelegram(message string) error {
	notifierURL := "http://telegram-notifier:8004/sendAlert"
	payload := map[string]string{"message": message}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(notifierURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to notify Telegram: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from Telegram Notifier: %s", resp.Status)
	}

	return nil
}
