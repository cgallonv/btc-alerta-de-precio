package mocks

import (
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockPriceClient is a mock implementation of interfaces.PriceClient
type MockPriceClient struct {
	mock.Mock
}

func (m *MockPriceClient) GetCurrentPrice() (*bitcoin.PriceData, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bitcoin.PriceData), args.Error(1)
}

func (m *MockPriceClient) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	args := m.Called(days)
	return args.Get(0).([]bitcoin.PriceData), args.Error(1)
}

// MockNotificationSender is a mock implementation of interfaces.NotificationSender
type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) SendAlert(data *notifications.NotificationData) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockNotificationSender) SendWebPushNotification(subscriptions []storage.WebPushSubscription, data *notifications.NotificationData) error {
	args := m.Called(subscriptions, data)
	return args.Error(0)
}

func (m *MockNotificationSender) TestTelegramNotification() error {
	args := m.Called()
	return args.Error(0)
}

// MockAlertEvaluator is a mock implementation of interfaces.AlertEvaluator
type MockAlertEvaluator struct {
	mock.Mock
}

func (m *MockAlertEvaluator) ShouldTrigger(alert *storage.Alert, currentPrice, previousPrice float64) bool {
	args := m.Called(alert, currentPrice, previousPrice)
	return args.Bool(0)
}

// MockConfigProvider is a mock implementation of interfaces.ConfigProvider
type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) GetCheckInterval() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfigProvider) IsEmailNotificationsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigProvider) IsWebPushNotificationsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigProvider) IsTelegramNotificationsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}
