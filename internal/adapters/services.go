package adapters

import (
	"fmt"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
)

// PriceClientAdapter adapts the bitcoin.BinanceClient to the interfaces.PriceClient interface
type PriceClientAdapter struct {
	client *bitcoin.BinanceClient
}

// NewPriceClientAdapter creates a new price client adapter
func NewPriceClientAdapter(configProvider interfaces.ConfigProvider) *PriceClientAdapter {
	apiKey := configProvider.GetString("binance.api_key")
	apiSecret := configProvider.GetString("binance.api_secret")
	client := bitcoin.NewBinanceClient(apiKey, apiSecret)
	return &PriceClientAdapter{client: client}
}

// GetCurrentPrice implements interfaces.PriceClient
func (a *PriceClientAdapter) GetCurrentPrice() (*bitcoin.PriceData, error) {
	return a.client.GetCurrentPrice()
}

// GetPriceHistory implements interfaces.PriceClient
func (a *PriceClientAdapter) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	// Not implemented in Binance client yet
	return nil, fmt.Errorf("price history not implemented")
}

// NotificationServiceAdapter adapts notifications.Service to implement NotificationSender interface.
//
// Example usage:
//
//	adapter := NewNotificationServiceAdapter(service)
//	err := adapter.SendAlert(data)
type NotificationServiceAdapter struct {
	service *notifications.Service
}

// NewNotificationServiceAdapter creates a new NotificationServiceAdapter.
//
// Example usage:
//
//	adapter := NewNotificationServiceAdapter(service)
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

// AlertEvaluatorImpl implements the AlertEvaluator interface.
//
// Example usage:
//
//	evaluator := NewAlertEvaluator()
//	shouldTrigger := evaluator.ShouldTrigger(alert, priceData)
type AlertEvaluatorImpl struct{}

// NewAlertEvaluator creates a new AlertEvaluatorImpl.
//
// Example usage:
//
//	evaluator := NewAlertEvaluator()
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

// ConfigAdapter adapts config.Config to implement ConfigProvider interface.
//
// Example usage:
//
//	adapter := NewConfigAdapter(cfg)
//	interval := adapter.GetCheckInterval()
type ConfigAdapter struct {
	config *config.Config
}

// NewConfigAdapter creates a new ConfigAdapter.
//
// Example usage:
//
//	adapter := NewConfigAdapter(cfg)
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

func (a *ConfigAdapter) GetString(key string) string {
	switch key {
	case "binance.api_key":
		return a.config.BinanceAPIKey
	case "binance.api_secret":
		return a.config.BinanceAPISecret
	default:
		return ""
	}
}
