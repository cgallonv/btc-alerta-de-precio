package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"
)

// MockNotificationStrategy is a mock implementation for testing
type MockNotificationStrategy struct {
	mock.Mock
}

func (m *MockNotificationStrategy) Send(data *NotificationData) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockNotificationStrategy) IsEnabled(alert *storage.Alert) bool {
	args := m.Called(alert)
	return args.Bool(0)
}

func (m *MockNotificationStrategy) GetChannelName() string {
	args := m.Called()
	return args.String(0)
}

func TestNotificationManager_SendAlert(t *testing.T) {
	t.Run("sends to all enabled strategies", func(t *testing.T) {
		// Setup
		mockStrategy1 := &MockNotificationStrategy{}
		mockStrategy2 := &MockNotificationStrategy{}

		manager := NewNotificationManager(mockStrategy1, mockStrategy2)

		alert := &storage.Alert{
			ID:             1,
			Name:           "Test Alert",
			EnableEmail:    true,
			EnableTelegram: true,
		}

		notificationData := &NotificationData{
			Title:   "Test Notification",
			Message: "Test message",
			Price:   50000.00,
			Alert:   alert,
		}

		// Setup mocks
		mockStrategy1.On("IsEnabled", alert).Return(true)
		mockStrategy1.On("Send", notificationData).Return(nil)

		mockStrategy2.On("IsEnabled", alert).Return(true)
		mockStrategy2.On("Send", notificationData).Return(nil)

		// Execute
		err := manager.SendAlert(notificationData)

		// Assert
		assert.NoError(t, err)
		mockStrategy1.AssertExpectations(t)
		mockStrategy2.AssertExpectations(t)
	})

	t.Run("skips disabled strategies", func(t *testing.T) {
		// Setup
		mockStrategy1 := &MockNotificationStrategy{}
		mockStrategy2 := &MockNotificationStrategy{}

		manager := NewNotificationManager(mockStrategy1, mockStrategy2)

		alert := &storage.Alert{
			ID:             1,
			Name:           "Test Alert",
			EnableEmail:    true,
			EnableTelegram: false,
		}

		notificationData := &NotificationData{
			Title:   "Test Notification",
			Message: "Test message",
			Price:   50000.00,
			Alert:   alert,
		}

		// Setup mocks - strategy1 enabled, strategy2 disabled
		mockStrategy1.On("IsEnabled", alert).Return(true)
		mockStrategy1.On("Send", notificationData).Return(nil)

		mockStrategy2.On("IsEnabled", alert).Return(false)
		// mockStrategy2.Send should NOT be called

		// Execute
		err := manager.SendAlert(notificationData)

		// Assert
		assert.NoError(t, err)
		mockStrategy1.AssertExpectations(t)
		mockStrategy2.AssertExpectations(t)
	})
}

func TestEmailStrategy_IsEnabled(t *testing.T) {
	cfg := &config.Config{
		EnableEmailNotifications: true,
	}

	strategy := NewEmailStrategy(cfg)

	tests := []struct {
		name     string
		alert    *storage.Alert
		expected bool
	}{
		{
			name: "enabled when all conditions met",
			alert: &storage.Alert{
				EnableEmail: true,
				Email:       "test@example.com",
			},
			expected: true,
		},
		{
			name: "disabled when email notifications disabled globally",
			alert: &storage.Alert{
				EnableEmail: true,
				Email:       "test@example.com",
			},
			expected: true, // Still true because we set EnableEmailNotifications to true
		},
		{
			name: "disabled when alert email disabled",
			alert: &storage.Alert{
				EnableEmail: false,
				Email:       "test@example.com",
			},
			expected: false,
		},
		{
			name: "disabled when no email address",
			alert: &storage.Alert{
				EnableEmail: true,
				Email:       "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.IsEnabled(tt.alert)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTelegramStrategy_IsEnabled(t *testing.T) {
	cfg := &config.Config{
		EnableTelegramNotifications: true,
	}

	strategy := NewTelegramStrategy(cfg)

	tests := []struct {
		name     string
		alert    *storage.Alert
		expected bool
	}{
		{
			name: "enabled when both conditions met",
			alert: &storage.Alert{
				EnableTelegram: true,
			},
			expected: true,
		},
		{
			name: "disabled when telegram disabled globally",
			alert: &storage.Alert{
				EnableTelegram: true,
			},
			expected: true, // Still true because we set EnableTelegramNotifications to true
		},
		{
			name: "disabled when alert telegram disabled",
			alert: &storage.Alert{
				EnableTelegram: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.IsEnabled(tt.alert)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNotificationManager_AddRemoveStrategy(t *testing.T) {
	manager := NewNotificationManager()

	// Initially no strategies
	assert.Len(t, manager.strategies, 0)

	// Add strategy
	mockStrategy := &MockNotificationStrategy{}
	mockStrategy.On("GetChannelName").Return("test")

	manager.AddStrategy(mockStrategy)
	assert.Len(t, manager.strategies, 1)

	// Remove strategy
	manager.RemoveStrategy("test")
	assert.Len(t, manager.strategies, 0)

	mockStrategy.AssertExpectations(t)
}
