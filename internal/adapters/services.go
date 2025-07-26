package adapters

import (
	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
	"time"
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
		return nil, errors.WrapError(err, "PRICE_CLIENT_ERROR", "Failed to get current price")
	}
	return price, nil
}

func (a *BitcoinClientAdapter) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	history, err := a.client.GetPriceHistory(days)
	if err != nil {
		return nil, errors.WrapError(err, "PRICE_CLIENT_HISTORY_ERROR", "Failed to get price history").WithField("days", days)
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

func (a *NotificationServiceAdapter) SendWebPushNotification(subscriptions []storage.WebPushSubscription, data *notifications.NotificationData) error {
	if err := a.service.SendWebPushNotification(subscriptions, data); err != nil {
		return errors.WrapError(err, "WEBPUSH_SEND_ERROR", "Failed to send web push notification")
	}
	return nil
}

func (a *NotificationServiceAdapter) TestTelegramNotification() error {
	if err := a.service.TestTelegramNotification(); err != nil {
		return errors.WrapError(err, "TELEGRAM_TEST_ERROR", "Failed to send test telegram notification")
	}
	return nil
}

// ConfigAdapter adapts config.Config to implement ConfigProvider interface
type ConfigAdapter struct {
	config *config.Config
}

func NewConfigAdapter(cfg *config.Config) interfaces.ConfigProvider {
	return &ConfigAdapter{config: cfg}
}

func (a *ConfigAdapter) GetCheckInterval() time.Duration {
	return a.config.CheckInterval
}

func (a *ConfigAdapter) GetPercentageUpdateInterval() time.Duration {
	return a.config.PercentageUpdateInterval
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

// AlertEvaluatorImpl implements the AlertEvaluator interface
type AlertEvaluatorImpl struct{}

func NewAlertEvaluator() interfaces.AlertEvaluator {
	return &AlertEvaluatorImpl{}
}

// ShouldTrigger evaluates if an alert should be triggered based on current conditions
//
// For percentage change alerts, the behavior depends on the sign of the percentage:
//
// **Positive Percentage (e.g., 3.0):**
//   - Triggers ONLY when price increases by 3% or more
//   - Example: Alert with 3.0% triggers when price goes from $50,000 to $51,500+ (+3% or more)
//   - Will NOT trigger on price decreases, even large ones
//
// **Negative Percentage (e.g., -3.0):**
//   - Triggers ONLY when price decreases by 3% or more (becomes -3% or lower)
//   - Example: Alert with -3.0% triggers when price goes from $50,000 to $48,500- (-3% or more)
//   - Will NOT trigger on price increases, even large ones
//
// **Zero Percentage (0.0):**
//   - Never triggers (considered invalid)
//
// **Examples:**
//   - Percentage: 5.0  → Triggers on: +5%, +10%, +15%... | Does NOT trigger on: -5%, -10%, +2%
//   - Percentage: -5.0 → Triggers on: -5%, -10%, -15%... | Does NOT trigger on: +5%, +10%, -2%
//
// This allows for precise directional alerts:
//   - Bull market alerts: Use positive percentages to catch upward momentum
//   - Bear market alerts: Use negative percentages to catch downward trends
//   - Risk management: Set negative percentage alerts for stop-loss notifications
func (e *AlertEvaluatorImpl) ShouldTrigger(alert *storage.Alert, currentPrice, previousPrice float64) bool {
	if !alert.IsActive {
		return false
	}

	// One-Shot: Only trigger if never activated before
	if alert.LastTriggered != nil {
		return false
	}

	switch alert.Type {
	case "above":
		return currentPrice >= alert.TargetPrice
	case "below":
		return currentPrice <= alert.TargetPrice
	case "change":
		if previousPrice == 0 {
			return false
		}

		// Calculate actual percentage change
		changePercent := ((currentPrice - previousPrice) / previousPrice) * 100

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
