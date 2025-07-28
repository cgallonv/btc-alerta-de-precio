package adapters

import (
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
)

// BitcoinClientAdapter adapts bitcoin.Client to implement PriceClient interface
type BitcoinClientAdapter struct {
	client *bitcoin.Client
}

func NewBitcoinClientAdapter(client *bitcoin.Client) interfaces.PriceClient {
	return &BitcoinClientAdapter{client: client}
}

func (a *BitcoinClientAdapter) GetCurrentPrice() (*bitcoin.PriceData, error) {
	price, err := a.client.GetCurrentPrice()
	if err != nil {
		return nil, errors.WrapError(err, "BITCOIN_CLIENT_ERROR", "Failed to get current price")
	}
	return price, nil
}

func (a *BitcoinClientAdapter) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	history, err := a.client.GetPriceHistory(days)
	if err != nil {
		return nil, errors.WrapError(err, "BITCOIN_CLIENT_HISTORY_ERROR", "Failed to get price history")
	}
	return history, nil
}

// NotificationServiceAdapter adapts notifications.Service to implement NotificationSender interface
type NotificationServiceAdapter struct {
	service *notifications.Service
}

func NewNotificationServiceAdapter(service *notifications.Service) interfaces.NotificationSender {
	return &NotificationServiceAdapter{service: service}
}

func (a *NotificationServiceAdapter) SendAlert(data *notifications.NotificationData) error {
	if err := a.service.SendAlert(data); err != nil {
		return errors.WrapError(err, "NOTIFICATION_SEND_ERROR", "Failed to send alert notification")
	}
	return nil
}

func (a *NotificationServiceAdapter) TestTelegramNotification() error {
	if err := a.service.TestTelegramNotification(); err != nil {
		return errors.WrapError(err, "TELEGRAM_TEST_ERROR", "Failed to test Telegram notification")
	}
	return nil
}

// AlertEvaluatorImpl implements the AlertEvaluator interface
type AlertEvaluatorImpl struct{}

func NewAlertEvaluator() interfaces.AlertEvaluator {
	return &AlertEvaluatorImpl{}
}

func (e *AlertEvaluatorImpl) ShouldTrigger(alert *storage.Alert, priceData *bitcoin.PriceData) bool {
	if !alert.IsActive {
		return false
	}

	// One-Shot: Only trigger if never activated before
	if alert.LastTriggered != nil {
		return false
	}

	switch alert.Type {
	case "above":
		return priceData.Price >= alert.TargetPrice
	case "below":
		return priceData.Price <= alert.TargetPrice
	case "change":
		// Use Binance API percentage directly (24h change)
		// Only available when source is Binance, fallback for other sources
		if priceData.Source != "Binance" {
			// For non-Binance sources, we don't have percentage data
			// This maintains compatibility when Binance is unavailable
			return false
		}

		changePercent := priceData.PriceChangePercent

		// Handle different percentage alert types
		if alert.Percentage > 0 {
			// Positive percentage: only trigger for positive changes >= threshold
			return changePercent >= alert.Percentage
		} else if alert.Percentage < 0 {
			// Negative percentage: only trigger for negative changes <= threshold
			return changePercent <= alert.Percentage
		} else {
			// Zero percentage: invalid, never trigger
			return false
		}
	default:
		return false
	}
}

// ConfigAdapter adapts config.Config to implement ConfigProvider interface
type ConfigAdapter struct {
	config *config.Config
}

func NewConfigAdapter(config *config.Config) interfaces.ConfigProvider {
	return &ConfigAdapter{config: config}
}

func (a *ConfigAdapter) GetCheckInterval() time.Duration {
	return a.config.CheckInterval
}

func (a *ConfigAdapter) IsEmailNotificationsEnabled() bool {
	return a.config.EnableEmailNotifications
}

func (a *ConfigAdapter) IsWebPushNotificationsEnabled() bool {
	return a.config.EnableWebPushNotifications
}

func (a *ConfigAdapter) IsTelegramNotificationsEnabled() bool {
	return a.config.EnableTelegramNotifications
}

func (a *ConfigAdapter) GetVAPIDPublicKey() string {
	return a.config.VAPIDPublicKey
}
