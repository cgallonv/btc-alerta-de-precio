package main

import (
	"log"
	"time"

	"github.com/cgallonv/btc-alerta-de-precio/config"
	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage/repositories"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	tickerRepo := repositories.NewTickerRepository(db.DB())
	tickerStorage := bitcoin.NewTickerStorage(tickerRepo)
	binanceClient := bitcoin.NewBinanceClient(
		cfg.BinanceAPIKey,
		cfg.BinanceAPISecret,
		cfg.BinanceBaseURL,
		tickerStorage,
	)

	// Calculate time range for past 60 days
	endTime := time.Now()
	startTime := endTime.Add(-60 * 24 * time.Hour)

	// Fetch historical data in chunks to avoid rate limits
	chunkDuration := 24 * time.Hour
	currentStart := startTime

	for currentStart.Before(endTime) {
		currentEnd := currentStart.Add(chunkDuration)
		if currentEnd.After(endTime) {
			currentEnd = endTime
		}

		log.Printf("Fetching data from %s to %s", currentStart.Format(time.RFC3339), currentEnd.Format(time.RFC3339))

		// Fetch historical klines
		tickers, err := binanceClient.GetHistoricalKlines("BTCUSDT", "1m", currentStart, currentEnd)
		if err != nil {
			log.Printf("Error fetching historical data: %v", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// Store each ticker
		for _, ticker := range tickers {
			if err := tickerStorage.StoreTicker24h("BTCUSDT", &ticker); err != nil {
				log.Printf("Error storing ticker: %v", err)
				continue
			}
		}

		log.Printf("Successfully stored %d records", len(tickers))
		currentStart = currentEnd
		time.Sleep(1 * time.Second) // Rate limiting
	}

	log.Println("Historical data backfill completed!")
}
