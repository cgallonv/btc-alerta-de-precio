package interfaces

import (
	"time"

	"btc-alerta-de-precio/internal/notifications"
)

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	SendAlert(data *notifications.NotificationData) error
	TestTelegramNotification() error
}

// ConfigProvider defines the interface for configuration operations
type ConfigProvider interface {
	GetCheckInterval() time.Duration
	IsEmailNotificationsEnabled() bool
	IsWebPushNotificationsEnabled() bool
	IsTelegramNotificationsEnabled() bool
	GetVAPIDPublicKey() string
	GetString(key string) string
}
