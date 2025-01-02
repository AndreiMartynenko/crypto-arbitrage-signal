package main

import (
	"encoding/json"
	"net/http"
)

// handleLatestPrice is the HTTP handler for the /latest-price endpoint.
// It returns the most recent bid/ask data in JSON format.
func handleLatestPrice(w http.ResponseWriter, r *http.Request) {
	// Acquire a read lock (RLock) since we're only reading tickerInfo
	tickerMutex.RLock()
	defer tickerMutex.RUnlock()

	// Convert tickerInfo to JSON
	data, err := json.Marshal(tickerInfo)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	// Send back a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
