package main

//Create a database connection
//Postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database configuration
var db *sql.DB

func initDB() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}

	log.Println("Connected to the database successfully!")
}

func saveSignal(symbol string, bid, ask float64, message string) {
	query := `INSERT INTO signals (symbol, bid_price, ask_price, message) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, symbol, bid, ask, message)
	if err != nil {
		log.Printf("Failed to save signal to database: %v", err)
	}
}
