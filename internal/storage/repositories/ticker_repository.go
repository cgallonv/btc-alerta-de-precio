package repositories

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/cgallonv/btc-alerta-de-precio/internal/storage/models"
)

// TickerRepository handles storage operations for ticker data
type TickerRepository struct {
	db *gorm.DB
}

// NewTickerRepository creates a new TickerRepository instance
func NewTickerRepository(db *gorm.DB) *TickerRepository {
	return &TickerRepository{db: db}
}

// Store saves a ticker data record to the database
func (r *TickerRepository) Store(ticker *models.TickerData) error {
	if ticker.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if ticker.Source == "" {
		return fmt.Errorf("source is required")
	}

	if ticker.Interval == "" {
		ticker.Interval = "1m" // Default interval
	}

	return r.db.Create(ticker).Error
}

// GetLatest returns the most recent ticker data for a given symbol
func (r *TickerRepository) GetLatest(symbol string) (*models.TickerData, error) {
	var ticker models.TickerData
	err := r.db.Where("symbol = ?", symbol).Order("timestamp desc").First(&ticker).Error
	if err != nil {
		return nil, err
	}
	return &ticker, nil
}

// GetHistory returns historical ticker data for a given symbol and time range
func (r *TickerRepository) GetHistory(symbol string, start, end time.Time, limit int) ([]models.TickerData, error) {
	var tickers []models.TickerData
	query := r.db.Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, start, end)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("timestamp desc").Find(&tickers).Error
	if err != nil {
		return nil, err
	}
	return tickers, nil
}

// GetPriceRange returns the highest and lowest prices for a symbol within a time range
func (r *TickerRepository) GetPriceRange(symbol string, start, end time.Time) (high, low float64, err error) {
	var result struct {
		HighPrice float64
		LowPrice  float64
	}

	err = r.db.Model(&models.TickerData{}).
		Select("MAX(high_price) as high_price, MIN(low_price) as low_price").
		Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, start, end).
		Scan(&result).Error

	return result.HighPrice, result.LowPrice, err
}

// GetAveragePrice calculates the volume-weighted average price for a symbol within a time range
func (r *TickerRepository) GetAveragePrice(symbol string, start, end time.Time) (float64, error) {
	var result struct {
		VWAP float64
	}

	err := r.db.Model(&models.TickerData{}).
		Select("SUM(last_price * volume) / SUM(volume) as vwap").
		Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, start, end).
		Scan(&result).Error

	return result.VWAP, err
}

// GetVolumeStats returns volume statistics for a symbol within a time range
func (r *TickerRepository) GetVolumeStats(symbol string, start, end time.Time) (totalVolume, quoteVolume float64, numTrades int64, err error) {
	var result struct {
		TotalVolume float64
		QuoteVolume float64
		TotalTrades int64
	}

	err = r.db.Model(&models.TickerData{}).
		Select("SUM(volume) as total_volume, SUM(quote_volume) as quote_volume, SUM(total_trades) as total_trades").
		Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, start, end).
		Scan(&result).Error

	return result.TotalVolume, result.QuoteVolume, result.TotalTrades, err
}

// Cleanup removes ticker data older than the specified duration
func (r *TickerRepository) Cleanup(age time.Duration) error {
	cutoff := time.Now().Add(-age)
	return r.db.Where("timestamp < ?", cutoff).Delete(&models.TickerData{}).Error
}
