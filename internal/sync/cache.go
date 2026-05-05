package sync

import (
	"sync"
	"time"
)

// CacheEntry holds a cached secret value with its expiry time.
type CacheEntry struct {
	Value     string
	ExpiresAt time.Time
}

// SecretCache provides a simple TTL-based in-memory cache for secret values,
// reducing redundant reads from Vault during a sync run.
type SecretCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}

// NewSecretCache creates a SecretCache with the given TTL. A zero TTL disables
// caching (every lookup is a miss).
func NewSecretCache(ttl time.Duration) *SecretCache {
	return &SecretCache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

// Get returns the cached value for key and true if a valid (non-expired) entry
// exists, otherwise returns empty string and false.
func (c *SecretCache) Get(key string) (string, bool) {
	if c.ttl == 0 {
		return "", false
	}
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return "", false
	}
	return entry.Value, true
}

// Set stores a value in the cache under key.
func (c *SecretCache) Set(key, value string) {
	if c.ttl == 0 {
		return
	}
	c.mu.Lock()
	c.entries[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Invalidate removes a single key from the cache.
func (c *SecretCache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Flush removes all entries from the cache.
func (c *SecretCache) Flush() {
	c.mu.Lock()
	c.entries = make(map[string]CacheEntry)
	c.mu.Unlock()
}

// Size returns the number of entries currently held (including possibly
// expired ones not yet evicted).
func (c *SecretCache) Size() int {
	c.mu.RLock()
	n := len(c.entries)
	c.mu.RUnlock()
	return n
}
