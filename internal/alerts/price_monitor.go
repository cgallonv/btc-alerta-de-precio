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

	// Percentage monitoring
	isPercentageMonitoring  bool
	percentageStopChannel   chan struct{}
	percentageMonitoringMux sync.RWMutex
	currentPercentage       float64
	currentPercentageMux    sync.RWMutex

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
		priceClient:           priceClient,
		configProvider:        configProvider,
		priceCache:            NewPriceCache(cacheSize),
		stopChannel:           make(chan struct{}),
		percentageStopChannel: make(chan struct{}),
		priceUpdateCallbacks:  make([]PriceUpdateCallback, 0),
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

	log.Printf("Starting Bitcoin price monitoring every %v", interval)

	go pm.monitoringLoop(ctx, interval)

	// Also start percentage monitoring if needed
	go pm.startPercentageMonitoring(ctx)

	return nil
}

// Stop stops price monitoring
func (pm *PriceMonitor) Stop() error {
	pm.monitoringMux.Lock()
	defer pm.monitoringMux.Unlock()

	if !pm.isMonitoring {
		return nil
	}

	log.Println("Stopping price monitoring...")
	pm.isMonitoring = false

	close(pm.stopChannel)
	close(pm.percentageStopChannel)

	// Recreate channels for potential restart
	pm.stopChannel = make(chan struct{})
	pm.percentageStopChannel = make(chan struct{})

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

// monitoringLoop is the main monitoring loop
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
			log.Println("Price monitoring stopped")
			return
		case <-ctx.Done():
			log.Println("Price monitoring cancelled via context")
			return
		}
	}
}

// checkAndUpdatePrice fetches current price and updates cache
func (pm *PriceMonitor) checkAndUpdatePrice() {
	currentPrice, err := pm.priceClient.GetCurrentPrice()
	if err != nil {
		log.Printf("Error fetching Bitcoin price: %v", err)
		return
	}

	log.Printf("Current Bitcoin price: %s", currentPrice.String())

	// Add to cache (replaces database storage)
	pm.priceCache.Add(currentPrice)

	// Update cached price
	pm.lastPriceMux.Lock()
	pm.lastPrice = currentPrice
	pm.lastPriceMux.Unlock()

	// Notify callbacks of price update (simplified - only current price)
	pm.notifyPriceUpdateCallbacks(currentPrice)
}

// notifyPriceUpdateCallbacks notifies all registered callbacks
func (pm *PriceMonitor) notifyPriceUpdateCallbacks(current *bitcoin.PriceData) {
	pm.callbackMux.RLock()
	callbacks := make([]PriceUpdateCallback, len(pm.priceUpdateCallbacks))
	copy(callbacks, pm.priceUpdateCallbacks)
	pm.callbackMux.RUnlock()

	for _, callback := range callbacks {
		// Execute callbacks in goroutines to avoid blocking
		go func(cb PriceUpdateCallback) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Price update callback panicked: %v", r)
				}
			}()
			cb(current)
		}(callback)
	}
}

// startPercentageMonitoring starts the percentage change monitoring
func (pm *PriceMonitor) startPercentageMonitoring(ctx context.Context) {
	pm.percentageMonitoringMux.Lock()
	defer pm.percentageMonitoringMux.Unlock()

	if pm.isPercentageMonitoring {
		return
	}

	pm.isPercentageMonitoring = true
	interval := pm.configProvider.GetPercentageUpdateInterval()

	log.Printf("Starting percentage change monitoring every %v", interval)

	go pm.percentageMonitoringLoop(ctx, interval)
}

// percentageMonitoringLoop monitors percentage changes from Binance
func (pm *PriceMonitor) percentageMonitoringLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// First update immediately
	pm.updatePercentageFromPrice()

	for {
		select {
		case <-ticker.C:
			pm.updatePercentageFromPrice()
		case <-pm.percentageStopChannel:
			log.Println("Percentage monitoring stopped")
			return
		case <-ctx.Done():
			log.Println("Percentage monitoring cancelled via context")
			return
		}
	}
}

// updatePercentageFromPrice updates percentage from current price data
func (pm *PriceMonitor) updatePercentageFromPrice() {
	currentPrice, err := pm.priceClient.GetCurrentPrice()
	if err != nil {
		log.Printf("Error fetching price for percentage update: %v", err)
		return
	}

	// Only update if source is Binance (which has percentage data)
	if currentPrice.Source == "Binance" {
		pm.currentPercentageMux.Lock()
		pm.currentPercentage = currentPrice.PriceChangePercent
		pm.currentPercentageMux.Unlock()

		log.Printf("📈 Percentage change updated: %s", currentPrice.FormatPriceChange())
	}
}
