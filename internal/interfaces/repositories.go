package interfaces

import (
	"btc-alerta-de-precio/internal/storage"
)

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	CreateAlert(alert *storage.Alert) error
	GetAlert(id uint) (*storage.Alert, error)
	GetAlerts() ([]storage.Alert, error)
	GetActiveAlerts() ([]storage.Alert, error)
	UpdateAlert(alert *storage.Alert) error
	DeleteAlert(id uint) error
	ToggleAlert(id uint) error
}

// PriceRepository defines the interface for price history operations
type PriceRepository interface {
	SavePriceHistory(price *storage.PriceHistory) error
	GetLatestPrice() (*storage.PriceHistory, error)
}

// NotificationRepository defines the interface for notification logging
type NotificationRepository interface {
	LogNotification(log *storage.NotificationLog) error
	GetNotificationLogs(alertID uint, limit int) ([]storage.NotificationLog, error)
}

// StatsRepository defines the interface for application statistics
type StatsRepository interface {
	GetStats() (map[string]interface{}, error)
}
