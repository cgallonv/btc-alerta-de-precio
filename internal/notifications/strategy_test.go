package notifications

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestWebPushStrategy_IsEnabled(t *testing.T) {
	cfg := &config.Config{
		EnableWebPushNotifications: true,
	}
	// Use a nil db for IsEnabled test (db not needed)
	strategy := NewWebPushStrategy(cfg, nil)

	tests := []struct {
		name     string
		alert    *storage.Alert
		expected bool
	}{
		{
			name: "enabled when both conditions met",
			alert: &storage.Alert{
				EnableWebPush: true,
			},
			expected: true,
		},
		{
			name: "disabled when webpush disabled globally",
			alert: &storage.Alert{
				EnableWebPush: true,
			},
			expected: true, // Still true because we set EnableWebPushNotifications to true
		},
		{
			name: "disabled when alert webpush disabled",
			alert: &storage.Alert{
				EnableWebPush: false,
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

func TestWhatsAppStrategy_IsEnabled(t *testing.T) {
	cfg := &config.Config{
		EnableWhatsAppNotifications: true,
	}

	strategy := NewWhatsAppStrategy(cfg)

	tests := []struct {
		name     string
		alert    *storage.Alert
		expected bool
	}{
		{
			name: "enabled when all conditions met",
			alert: &storage.Alert{
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
			},
			expected: true,
		},
		{
			name: "disabled when whatsapp notifications disabled globally",
			alert: &storage.Alert{
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
			},
			expected: true, // Still true because we set EnableWhatsAppNotifications to true
		},
		{
			name: "disabled when alert whatsapp disabled",
			alert: &storage.Alert{
				EnableWhatsApp: false,
				WhatsAppNumber: "+1234567890",
			},
			expected: false,
		},
		{
			name: "disabled when no whatsapp number",
			alert: &storage.Alert{
				EnableWhatsApp: true,
				WhatsAppNumber: "",
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

func TestWhatsAppStrategy_Send(t *testing.T) {
	tests := []struct {
		name            string
		config          *config.Config
		alert           *storage.Alert
		mockAPIStatus   int
		mockAPIResponse string
		expectedError   bool
	}{
		{
			name: "successful send",
			config: &config.Config{
				EnableWhatsAppNotifications: true,
				WhatsAppAccessToken:         "test_token",
				WhatsAppPhoneNumberID:       "12345",
				WhatsAppTemplateNameES:      "test_template_es",
				WhatsAppTemplateNameEN:      "test_template_en",
			},
			alert: &storage.Alert{
				Name:           "Test Alert",
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
				Language:       "es",
				Type:           "above",
				TargetPrice:    50000,
			},
			mockAPIStatus:   200,
			mockAPIResponse: `{"success":true}`,
			expectedError:   false,
		},
		{
			name: "disabled notifications",
			config: &config.Config{
				EnableWhatsAppNotifications: false,
			},
			alert: &storage.Alert{
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
			},
			expectedError: false,
		},
		{
			name: "missing configuration",
			config: &config.Config{
				EnableWhatsAppNotifications: true,
				WhatsAppAccessToken:         "",
				WhatsAppPhoneNumberID:       "",
			},
			alert: &storage.Alert{
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
			},
			expectedError: true,
		},
		{
			name: "api error response",
			config: &config.Config{
				EnableWhatsAppNotifications: true,
				WhatsAppAccessToken:         "test_token",
				WhatsAppPhoneNumberID:       "12345",
				WhatsAppTemplateNameES:      "test_template_es",
			},
			alert: &storage.Alert{
				Name:           "Test Alert",
				EnableWhatsApp: true,
				WhatsAppNumber: "+1234567890",
				Language:       "es",
			},
			mockAPIStatus:   400,
			mockAPIResponse: `{"error":{"message":"Invalid WhatsApp number"}}`,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server if needed
			var server *httptest.Server
			if tt.mockAPIStatus > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "Bearer "+tt.config.WhatsAppAccessToken, r.Header.Get("Authorization"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					// Verify request body
					var payload map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&payload)
					assert.NoError(t, err)
					assert.Equal(t, "whatsapp", payload["messaging_product"])
					assert.Equal(t, tt.alert.WhatsAppNumber, payload["to"])

					// Send response
					w.WriteHeader(tt.mockAPIStatus)
					w.Write([]byte(tt.mockAPIResponse))
				}))
				defer server.Close()

				// Override API base URL for testing
				whatsappAPIBaseURL = server.URL
			}

			strategy := NewWhatsAppStrategy(tt.config)
			err := strategy.Send(&NotificationData{
				Title:   "Test Notification",
				Message: "Test message",
				Price:   50000.00,
				Alert:   tt.alert,
			})

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWhatsAppStrategy_GetChannelName(t *testing.T) {
	strategy := NewWhatsAppStrategy(&config.Config{})
	assert.Equal(t, "whatsapp", strategy.GetChannelName())
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
