package adapters

import (
	"errors"
	"testing"
	"time"

	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	apperrors "github.com/cgallonv/btc-alerta-de-precio/internal/errors"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertEvaluatorImpl_ShouldTrigger(t *testing.T) {
	evaluator := NewAlertEvaluator()

	tests := []struct {
		name      string
		alert     *storage.Alert
		priceData *bitcoin.PriceData
		expected  bool
	}{
		{
			name: "inactive alert should not trigger",
			alert: &storage.Alert{
				Type:     "above",
				IsActive: false,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 2.5,
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "already triggered alert should not trigger again",
			alert: &storage.Alert{
				Type:          "above",
				IsActive:      true,
				LastTriggered: &[]time.Time{time.Now()}[0],
			},
			priceData: &bitcoin.PriceData{
				Price:              60000,
				PriceChangePercent: 5.0,
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "above alert should trigger when price exceeds target",
			alert: &storage.Alert{
				Type:        "above",
				TargetPrice: 45000,
				IsActive:    true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 2.5,
				Source:             "Binance",
			},
			expected: true,
		},
		{
			name: "above alert should not trigger when price below target",
			alert: &storage.Alert{
				Type:        "above",
				TargetPrice: 55000,
				IsActive:    true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 2.5,
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "below alert should trigger when price below target",
			alert: &storage.Alert{
				Type:        "below",
				TargetPrice: 55000,
				IsActive:    true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: -1.5,
				Source:             "Binance",
			},
			expected: true,
		},
		{
			name: "below alert should not trigger when price above target",
			alert: &storage.Alert{
				Type:        "below",
				TargetPrice: 45000,
				IsActive:    true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 3.0,
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "positive percentage alert should trigger on sufficient positive change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 3.0, // 3% positive change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 4.0, // 4% positive change from Binance API
				Source:             "Binance",
			},
			expected: true,
		},
		{
			name: "positive percentage alert should NOT trigger on negative change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 3.0, // 3% positive change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: -5.0, // Negative change should NOT trigger positive alert
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "negative percentage alert should trigger on sufficient negative change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: -6.0, // -6% negative change from Binance API
				Source:             "Binance",
			},
			expected: true,
		},
		{
			name: "negative percentage alert should NOT trigger on positive change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 5.0, // Positive change should NOT trigger negative alert
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "negative percentage alert should NOT trigger on insufficient negative change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: -5.0, // -5% negative change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: -3.0, // -3% is not enough for -5% threshold
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "positive percentage alert should NOT trigger on insufficient positive change (Binance)",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0, // 5% positive change threshold
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 3.0, // 3% is not enough for 5% threshold
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "zero percentage alert should never trigger",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 0.0, // Invalid zero percentage
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 10.0,
				Source:             "Binance",
			},
			expected: false,
		},
		{
			name: "change alert should not trigger for non-Binance sources",
			alert: &storage.Alert{
				Type:       "change",
				Percentage: 5.0,
				IsActive:   true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 0.0, // CoinDesk/CoinGecko don't provide percentage
				Source:             "CoinDesk",
			},
			expected: false,
		},
		{
			name: "unknown alert type should not trigger",
			alert: &storage.Alert{
				Type:     "unknown",
				IsActive: true,
			},
			priceData: &bitcoin.PriceData{
				Price:              50000,
				PriceChangePercent: 5.0,
				Source:             "Binance",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.ShouldTrigger(tt.alert, tt.priceData)
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
