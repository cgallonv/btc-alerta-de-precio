package alerts

import (
	"context"
	"fmt"
	"log"
	"time"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
)

// AlertManager coordinates alert operations and integrates with price monitoring
type AlertManager struct {
	alertRepo          interfaces.AlertRepository
	notificationRepo   interfaces.NotificationRepository
	notificationSender interfaces.NotificationSender
	alertEvaluator     interfaces.AlertEvaluator
	priceMonitor       *PriceMonitor
	configProvider     interfaces.ConfigProvider
}

// NewAlertManager creates a new alert manager
func NewAlertManager(
	alertRepo interfaces.AlertRepository,
	notificationRepo interfaces.NotificationRepository,
	notificationSender interfaces.NotificationSender,
	alertEvaluator interfaces.AlertEvaluator,
	priceMonitor *PriceMonitor,
	configProvider interfaces.ConfigProvider,
) *AlertManager {
	am := &AlertManager{
		alertRepo:          alertRepo,
		notificationRepo:   notificationRepo,
		notificationSender: notificationSender,
		alertEvaluator:     alertEvaluator,
		priceMonitor:       priceMonitor,
		configProvider:     configProvider,
	}

	// Register for price updates (simplified callback)
	priceMonitor.AddPriceUpdateCallback(am.onPriceUpdate)

	return am
}

// Start begins alert monitoring
func (am *AlertManager) Start(ctx context.Context) error {
	log.Println("Starting Alert Manager...")

	// Start price monitoring
	if err := am.priceMonitor.Start(ctx); err != nil {
		return errors.WrapError(err, "ALERT_MANAGER_START_ERROR", "Failed to start price monitoring")
	}

	return nil
}

// Stop stops alert monitoring
func (am *AlertManager) Stop() error {
	log.Println("Stopping Alert Manager...")

	if err := am.priceMonitor.Stop(); err != nil {
		return errors.WrapError(err, "ALERT_MANAGER_STOP_ERROR", "Failed to stop price monitoring")
	}

	return nil
}

// IsMonitoring returns true if alert monitoring is active
func (am *AlertManager) IsMonitoring() bool {
	return am.priceMonitor.IsMonitoring()
}

// onPriceUpdate is called when price is updated by the price monitor (simplified)
func (am *AlertManager) onPriceUpdate(current *bitcoin.PriceData) {
	if current == nil {
		return
	}

	am.checkAlerts(current)
}

// checkAlerts evaluates all active alerts against current price data
func (am *AlertManager) checkAlerts(priceData *bitcoin.PriceData) {
	alerts, err := am.alertRepo.GetActiveAlerts()
	if err != nil {
		log.Printf("Error fetching active alerts: %v", err)
		return
	}

	for _, alert := range alerts {
		if am.alertEvaluator.ShouldTrigger(&alert, priceData) {
			if err := am.triggerAlert(&alert, priceData.Price); err != nil {
				log.Printf("Error triggering alert %d: %v", alert.ID, err)
			}
		}
	}
}

// triggerAlert processes an alert trigger
func (am *AlertManager) triggerAlert(alert *storage.Alert, currentPrice float64) error {
	log.Printf("🚨 Triggering alert: %s (Price: $%.2f)", alert.Name, currentPrice)

	// Mark alert as triggered
	alert.MarkTriggered()
	if err := am.alertRepo.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "ALERT_UPDATE_ERROR", "Failed to update triggered alert")
	}

	// Prepare notification data
	notificationData := &notifications.NotificationData{
		Title:   fmt.Sprintf("🚨 Bitcoin Alert: %s", alert.Name),
		Message: am.generateAlertMessage(alert, currentPrice),
		Price:   currentPrice,
		Alert:   alert,
	}

	// Send notification
	if err := am.notificationSender.SendAlert(notificationData); err != nil {
		log.Printf("Error sending notification for alert %s: %v", alert.Name, err)

		// Log notification failure
		am.logNotification(alert.ID, "error", err.Error())

		return errors.WrapError(err, "NOTIFICATION_SEND_ERROR", "Failed to send alert notification")
	}

	log.Printf("✅ Notification sent successfully for alert: %s", alert.Name)

	// Log successful notification
	am.logNotification(alert.ID, "sent", "Notification sent successfully")

	return nil
}

// generateAlertMessage creates the notification message based on alert type
func (am *AlertManager) generateAlertMessage(alert *storage.Alert, currentPrice float64) string {
	var message string
	switch alert.Type {
	case "above":
		message = fmt.Sprintf("Bitcoin price has exceeded $%.2f. Current price: $%.2f", alert.TargetPrice, currentPrice)
	case "below":
		message = fmt.Sprintf("Bitcoin price has fallen below $%.2f. Current price: $%.2f", alert.TargetPrice, currentPrice)
	case "change":
		// Use the 24h percentage change from Binance for the message
		lastPrice := am.priceMonitor.GetLastPrice()
		if lastPrice != nil && lastPrice.Source == "Binance" {
			changePercent := lastPrice.PriceChangePercent
			direction := "increased"
			if changePercent < 0 {
				direction = "decreased"
				changePercent = -changePercent
			}
			message = fmt.Sprintf("Bitcoin price has %s %.2f%% (24h). Current price: $%.2f", direction, changePercent, currentPrice)
		} else {
			message = fmt.Sprintf("Significant Bitcoin price change detected. Current price: $%.2f", currentPrice)
		}
	default:
		message = fmt.Sprintf("Bitcoin alert triggered. Current price: $%.2f", currentPrice)
	}

	// Agregar el nombre de la alerta al principio del mensaje
	return fmt.Sprintf("🔔 %s\n%s", alert.Name, message)
}

// logNotification logs notification attempts
func (am *AlertManager) logNotification(alertID uint, status, message string) {
	notificationLog := &storage.NotificationLog{
		AlertID: alertID,
		Type:    "combined", // email + desktop + web push
		Status:  status,
		Message: message,
		SentAt:  time.Now(),
	}

	if err := am.notificationRepo.LogNotification(notificationLog); err != nil {
		log.Printf("Error logging notification: %v", err)
	}
}

// CRUD operations for alerts
func (am *AlertManager) CreateAlert(alert *storage.Alert) error {
	if err := am.alertRepo.CreateAlert(alert); err != nil {
		return errors.WrapError(err, "CREATE_ALERT_ERROR", "Failed to create alert")
	}
	return nil
}

func (am *AlertManager) GetAlert(id uint) (*storage.Alert, error) {
	alert, err := am.alertRepo.GetAlert(id)
	if err != nil {
		return nil, errors.WrapError(err, "GET_ALERT_ERROR", "Failed to get alert").WithField("alert_id", id)
	}
	return alert, nil
}

func (am *AlertManager) GetAlerts() ([]storage.Alert, error) {
	alerts, err := am.alertRepo.GetAlerts()
	if err != nil {
		return nil, errors.WrapError(err, "GET_ALERTS_ERROR", "Failed to get alerts")
	}
	return alerts, nil
}

func (am *AlertManager) UpdateAlert(alert *storage.Alert) error {
	if err := am.alertRepo.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "UPDATE_ALERT_ERROR", "Failed to update alert").WithField("alert_id", alert.ID)
	}
	return nil
}

func (am *AlertManager) DeleteAlert(id uint) error {
	if err := am.alertRepo.DeleteAlert(id); err != nil {
		return errors.WrapError(err, "DELETE_ALERT_ERROR", "Failed to delete alert").WithField("alert_id", id)
	}
	return nil
}

func (am *AlertManager) ToggleAlert(id uint) error {
	if err := am.alertRepo.ToggleAlert(id); err != nil {
		return errors.WrapError(err, "TOGGLE_ALERT_ERROR", "Failed to toggle alert").WithField("alert_id", id)
	}
	return nil
}

func (am *AlertManager) ResetAlert(alertID uint) error {
	alert, err := am.alertRepo.GetAlert(alertID)
	if err != nil {
		return errors.WrapError(err, "RESET_ALERT_ERROR", "Alert not found").WithField("alert_id", alertID)
	}

	alert.Reset()

	if err := am.alertRepo.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "RESET_ALERT_UPDATE_ERROR", "Failed to reset alert").WithField("alert_id", alertID)
	}

	log.Printf("🔄 Alert reset: %s (ID: %d)", alert.Name, alertID)
	return nil
}

func (am *AlertManager) TestAlert(id uint) error {
	alert, err := am.alertRepo.GetAlert(id)
	if err != nil {
		return errors.WrapError(err, "TEST_ALERT_ERROR", "Alert not found").WithField("alert_id", id)
	}

	lastPrice := am.priceMonitor.GetLastPrice()
	if lastPrice == nil {
		return errors.NewAppError("TEST_ALERT_NO_PRICE", "No current price data available")
	}

	// Create test alert (copy original but mark as test)
	testAlert := *alert
	testAlert.Name = "🧪 TEST: " + alert.Name

	notificationData := &notifications.NotificationData{
		Title:   fmt.Sprintf("🧪 Test Alert: %s", alert.Name),
		Message: fmt.Sprintf("This is a test of alert '%s'. Current price: $%.2f", alert.Name, lastPrice.Price),
		Price:   lastPrice.Price,
		Alert:   &testAlert,
	}

	if err := am.notificationSender.SendAlert(notificationData); err != nil {
		return errors.WrapError(err, "TEST_ALERT_SEND_ERROR", "Failed to send test notification")
	}

	return nil
}

// GetCurrentPrice returns current price from price monitor
func (am *AlertManager) GetCurrentPrice() (*bitcoin.PriceData, error) {
	price := am.priceMonitor.GetLastPrice()
	if price == nil {
		return nil, errors.NewAppError("NO_CURRENT_PRICE", "No current price data available")
	}
	return price, nil
}

// GetCurrentPercentage returns current percentage change
func (am *AlertManager) GetCurrentPercentage() float64 {
	return am.priceMonitor.GetCurrentPercentage()
}

// GetPriceHistory returns cached price history
func (am *AlertManager) GetPriceHistory(limit int) []interfaces.PriceCacheEntry {
	return am.priceMonitor.GetPriceHistory(limit)
}
