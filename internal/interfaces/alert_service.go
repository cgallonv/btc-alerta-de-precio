package interfaces

import (
	"time"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/storage"
)

// PriceCacheEntry represents a cached price entry for API responses.
//
// Example usage:
//
//	entry := PriceCacheEntry{Price: 30000, Currency: "USD", Timestamp: time.Now()}
type PriceCacheEntry struct {
	Price              float64   `json:"price"`
	PriceChangePercent float64   `json:"price_change_percent"`
	Currency           string    `json:"currency"`
	Source             string    `json:"source"`
	Timestamp          time.Time `json:"timestamp"`
}

// AlertService defines the interface for alert service operations.
// This is used by the API layer to interact with alert functionality.
//
// Example usage:
//
//	var svc AlertService = NewAlertService(...)
//	err := svc.CreateAlert(alert)
type AlertService interface {
	// Alert CRUD operations
	CreateAlert(alert *storage.Alert) error
	GetAlert(id uint) (*storage.Alert, error)
	GetAlerts() ([]storage.Alert, error)
	UpdateAlert(alert *storage.Alert) error
	DeleteAlert(id uint) error
	ToggleAlert(id uint) error

	// Alert actions
	TestAlert(id uint) error
	ResetAlert(alertID uint) error

	// Price operations
	GetCurrentPrice() (*bitcoin.PriceData, error)
	GetPriceHistory(limit int) ([]PriceCacheEntry, error)
	GetCurrentPercentage() float64

	// System operations
	GetStats() (map[string]interface{}, error)
	IsMonitoring() bool
}
