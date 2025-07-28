package interfaces

import (
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
	"time"
)

// PriceClient defines the interface for external price data sources
type PriceClient interface {
	GetCurrentPrice() (*bitcoin.PriceData, error)
	GetPriceHistory(days int) ([]bitcoin.PriceData, error)
}

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	SendAlert(data *notifications.NotificationData) error
	TestTelegramNotification() error
}

// AlertEvaluator defines the interface for alert condition evaluation
type AlertEvaluator interface {
	ShouldTrigger(alert *storage.Alert, priceData *bitcoin.PriceData) bool
}

// ConfigProvider defines the interface for configuration access
type ConfigProvider interface {
	GetCheckInterval() time.Duration
	IsEmailNotificationsEnabled() bool
	IsWebPushNotificationsEnabled() bool
	IsTelegramNotificationsEnabled() bool
	GetVAPIDPublicKey() string
}
