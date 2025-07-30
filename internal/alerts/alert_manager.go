// Package alerts provides functionality for monitoring Bitcoin prices
// and managing price-based alerts.
package alerts

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
)

// AlertManager coordinates alert operations and integrates with price monitoring.
// It manages price alerts, evaluates alert conditions, and sends notifications
// when alerts are triggered.
//
// Example usage:
//
//	manager, err := NewAlertManager(configProvider, notificationSender, alertEvaluator, alertRepo, notificationRepo)
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	if err := manager.Start(context.Background()); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	defer manager.Stop()
type AlertManager struct {
	// Dependencies
	binanceClient      *bitcoin.BinanceClient
	configProvider     interfaces.ConfigProvider
	notificationSender interfaces.NotificationSender
	alertEvaluator     interfaces.AlertEvaluator
	alertRepo          interfaces.AlertRepository
	notificationRepo   interfaces.NotificationRepository

	// Price monitoring
	priceMonitor *PriceMonitor

	// Alert processing
	processingMux sync.RWMutex
	isProcessing  bool
}

// NewAlertManager creates a new alert manager with the provided dependencies.
// It initializes the Binance client and price monitor, and sets up price update callbacks.
//
// Example usage:
//
//	manager, err := NewAlertManager(
//	    configProvider,
//	    notificationSender,
//	    alertEvaluator,
//	    alertRepo,
//	    notificationRepo,
//	)
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func NewAlertManager(
	configProvider interfaces.ConfigProvider,
	notificationSender interfaces.NotificationSender,
	alertEvaluator interfaces.AlertEvaluator,
	alertRepo interfaces.AlertRepository,
	notificationRepo interfaces.NotificationRepository,
	tickerStorage *bitcoin.TickerStorage,
) (*AlertManager, error) {
	// Create Binance client with API credentials
	apiKey := configProvider.GetString("binance.api_key")
	apiSecret := configProvider.GetString("binance.api_secret")
	binanceClient := bitcoin.NewBinanceClient(apiKey, apiSecret, tickerStorage)

	// Create price monitor
	priceMonitor := NewPriceMonitor(configProvider, 20, tickerStorage)

	manager := &AlertManager{
		binanceClient:      binanceClient,
		configProvider:     configProvider,
		notificationSender: notificationSender,
		alertEvaluator:     alertEvaluator,
		alertRepo:          alertRepo,
		notificationRepo:   notificationRepo,
		priceMonitor:       priceMonitor,
	}

	// Register for price updates
	priceMonitor.AddPriceUpdateCallback(manager.checkAlerts)

	return manager, nil
}

// Start begins alert monitoring.
// It starts the price monitor and begins evaluating alert conditions.
//
// Example usage:
//
//	if err := manager.Start(context.Background()); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) Start(ctx context.Context) error {
	log.Printf("Starting Alert Manager...")
	return am.priceMonitor.Start(ctx)
}

// Stop stops alert monitoring.
// It's safe to call Stop multiple times.
// After stopping, the manager can be restarted with Start.
//
// Example usage:
//
//	defer manager.Stop()
func (am *AlertManager) Stop() error {
	return am.priceMonitor.Stop()
}

// IsMonitoring returns true if alert monitoring is active.
//
// Example usage:
//
//	if manager.IsMonitoring() {
//	    log.Printf("Alert monitoring is active")
//	}
func (am *AlertManager) IsMonitoring() bool {
	return am.priceMonitor.IsMonitoring()
}

// checkAlerts evaluates all active alerts against the current price.
func (am *AlertManager) checkAlerts(priceData *bitcoin.PriceData) {
	am.processingMux.Lock()
	defer am.processingMux.Unlock()

	if am.isProcessing {
		return
	}
	am.isProcessing = true
	defer func() { am.isProcessing = false }()

	alerts, err := am.alertRepo.GetActiveAlerts()
	if err != nil {
		log.Printf("Error getting active alerts: %v", err)
		return
	}

	for _, alert := range alerts {
		if am.alertEvaluator.ShouldTrigger(&alert, priceData) {
			if err := am.triggerAlert(&alert, priceData); err != nil {
				log.Printf("Error triggering alert %d: %v", alert.ID, err)
				continue
			}

			// Mark alert as triggered
			alert.MarkTriggered()
			if err := am.alertRepo.UpdateAlert(&alert); err != nil {
				log.Printf("Error updating alert %d: %v", alert.ID, err)
			}
		}
	}
}

// triggerAlert sends notifications for a triggered alert.
func (am *AlertManager) triggerAlert(alert *storage.Alert, priceData *bitcoin.PriceData) error {
	// Prepare notification data
	notificationData := &notifications.NotificationData{
		Title:         "🚨 Bitcoin Alert",
		Message:       alert.GetDescription(),
		Price:         priceData.Price,
		Alert:         alert,
		AlertID:       alert.ID,
		AlertName:     alert.Name,
		AlertType:     alert.Type,
		Percentage:    priceData.PriceChangePercent,
		Email:         alert.Email,
		EnableEmail:   alert.EnableEmail,
	}

	// Send notification
	if err := am.notificationSender.SendAlert(notificationData); err != nil {
		// Log notification failure
		notificationLog := &storage.NotificationLog{
			AlertID:   alert.ID,
			Type:      "error",
			Status:    "failed",
			Message:   "Failed to send notification",
			Error:     err.Error(),
			Price:     priceData.Price,
			SentAt:    time.Now(),
			Timestamp: time.Now(),
		}
		if err := am.notificationRepo.LogNotification(notificationLog); err != nil {
			log.Printf("Error logging notification failure: %v", err)
		}

		return errors.WrapError(err, "NOTIFICATION_SEND_ERROR", "Failed to send alert notification")
	}

	// Log successful notification
	notificationLog := &storage.NotificationLog{
		AlertID:   alert.ID,
		Type:      "success",
		Status:    "sent",
		Message:   "Notification sent successfully",
		Price:     priceData.Price,
		SentAt:    time.Now(),
		Timestamp: time.Now(),
	}
	if err := am.notificationRepo.LogNotification(notificationLog); err != nil {
		log.Printf("Error logging successful notification: %v", err)
	}

	return nil
}

// CRUD operations for alerts

// CreateAlert creates a new alert.
//
// Example usage:
//
//	alert := &storage.Alert{
//	    Name: "Price above $50,000",
//	    Type: "above",
//	    TargetPrice: 50000,
//	    IsActive: true,
//	}
//	if err := manager.CreateAlert(alert); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) CreateAlert(alert *storage.Alert) error {
	if err := am.alertRepo.CreateAlert(alert); err != nil {
		return errors.WrapError(err, "CREATE_ALERT_ERROR", "Failed to create alert")
	}
	return nil
}

// GetAlert retrieves an alert by ID.
//
// Example usage:
//
//	alert, err := manager.GetAlert(123)
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	log.Printf("Alert: %s", alert.Name)
func (am *AlertManager) GetAlert(id uint) (*storage.Alert, error) {
	alert, err := am.alertRepo.GetAlert(id)
	if err != nil {
		return nil, errors.WrapError(err, "GET_ALERT_ERROR", "Failed to get alert").WithField("alert_id", id)
	}
	return alert, nil
}

// GetAlerts retrieves all alerts.
//
// Example usage:
//
//	alerts, err := manager.GetAlerts()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	for _, alert := range alerts {
//	    log.Printf("Alert: %s", alert.Name)
//	}
func (am *AlertManager) GetAlerts() ([]storage.Alert, error) {
	alerts, err := am.alertRepo.GetAlerts()
	if err != nil {
		return nil, errors.WrapError(err, "GET_ALERTS_ERROR", "Failed to get alerts")
	}
	return alerts, nil
}

// UpdateAlert updates an existing alert.
//
// Example usage:
//
//	alert.TargetPrice = 55000
//	if err := manager.UpdateAlert(alert); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) UpdateAlert(alert *storage.Alert) error {
	if err := am.alertRepo.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "UPDATE_ALERT_ERROR", "Failed to update alert")
	}
	return nil
}

// DeleteAlert deletes an alert by ID.
//
// Example usage:
//
//	if err := manager.DeleteAlert(123); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) DeleteAlert(id uint) error {
	if err := am.alertRepo.DeleteAlert(id); err != nil {
		return errors.WrapError(err, "DELETE_ALERT_ERROR", "Failed to delete alert")
	}
	return nil
}

// ToggleAlert toggles an alert's active status.
//
// Example usage:
//
//	if err := manager.ToggleAlert(123); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) ToggleAlert(id uint) error {
	if err := am.alertRepo.ToggleAlert(id); err != nil {
		return errors.WrapError(err, "TOGGLE_ALERT_ERROR", "Failed to toggle alert")
	}
	return nil
}

// ResetAlert resets an alert's trigger status.
//
// Example usage:
//
//	if err := manager.ResetAlert(123); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) ResetAlert(alertID uint) error {
	alert, err := am.GetAlert(alertID)
	if err != nil {
		return errors.WrapError(err, "RESET_ALERT_ERROR", "Failed to get alert for reset")
	}

	alert.Reset()
	if err := am.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "RESET_ALERT_ERROR", "Failed to update alert after reset")
	}

	return nil
}

// TestAlert sends a test notification for an alert.
//
// Example usage:
//
//	if err := manager.TestAlert(123); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (am *AlertManager) TestAlert(id uint) error {
	alert, err := am.GetAlert(id)
	if err != nil {
		return errors.WrapError(err, "TEST_ALERT_ERROR", "Failed to get alert for test")
	}

	// Get current price for test
	priceData, err := am.GetCurrentPrice()
	if err != nil {
		return errors.WrapError(err, "TEST_ALERT_ERROR", "Failed to get current price")
	}

	// Send test notification
	notificationData := &notifications.NotificationData{
		Title:         fmt.Sprintf("🧪 Test Alert: %s", alert.Name),
		Message:       fmt.Sprintf("This is a test of alert '%s'. Current price: $%.2f (%+.2f%%)", alert.Name, priceData.Price, priceData.PriceChangePercent),
		Price:         priceData.Price,
		Alert:         alert,
		AlertID:       alert.ID,
		AlertName:     alert.Name,
		AlertType:     alert.Type,
		Percentage:    priceData.PriceChangePercent,
		Email:         alert.Email,
		EnableEmail:   alert.EnableEmail,
		IsTest:        true,
	}

	if err := am.notificationSender.SendAlert(notificationData); err != nil {
		return errors.WrapError(err, "TEST_ALERT_ERROR", "Failed to send test notification")
	}

	return nil
}

// Price-related methods

// GetCurrentPrice returns the current Bitcoin price.
//
// Example usage:
//
//	price, err := manager.GetCurrentPrice()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	log.Printf("Current price: $%.2f", price.Price)
func (am *AlertManager) GetCurrentPrice() (*bitcoin.PriceData, error) {
	return am.binanceClient.GetCurrentPrice()
}

// GetPriceHistory returns the price history.
//
// Example usage:
//
//	history, err := manager.GetPriceHistory(24)
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	for _, entry := range history {
//	    log.Printf("Price at %s: $%.2f", entry.Timestamp, entry.Price)
//	}
func (am *AlertManager) GetPriceHistory(limit int) ([]interfaces.PriceCacheEntry, error) {
	return am.priceMonitor.GetPriceHistory(limit), nil
}

// GetCurrentPercentage returns the current price change percentage.
//
// Example usage:
//
//	change := manager.GetCurrentPercentage()
//	log.Printf("24h change: %+.2f%%", change)
func (am *AlertManager) GetCurrentPercentage() float64 {
	return am.priceMonitor.GetCurrentPercentage()
}

// System operations

// GetStats returns system statistics including current price and active alerts.
//
// Example usage:
//
//	stats, err := manager.GetStats()
//	if err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	log.Printf("Active alerts: %d", stats["active_alerts"])
func (am *AlertManager) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get current price
	price, err := am.GetCurrentPrice()
	if err == nil {
		stats["current_price"] = price.Price
		stats["price_change"] = price.PriceChangePercent
	}

	// Get active alerts count
	alerts, err := am.alertRepo.GetActiveAlerts()
	if err == nil {
		stats["active_alerts"] = len(alerts)
	}

	return stats, nil
}