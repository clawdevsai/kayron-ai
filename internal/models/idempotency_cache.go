package models

import (
	"sync"
	"time"
)

// IdempotencyCache stores idempotency keys with 24h TTL to prevent duplicate order fills
type IdempotencyCache struct {
	mu    sync.RWMutex
	cache map[string]*IdempotencyCacheEntry
	ttl   time.Duration
}

// IdempotencyCacheEntry represents a cached idempotency key
type IdempotencyCacheEntry struct {
	Ticket    int64
	ExpiresAt time.Time
}

// NewIdempotencyCache creates a new IdempotencyCache with 24h TTL
func NewIdempotencyCache() *IdempotencyCache {
	cache := &IdempotencyCache{
		cache: make(map[string]*IdempotencyCacheEntry),
		ttl:   24 * time.Hour,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a cached idempotency key
func (ic *IdempotencyCache) Get(key string) (int64, bool) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	entry, exists := ic.cache[key]
	if !exists {
		return 0, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return 0, false
	}

	return entry.Ticket, true
}

// Set stores a ticket for an idempotency key
func (ic *IdempotencyCache) Set(key string, ticket int64) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.cache[key] = &IdempotencyCacheEntry{
		Ticket:    ticket,
		ExpiresAt: time.Now().Add(ic.ttl),
	}
}

// cleanup removes expired entries every 1 hour
func (ic *IdempotencyCache) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ic.mu.Lock()
		now := time.Now()
		for key, entry := range ic.cache {
			if now.After(entry.ExpiresAt) {
				delete(ic.cache, key)
			}
		}
		ic.mu.Unlock()
	}
}
