package main

// TickerData struct matches the JSON response from the binance-connector
type TickerData struct {
	Symbol     string  `json:"symbol"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	LastUpdate string  `json:"last_update"`
}
