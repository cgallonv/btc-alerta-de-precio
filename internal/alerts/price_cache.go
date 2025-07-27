package alerts

import (
	"sync"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/interfaces"
)

// PriceCache manages a limited cache of recent price data
type PriceCache struct {
	entries []interfaces.PriceCacheEntry
	maxSize int
	mutex   sync.RWMutex
}

// NewPriceCache creates a new price cache with specified maximum size
func NewPriceCache(maxSize int) *PriceCache {
	return &PriceCache{
		entries: make([]interfaces.PriceCacheEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a new price entry to the cache
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

// GetLatest returns the most recent price entry
func (pc *PriceCache) GetLatest() *interfaces.PriceCacheEntry {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if len(pc.entries) == 0 {
		return nil
	}

	return &pc.entries[len(pc.entries)-1]
}

// GetAll returns all cached entries (most recent first)
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

// GetHistory returns the specified number of recent entries (most recent first)
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

// Size returns the current number of cached entries
func (pc *PriceCache) Size() int {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return len(pc.entries)
}

// Clear removes all cached entries
func (pc *PriceCache) Clear() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.entries = pc.entries[:0]
}
