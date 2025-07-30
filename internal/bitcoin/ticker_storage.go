// Package bitcoin provides functionality for interacting with cryptocurrency APIs.
package bitcoin

import (
	"log"
	"strconv"
	"time"

	"btc-alerta-de-precio/internal/storage/models"
	"btc-alerta-de-precio/internal/storage/repositories"
)

// TickerStorage handles the persistence of ticker data from Binance API.
// It converts API responses to database models and manages their storage.
//
// Example usage:
//
//	storage := NewTickerStorage(tickerRepo)
//	if err := storage.StoreTicker24h(symbol, response); err != nil {
//	    log.Printf("Error storing ticker: %v", err)
//	}
type TickerStorage struct {
	repo *repositories.TickerRepository
}

// NewTickerStorage creates a new TickerStorage instance.
func NewTickerStorage(repo *repositories.TickerRepository) *TickerStorage {
	return &TickerStorage{repo: repo}
}

// StoreTicker24h stores the response from /api/v3/ticker/24hr endpoint.
func (s *TickerStorage) StoreTicker24h(symbol string, response *Ticker24hResponse) error {
	// Parse numeric values
	lastPrice, _ := strconv.ParseFloat(response.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(response.PriceChange, 64)
	priceChangePercent, _ := strconv.ParseFloat(response.PriceChangePercent, 64)
	weightedAvgPrice, _ := strconv.ParseFloat(response.WeightedAvgPrice, 64)
	prevClosePrice, _ := strconv.ParseFloat(response.PrevClosePrice, 64)
	lastQty, _ := strconv.ParseFloat(response.LastQty, 64)
	highPrice, _ := strconv.ParseFloat(response.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(response.LowPrice, 64)
	openPrice, _ := strconv.ParseFloat(response.OpenPrice, 64)
	volume, _ := strconv.ParseFloat(response.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(response.QuoteVolume, 64)

	// Create ticker data model
	ticker := &models.TickerData{
		Symbol:             symbol,
		Timestamp:          time.Now(),
		Source:             "Binance",
		Interval:           "1m",
		LastPrice:          lastPrice,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		WeightedAvgPrice:   weightedAvgPrice,
		PrevClosePrice:     prevClosePrice,
		LastQty:            lastQty,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		OpenPrice:          openPrice,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		OpenTime:           time.Unix(response.OpenTime/1000, 0),
		CloseTime:          time.Unix(response.CloseTime/1000, 0),
		FirstTradeID:       response.FirstID,
		LastTradeID:        response.LastID,
		TotalTrades:        response.Count,
	}

	// Store in database
	if err := s.repo.Store(ticker); err != nil {
		log.Printf("❌ Error storing ticker data: %v", err)
		return err
	}

	log.Printf("✅ Stored ticker data for %s: $%.2f (%+.2f%%)",
		symbol, lastPrice, priceChangePercent)
	return nil
}

// Ticker24hResponse represents the response from Binance /api/v3/ticker/24hr endpoint.
type Ticker24hResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}
