package main

import "fmt"

// formatSymbol is a helper that just ensures "BTC/USDT" -> "BTCUSDT"
// This is what Binance expects in its 'symbol' query parameter.
func formatSymbol(pair string) string {
	return fmt.Sprintf("%s%s", stripSlash(pair))
}

// stripSlash removes "/" from a string (e.g., "BTC/USDT" -> "BTCUSDT").
func stripSlash(pair string) string {
	return stringReplace(pair, "/", "")
}

// stringReplace is a basic function that replaces all occurrences of `old` with `new`.
// We could also use strings.ReplaceAll(pair, "/", "") for brevity.
func stringReplace(str, old, new string) string {
	result := ""
	for _, ch := range str {
		if string(ch) == old {
			result += new
		} else {
			result += string(ch)
		}
	}
	return result
}
