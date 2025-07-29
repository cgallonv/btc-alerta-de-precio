package interfaces

import (
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
	"time"
)

// PriceClient defines the interface for external price data sources.
//
// Example usage:
//
//	var client PriceClient = NewBitcoinClient(...)
//	price, err := client.GetCurrentPrice()
type PriceClient interface {
	GetCurrentPrice() (*bitcoin.PriceData, error)
	GetPriceHistory(days int) ([]bitcoin.PriceData, error)
}

// NotificationSender defines the interface for sending notifications.
//
// Example usage:
//
//	var sender NotificationSender = NewNotificationService(...)
//	err := sender.SendAlert(data)
type NotificationSender interface {
	SendAlert(data *notifications.NotificationData) error
	TestTelegramNotification() error
}

// AlertEvaluator defines the interface for alert condition evaluation.
//
// Example usage:
//
//	var evaluator AlertEvaluator = NewAlertEvaluator()
//	shouldTrigger := evaluator.ShouldTrigger(alert, priceData)
type AlertEvaluator interface {
	ShouldTrigger(alert *storage.Alert, priceData *bitcoin.PriceData) bool
}

// ConfigProvider defines the interface for configuration access.
//
// Example usage:
//
//	var cfg ConfigProvider = NewConfigAdapter(...)
//	interval := cfg.GetCheckInterval()
type ConfigProvider interface {
	GetCheckInterval() time.Duration
	IsEmailNotificationsEnabled() bool
	IsWebPushNotificationsEnabled() bool
	IsTelegramNotificationsEnabled() bool
	GetVAPIDPublicKey() string
}
