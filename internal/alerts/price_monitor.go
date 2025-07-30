// Package alerts provides functionality for monitoring Bitcoin prices
// and managing price-based alerts.
package alerts

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	"github.com/cgallonv/btc-alerta-de-precio/internal/errors"
	"github.com/cgallonv/btc-alerta-de-precio/internal/interfaces"
)

var (
	lastLogTime     time.Time
	lastLoggedPrice float64
	logMutex        sync.RWMutex
)

// PriceMonitor handles price monitoring and caching operations.
// It fetches Bitcoin prices at regular intervals, maintains a price history cache,
// and notifies registered callbacks when prices are updated.
//
// Example usage:
//
//	monitor := NewPriceMonitor(configProvider, 20)
//	monitor.AddPriceUpdateCallback(func(price *bitcoin.PriceData) {
//	    log.Printf("New price: $%.2f", price.Price)
//	})
//	if err := monitor.Start(context.Background()); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
//	defer monitor.Stop()
type PriceMonitor struct {
	binanceClient  *bitcoin.BinanceClient
	configProvider interfaces.ConfigProvider

	// Monitoring state
	isMonitoring  bool
	stopChannel   chan struct{}
	monitoringMux sync.RWMutex

	// Price data cache (replaces database storage)
	priceCache   *PriceCache
	lastPrice    *bitcoin.PriceData
	lastPriceMux sync.RWMutex

	// Current percentage (updated with every price fetch)
	currentPercentage    float64
	currentPercentageMux sync.RWMutex

	// Callbacks for price updates
	priceUpdateCallbacks []PriceUpdateCallback
	callbackMux          sync.RWMutex
}

// PriceUpdateCallback is called when price is updated.
// The callback receives the current price data from Binance.
//
// Example usage:
//
//	monitor.AddPriceUpdateCallback(func(price *bitcoin.PriceData) {
//	    if price.PriceChangePercent > 5.0 {
//	        log.Printf("Large price increase: %+.2f%%", price.PriceChangePercent)
//	    }
//	})
type PriceUpdateCallback func(current *bitcoin.PriceData)

// NewPriceMonitor creates a new price monitoring service with cache.
// The cacheSize parameter determines how many historical price entries to keep.
// If cacheSize is <= 0, it defaults to 20 entries.
//
// Example usage:
//
//	monitor := NewPriceMonitor(configProvider, 20)
//	if err := monitor.Start(context.Background()); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func NewPriceMonitor(
	configProvider interfaces.ConfigProvider,
	cacheSize int,
	tickerStorage *bitcoin.TickerStorage,
) *PriceMonitor {
	if cacheSize <= 0 {
		cacheSize = 20 // Default to 20 entries
	}

	// Create Binance client with API credentials
	apiKey := configProvider.GetString("binance.api_key")
	apiSecret := configProvider.GetString("binance.api_secret")
	binanceClient := bitcoin.NewBinanceClient(apiKey, apiSecret, tickerStorage)

	return &PriceMonitor{
		binanceClient:        binanceClient,
		configProvider:       configProvider,
		priceCache:           NewPriceCache(cacheSize),
		stopChannel:          make(chan struct{}),
		priceUpdateCallbacks: make([]PriceUpdateCallback, 0),
	}
}

// Start begins price monitoring with the configured interval.
// It returns an error if monitoring is already active.
// The monitoring continues until Stop is called or the context is cancelled.
//
// Example usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	if err := monitor.Start(ctx); err != nil {
//	    log.Printf("Error: %v", err)
//	    return
//	}
func (pm *PriceMonitor) Start(ctx context.Context) error {
	pm.monitoringMux.Lock()
	defer pm.monitoringMux.Unlock()

	if pm.isMonitoring {
		return errors.NewAppError("MONITOR_ALREADY_RUNNING", "Price monitoring is already active")
	}

	pm.isMonitoring = true
	interval := pm.configProvider.GetCheckInterval()

	log.Printf("🔄 Starting Bitcoin price monitoring (interval: %v)", interval)

	go pm.monitoringLoop(ctx, interval)

	return nil
}

// Stop stops price monitoring.
// It's safe to call Stop multiple times.
// After stopping, the monitor can be restarted with Start.
//
// Example usage:
//
//	defer monitor.Stop()
func (pm *PriceMonitor) Stop() error {
	pm.monitoringMux.Lock()
	defer pm.monitoringMux.Unlock()

	if !pm.isMonitoring {
		return nil
	}

	pm.isMonitoring = false
	close(pm.stopChannel)
	pm.stopChannel = make(chan struct{}) // Recreate channel for potential restart
	return nil
}

// IsMonitoring returns true if price monitoring is active.
//
// Example usage:
//
//	if monitor.IsMonitoring() {
//	    log.Printf("Price monitoring is active")
//	}
func (pm *PriceMonitor) IsMonitoring() bool {
	pm.monitoringMux.RLock()
	defer pm.monitoringMux.RUnlock()
	return pm.isMonitoring
}

// GetLastPrice returns the last cached price data.
// Returns nil if no price data is available.
//
// Example usage:
//
//	if price := monitor.GetLastPrice(); price != nil {
//	    log.Printf("Last price: $%.2f", price.Price)
//	}
func (pm *PriceMonitor) GetLastPrice() *bitcoin.PriceData {
	pm.lastPriceMux.RLock()
	defer pm.lastPriceMux.RUnlock()
	return pm.lastPrice
}

// GetCurrentPercentage returns the current price change percentage.
//
// Example usage:
//
//	change := monitor.GetCurrentPercentage()
//	log.Printf("24h change: %+.2f%%", change)
func (pm *PriceMonitor) GetCurrentPercentage() float64 {
	pm.currentPercentageMux.RLock()
	defer pm.currentPercentageMux.RUnlock()
	return pm.currentPercentage
}

// GetPriceHistory returns cached price history.
// The limit parameter determines how many entries to return.
//
// Example usage:
//
//	history := monitor.GetPriceHistory(24) // Get last 24 entries
//	for _, entry := range history {
//	    log.Printf("Price at %s: $%.2f", entry.Timestamp, entry.Price)
//	}
func (pm *PriceMonitor) GetPriceHistory(limit int) []interfaces.PriceCacheEntry {
	return pm.priceCache.GetHistory(limit)
}

// AddPriceUpdateCallback adds a callback for price updates.
// The callback will be called whenever a new price is fetched.
//
// Example usage:
//
//	monitor.AddPriceUpdateCallback(func(price *bitcoin.PriceData) {
//	    if price.Price > 50000 {
//	        log.Printf("Bitcoin price above $50,000!")
//	    }
//	})
func (pm *PriceMonitor) AddPriceUpdateCallback(callback PriceUpdateCallback) {
	pm.callbackMux.Lock()
	defer pm.callbackMux.Unlock()
	pm.priceUpdateCallbacks = append(pm.priceUpdateCallbacks, callback)
}

// monitoringLoop is the main price monitoring loop.
// It fetches prices at regular intervals and notifies callbacks.
func (pm *PriceMonitor) monitoringLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// First check immediately
	pm.checkAndUpdatePrice()

	for {
		select {
		case <-ticker.C:
			pm.checkAndUpdatePrice()
		case <-pm.stopChannel:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkAndUpdatePrice fetches current price and updates both price and percentage.
func (pm *PriceMonitor) checkAndUpdatePrice() {
	currentPrice, err := pm.binanceClient.GetCurrentPrice()
	if err != nil {
		log.Printf("❌ Error fetching Bitcoin price: %v", err)
		return
	}

	// Add to cache (replaces database storage)
	pm.priceCache.Add(currentPrice)

	// Update cached price
	pm.lastPriceMux.Lock()
	pm.lastPrice = currentPrice
	pm.lastPriceMux.Unlock()

	// Update percentage
	pm.currentPercentageMux.Lock()
	pm.currentPercentage = currentPrice.PriceChangePercent
	pm.currentPercentageMux.Unlock()

	// Log price update only when there's a significant change (>0.1%) or every 5 minutes
	if shouldLogPrice(currentPrice) {
		log.Printf("💰 BTC: $%.2f (%+.2f%%) [%s]",
			currentPrice.Price,
			currentPrice.PriceChangePercent,
			currentPrice.Source)
	}

	// Notify callbacks of price update
	pm.notifyPriceUpdateCallbacks(currentPrice)
}

// shouldLogPrice determines if we should log the current price update.
// Returns true if:
// 1. First price update (lastLoggedPrice is 0)
// 2. Price changed by more than 0.1%
// 3. It's been more than 5 minutes since last log
func shouldLogPrice(price *bitcoin.PriceData) bool {
	logMutex.Lock()
	defer logMutex.Unlock()

	now := time.Now()

	significantChange := lastLoggedPrice != 0 &&
		abs((price.Price-lastLoggedPrice)/lastLoggedPrice) > 0.001
	timeToLog := now.Sub(lastLogTime) >= 5*time.Minute

	shouldLog := lastLoggedPrice == 0 || significantChange || timeToLog

	if shouldLog {
		lastLogTime = now
		lastLoggedPrice = price.Price
	}

	return shouldLog
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// notifyPriceUpdateCallbacks notifies all registered callbacks.
// Each callback is run in its own goroutine to prevent blocking.
func (pm *PriceMonitor) notifyPriceUpdateCallbacks(current *bitcoin.PriceData) {
	pm.callbackMux.RLock()
	callbacks := make([]PriceUpdateCallback, len(pm.priceUpdateCallbacks))
	copy(callbacks, pm.priceUpdateCallbacks)
	pm.callbackMux.RUnlock()

	for _, callback := range callbacks {
		go func(cb PriceUpdateCallback) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("❌ Callback panic: %v", r)
				}
			}()
			cb(current)
		}(callback)
	}
}
