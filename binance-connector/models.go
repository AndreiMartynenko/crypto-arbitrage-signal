package main

import "time" // For time.Duration, sleeping, and timestamps

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
