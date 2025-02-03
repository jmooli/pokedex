package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entry map[string]cacheEntry
	mu    sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		entry: make(map[string]cacheEntry),
	}

	go cache.reaploop(interval)
	return &cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entry == nil {
		c.entry = make(map[string]cacheEntry)
	}

	if _, exists := c.entry[key]; exists {
		// propably not needed but error if key exists
	}

	c.entry[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.entry[key]

	if !exists {
		return nil, false
	}

	return c.entry[key].val, true
}

func (c *Cache) reaploop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.Clean(interval)
	}
}

func (c *Cache) Clean(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entry {
		if now.Sub(entry.createdAt) > interval {
			delete(c.entry, key)
		}
	}
}
