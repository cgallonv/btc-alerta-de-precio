package notifications

import (
	"github.com/cgallonv/btc-alerta-de-precio/internal/errors"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"
)

// NotificationStrategy defines the interface for different notification channels
type NotificationStrategy interface {
	Send(data *NotificationData) error
	IsEnabled(alert *storage.Alert) bool
	GetChannelName() string
}

// NotificationManager coordinates multiple notification strategies
type NotificationManager struct {
	strategies []NotificationStrategy
}

// NewNotificationManager creates a new notification manager with multiple strategies
func NewNotificationManager(strategies ...NotificationStrategy) *NotificationManager {
	return &NotificationManager{
		strategies: strategies,
	}
}

// SendAlert sends notifications through all enabled strategies
func (nm *NotificationManager) SendAlert(data *NotificationData) error {
	var sendErrors []error
	successCount := 0

	for _, strategy := range nm.strategies {
		if !strategy.IsEnabled(data.Alert) {
			continue
		}

		if err := strategy.Send(data); err != nil {
			strategyErr := errors.WrapError(err, "STRATEGY_SEND_ERROR", "Failed to send via "+strategy.GetChannelName())
			sendErrors = append(sendErrors, strategyErr)
		} else {
			successCount++
		}
	}

	// If all strategies failed, return combined error
	if successCount == 0 && len(sendErrors) > 0 {
		return errors.CombineErrors(sendErrors)
	}

	// If some strategies failed but at least one succeeded, log errors but don't fail
	if len(sendErrors) > 0 {
		for _, err := range sendErrors {
			// Log individual strategy errors (these would be logged elsewhere)
			_ = err
		}
	}

	return nil
}

// AddStrategy adds a new notification strategy
func (nm *NotificationManager) AddStrategy(strategy NotificationStrategy) {
	nm.strategies = append(nm.strategies, strategy)
}

// RemoveStrategy removes a notification strategy by channel name
func (nm *NotificationManager) RemoveStrategy(channelName string) {
	for i, strategy := range nm.strategies {
		if strategy.GetChannelName() == channelName {
			// Remove strategy from slice
			nm.strategies = append(nm.strategies[:i], nm.strategies[i+1:]...)
			break
		}
	}
}
