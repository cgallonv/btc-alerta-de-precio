package interfaces

import (
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/storage"
)

// AlertService defines the interface for alert service operations
// This is used by the API layer to interact with alert functionality
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
	GetPriceHistory(limit int) ([]storage.PriceHistory, error)
	GetCurrentPercentage() float64

	// System operations
	GetStats() (map[string]interface{}, error)
	IsMonitoring() bool
}
