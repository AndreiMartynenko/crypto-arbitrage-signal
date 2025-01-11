package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// TelegramMessage represents the message payload to send to Telegram
type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	// Read environment variables
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not set")
	}

	log.Println("Starting Telegram Notifier on port 8004")

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Send alert endpoint
	http.HandleFunc("/sendAlert", func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON body
		var requestData map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		message, ok := requestData["message"]
		if !ok {
			http.Error(w, "Missing 'message' field in request body", http.StatusBadRequest)
			return
		}

		// Send message to Telegram
		if err := sendTelegramMessage(botToken, chatID, message); err != nil {
			log.Printf("Failed to send message: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		log.Printf("Message sent: %s", message)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message sent"))
	})

	log.Fatal(http.ListenAndServe(":8004", nil))
}

// sendTelegramMessage sends a message to Telegram
func sendTelegramMessage(botToken, chatID, text string) error {
	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := TelegramMessage{
		ChatID: chatID,
		Text:   text,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(telegramURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to call Telegram API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from Telegram API: %s", resp.Status)
	}

	return nil
}
