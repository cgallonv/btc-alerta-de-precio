package alerts

import (
	"context"
	"log"
	"sync"
	"time"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
)

var (
	lastLogTime     time.Time
	lastLoggedPrice float64
	logMutex        sync.RWMutex
)

// PriceMonitor handles price monitoring and caching operations
type PriceMonitor struct {
	priceClient    interfaces.PriceClient
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

// PriceUpdateCallback is called when price is updated (simplified to only pass current price)
type PriceUpdateCallback func(current *bitcoin.PriceData)

// NewPriceMonitor creates a new price monitoring service with cache
func NewPriceMonitor(
	priceClient interfaces.PriceClient,
	configProvider interfaces.ConfigProvider,
	cacheSize int,
) *PriceMonitor {
	if cacheSize <= 0 {
		cacheSize = 20 // Default to 20 entries
	}

	return &PriceMonitor{
		priceClient:          priceClient,
		configProvider:       configProvider,
		priceCache:           NewPriceCache(cacheSize),
		stopChannel:          make(chan struct{}),
		priceUpdateCallbacks: make([]PriceUpdateCallback, 0),
	}
}

// Start begins price monitoring with the configured interval
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

// Stop stops price monitoring
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

// IsMonitoring returns true if monitoring is active
func (pm *PriceMonitor) IsMonitoring() bool {
	pm.monitoringMux.RLock()
	defer pm.monitoringMux.RUnlock()
	return pm.isMonitoring
}

// GetLastPrice returns the last cached price data
func (pm *PriceMonitor) GetLastPrice() *bitcoin.PriceData {
	pm.lastPriceMux.RLock()
	defer pm.lastPriceMux.RUnlock()
	return pm.lastPrice
}

// GetCurrentPercentage returns the current price change percentage
func (pm *PriceMonitor) GetCurrentPercentage() float64 {
	pm.currentPercentageMux.RLock()
	defer pm.currentPercentageMux.RUnlock()
	return pm.currentPercentage
}

// GetPriceHistory returns cached price history
func (pm *PriceMonitor) GetPriceHistory(limit int) []interfaces.PriceCacheEntry {
	return pm.priceCache.GetHistory(limit)
}

// AddPriceUpdateCallback adds a callback for price updates
func (pm *PriceMonitor) AddPriceUpdateCallback(callback PriceUpdateCallback) {
	pm.callbackMux.Lock()
	defer pm.callbackMux.Unlock()
	pm.priceUpdateCallbacks = append(pm.priceUpdateCallbacks, callback)
}

// monitoringLoop is the unified monitoring loop for both price and percentage
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

// checkAndUpdatePrice fetches current price and updates both price and percentage
func (pm *PriceMonitor) checkAndUpdatePrice() {
	currentPrice, err := pm.priceClient.GetCurrentPrice()
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

	// Update percentage if source is Binance (unified in single operation)
	if currentPrice.Source == "Binance" {
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
	}

	// Notify callbacks of price update
	pm.notifyPriceUpdateCallbacks(currentPrice)
}

// shouldLogPrice determines if we should log the current price update
func shouldLogPrice(price *bitcoin.PriceData) bool {
	logMutex.Lock()
	defer logMutex.Unlock()

	now := time.Now()

	// Log if:
	// 1. First price update (lastLoggedPrice is 0)
	// 2. Price changed by more than 0.1%
	// 3. It's been more than 5 minutes since last log
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

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// notifyPriceUpdateCallbacks notifies all registered callbacks
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
