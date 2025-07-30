// Package alerts provides functionality for monitoring Bitcoin prices
// and managing price-based alerts.
package alerts

import (
	"sync"

	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	"github.com/cgallonv/btc-alerta-de-precio/internal/interfaces"
)

// PriceCache manages a limited cache of recent price data.
// It maintains a fixed-size circular buffer of price entries,
// automatically removing the oldest entries when the cache is full.
//
// Example usage:
//
//	cache := NewPriceCache(20) // Keep last 20 price entries
//	cache.Add(priceData)
//	latest := cache.GetLatest()
//	if latest != nil {
//	    log.Printf("Latest price: $%.2f", latest.Price)
//	}
type PriceCache struct {
	entries []interfaces.PriceCacheEntry
	maxSize int
	mutex   sync.RWMutex
}

// NewPriceCache creates a new price cache with specified maximum size.
// The maxSize parameter determines how many price entries to keep in memory.
//
// Example usage:
//
//	cache := NewPriceCache(24) // Keep 24 hours of price history
//	for _, price := range prices {
//	    cache.Add(price)
//	}
func NewPriceCache(maxSize int) *PriceCache {
	return &PriceCache{
		entries: make([]interfaces.PriceCacheEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a new price entry to the cache.
// If the cache is full, the oldest entry is removed.
//
// Example usage:
//
//	priceData := &bitcoin.PriceData{
//	    Price: 50000,
//	    PriceChangePercent: 2.5,
//	    Currency: "USD",
//	    Source: "Binance",
//	    Timestamp: time.Now(),
//	}
//	cache.Add(priceData)
func (pc *PriceCache) Add(priceData *bitcoin.PriceData) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	entry := interfaces.PriceCacheEntry{
		Price:              priceData.Price,
		PriceChangePercent: priceData.PriceChangePercent,
		Currency:           priceData.Currency,
		Source:             priceData.Source,
		Timestamp:          priceData.Timestamp,
	}

	// Add new entry
	pc.entries = append(pc.entries, entry)

	// Keep only last maxSize entries
	if len(pc.entries) > pc.maxSize {
		pc.entries = pc.entries[len(pc.entries)-pc.maxSize:]
	}
}

// GetLatest returns the most recent price entry.
// Returns nil if the cache is empty.
//
// Example usage:
//
//	if latest := cache.GetLatest(); latest != nil {
//	    log.Printf("Latest price: $%.2f (%+.2f%%)",
//	        latest.Price, latest.PriceChangePercent)
//	}
func (pc *PriceCache) GetLatest() *interfaces.PriceCacheEntry {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if len(pc.entries) == 0 {
		return nil
	}

	return &pc.entries[len(pc.entries)-1]
}

// GetAll returns all cached entries (most recent first).
// Returns an empty slice if the cache is empty.
//
// Example usage:
//
//	entries := cache.GetAll()
//	for _, entry := range entries {
//	    log.Printf("Price at %s: $%.2f",
//	        entry.Timestamp.Format("15:04:05"), entry.Price)
//	}
func (pc *PriceCache) GetAll() []interfaces.PriceCacheEntry {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if len(pc.entries) == 0 {
		return []interfaces.PriceCacheEntry{}
	}

	// Return reversed slice (most recent first)
	result := make([]interfaces.PriceCacheEntry, len(pc.entries))
	for i, entry := range pc.entries {
		result[len(pc.entries)-1-i] = entry
	}

	return result
}

// GetHistory returns the specified number of recent entries (most recent first).
// If limit is greater than the number of entries, returns all entries.
//
// Example usage:
//
//	// Get last hour of price history (assuming 5-minute intervals)
//	history := cache.GetHistory(12)
//	for _, entry := range history {
//	    log.Printf("Price at %s: $%.2f",
//	        entry.Timestamp.Format("15:04:05"), entry.Price)
//	}
func (pc *PriceCache) GetHistory(limit int) []interfaces.PriceCacheEntry {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if len(pc.entries) == 0 {
		return []interfaces.PriceCacheEntry{}
	}

	// Determine how many entries to return
	start := len(pc.entries) - limit
	if start < 0 {
		start = 0
	}

	// Get the entries and reverse them (most recent first)
	selected := pc.entries[start:]
	result := make([]interfaces.PriceCacheEntry, len(selected))
	for i, entry := range selected {
		result[len(selected)-1-i] = entry
	}

	return result
}

// Size returns the current number of entries in the cache.
//
// Example usage:
//
//	size := cache.Size()
//	log.Printf("Cache contains %d price entries", size)
func (pc *PriceCache) Size() int {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return len(pc.entries)
}

// Clear removes all entries from the cache.
//
// Example usage:
//
//	cache.Clear() // Reset cache
//	if cache.Size() == 0 {
//	    log.Printf("Cache cleared successfully")
//	}
func (pc *PriceCache) Clear() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.entries = make([]interfaces.PriceCacheEntry, 0, pc.maxSize)
}
