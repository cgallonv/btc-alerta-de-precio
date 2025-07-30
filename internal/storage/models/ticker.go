package models

import (
	"time"
)

// TickerData represents price ticker information from exchanges
type TickerData struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Symbol             string    `json:"symbol" gorm:"not null"`
	Timestamp          time.Time `json:"timestamp" gorm:"index"`
	Source             string    `json:"source" gorm:"not null"`
	Interval           string    `json:"interval" gorm:"not null"`
	LastPrice          float64   `json:"last_price"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	WeightedAvgPrice   float64   `json:"weighted_avg_price"`
	PrevClosePrice     float64   `json:"prev_close_price"`
	LastQty            float64   `json:"last_qty"`
	HighPrice          float64   `json:"high_price"`
	LowPrice           float64   `json:"low_price"`
	OpenPrice          float64   `json:"open_price"`
	Volume             float64   `json:"volume"`
	QuoteVolume        float64   `json:"quote_volume"`
	OpenTime           time.Time `json:"open_time"`
	CloseTime          time.Time `json:"close_time"`
	FirstTradeID       int64     `json:"first_trade_id"`
	LastTradeID        int64     `json:"last_trade_id"`
	TotalTrades        int64     `json:"total_trades"`
	CreatedAt          time.Time `json:"created_at"`
}

// Indexes returns the fields that should be indexed in the database
func (TickerData) Indexes() [][]string {
	return [][]string{
		{"symbol"},
		{"timestamp"},
		{"source"},
		{"symbol", "timestamp"},
	}
}
