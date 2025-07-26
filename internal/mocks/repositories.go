package mocks

import (
	"btc-alerta-de-precio/internal/storage"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockAlertRepository is a mock implementation of interfaces.AlertRepository
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) CreateAlert(alert *storage.Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAlert(id uint) (*storage.Alert, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlerts() ([]storage.Alert, error) {
	args := m.Called()
	return args.Get(0).([]storage.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetActiveAlerts() ([]storage.Alert, error) {
	args := m.Called()
	return args.Get(0).([]storage.Alert), args.Error(1)
}

func (m *MockAlertRepository) UpdateAlert(alert *storage.Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

func (m *MockAlertRepository) DeleteAlert(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAlertRepository) ToggleAlert(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockPriceRepository is a mock implementation of interfaces.PriceRepository
type MockPriceRepository struct {
	mock.Mock
}

func (m *MockPriceRepository) SavePriceHistory(price *storage.PriceHistory) error {
	args := m.Called(price)
	return args.Error(0)
}

func (m *MockPriceRepository) GetLatestPrice() (*storage.PriceHistory, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.PriceHistory), args.Error(1)
}

func (m *MockPriceRepository) GetPriceHistory(limit int) ([]storage.PriceHistory, error) {
	args := m.Called(limit)
	return args.Get(0).([]storage.PriceHistory), args.Error(1)
}

func (m *MockPriceRepository) GetPriceHistoryByDateRange(start, end time.Time) ([]storage.PriceHistory, error) {
	args := m.Called(start, end)
	return args.Get(0).([]storage.PriceHistory), args.Error(1)
}

func (m *MockPriceRepository) CleanOldPriceHistory(days int) error {
	args := m.Called(days)
	return args.Error(0)
}

// MockNotificationRepository is a mock implementation of interfaces.NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) LogNotification(log *storage.NotificationLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationLogs(alertID uint, limit int) ([]storage.NotificationLog, error) {
	args := m.Called(alertID, limit)
	return args.Get(0).([]storage.NotificationLog), args.Error(1)
}
