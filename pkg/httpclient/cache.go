package httpclient

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
}

type InMemoryCache struct {
	mu    sync.RWMutex
	store map[string]cacheEntry
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{store: make(map[string]cacheEntry)}
}

func (c *InMemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *InMemoryCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = cacheEntry{
		data:      value,
		expiresAt: time.Now().Add(ttl),
	}
}
