package adapters

import (
	"btc-alerta-de-precio/internal/bitcoin"
	apperrors "btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/storage"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertEvaluatorImpl_ShouldTrigger(t *testing.T) {
	evaluator := NewAlertEvaluator()

	tests := []struct {
		name          string
		alert         *storage.Alert
		currentPrice  float64
		previousPrice float64
		expected      bool
	}{
		{
			name: "inactive alert should not trigger",
			alert: &storage.Alert{
				Type:        "above",
				TargetPrice: 50000,
				IsActive:    false,
			},
			currentPrice:  55000,
			previousPrice: 45000,
			expected:      false,
		},
		{
			name: "already triggered alert should not trigger again",
			alert: &storage.Alert{
				Type:          "above",
				TargetPrice:   50000,
				IsActive:      true,
				LastTriggered: &time.Time{}, // Non-nil means already triggered
			},
			currentPrice:  55000,
			previousPrice: 45000,
			expected:      false,
		},
		{
			name: "above alert should trigger when price exceeds target",
			alert: &storage.Alert{
				Type:        "above",
				TargetPrice: 50000,
				IsActive:    true,
			},
			currentPrice:  55000,
			previousPrice: 45000,
			expected:      true,
		},
		{
			name: "above alert should not trigger when price below target",
			alert: &storage.Alert{
				Type:        "above",
				TargetPrice: 50000,
				IsActive:    true,
			},
			currentPrice:  45000,
			previousPrice: 40000,
			expected:      false,
		},
		{
			name: "below alert should trigger when price falls below target",
			alert: &storage.Alert{
				Type:        "below",
				TargetPrice: 50000,
				IsActive:    true,
			},
			currentPrice:  45000,
			previousPrice: 55000,
			expected:      true,
		},
		{
			name: "below alert should not trigger when price above target",
			alert: &storage.Alert{
				Type:        "below",
				TargetPrice: 50000,
				IsActive:    true,
			},
			currentPrice:  55000,
			previousPrice: 45000,
			expected:      false,
		},
		{
			name: "change alert should trigger on positive percentage change",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0, // 5% positive change threshold
				IsActive:   true,
			},
			currentPrice:  52500, // 5% increase from 50000
			previousPrice: 50000,
			expected:      true,
		},
		{
			name: "change alert should NOT trigger on negative percentage change when expecting positive",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0, // 5% positive change threshold
				IsActive:   true,
			},
			currentPrice:  47500, // 5% decrease from 50000 - should NOT trigger
			previousPrice: 50000,
			expected:      false,
		},
		{
			name: "negative percentage alert should trigger on sufficient negative change",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			currentPrice:  47500, // 5% decrease from 50000
			previousPrice: 50000,
			expected:      true,
		},
		{
			name: "negative percentage alert should NOT trigger on positive change",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			currentPrice:  52500, // 5% increase from 50000 - should NOT trigger
			previousPrice: 50000,
			expected:      false,
		},
		{
			name: "negative percentage alert should NOT trigger on insufficient negative change",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			currentPrice:  48000, // 4% decrease from 50000 - not enough
			previousPrice: 50000,
			expected:      false,
		},
		{
			name: "change alert should not trigger when change below positive threshold",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0, // 5% positive change threshold
				IsActive:   true,
			},
			currentPrice:  51000, // 2% increase from 50000 - not enough
			previousPrice: 50000,
			expected:      false,
		},
		{
			name: "zero percentage alert should never trigger",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 0.0, // Invalid zero percentage
				IsActive:   true,
			},
			currentPrice:  55000,
			previousPrice: 50000,
			expected:      false,
		},
		{
			name: "change alert should not trigger when previous price is zero",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0,
				IsActive:   true,
			},
			currentPrice:  50000,
			previousPrice: 0,
			expected:      false,
		},
		{
			name: "unknown alert type should not trigger",
			alert: &storage.Alert{
				Type:     "unknown",
				IsActive: true,
			},
			currentPrice:  50000,
			previousPrice: 45000,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.ShouldTrigger(tt.alert, tt.currentPrice, tt.previousPrice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBitcoinClientAdapter_GetCurrentPrice(t *testing.T) {
	t.Run("successful price retrieval", func(t *testing.T) {
		// Setup
		mockClient := &mockBitcoinClient{
			price: &bitcoin.PriceData{
				Price:     50000.0,
				Currency:  "USD",
				Timestamp: time.Now(),
				Source:    "Test",
			},
		}

		adapter := &testBitcoinClientAdapter{client: mockClient}

		// Execute
		price, err := adapter.GetCurrentPrice()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 50000.0, price.Price)
		assert.Equal(t, "USD", price.Currency)
	})

	t.Run("client error handling", func(t *testing.T) {
		// Setup
		mockClient := &mockBitcoinClient{
			err: errors.New("network error"),
		}

		adapter := &testBitcoinClientAdapter{client: mockClient}

		// Execute
		price, err := adapter.GetCurrentPrice()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, price)
		assert.Contains(t, err.Error(), "Failed to get current price")
	})
}

// Test adapter for bitcoin client
type testBitcoinClientAdapter struct {
	client *mockBitcoinClient
}

func (a *testBitcoinClientAdapter) GetCurrentPrice() (*bitcoin.PriceData, error) {
	price, err := a.client.GetCurrentPrice()
	if err != nil {
		return nil, apperrors.WrapError(err, "PRICE_CLIENT_ERROR", "Failed to get current price")
	}
	return price, nil
}

func (a *testBitcoinClientAdapter) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	history, err := a.client.GetPriceHistory(days)
	if err != nil {
		return nil, apperrors.WrapError(err, "PRICE_CLIENT_HISTORY_ERROR", "Failed to get price history")
	}
	return history, nil
}

// Mock implementation for testing
type mockBitcoinClient struct {
	price *bitcoin.PriceData
	err   error
}

func (m *mockBitcoinClient) GetCurrentPrice() (*bitcoin.PriceData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.price, nil
}

func (m *mockBitcoinClient) GetPriceHistory(days int) ([]bitcoin.PriceData, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return mock data
	return []bitcoin.PriceData{*m.price}, nil
}
