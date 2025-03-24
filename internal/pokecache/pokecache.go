package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	CacheEntries map[string]cacheEntry
	mu           sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CacheEntries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

}

func (c *Cache) reapLoop(ticker *time.Ticker, interval time.Duration) {
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, entry := range c.CacheEntries {
				if time.Now().Sub(entry.createdAt) >= interval {
					delete(c.CacheEntries, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.CacheEntries[key]; !ok {
		return nil, false
	} else {
		return entry.val, true
	}
}

func NewCache(interval time.Duration) *Cache {
	ticker := time.NewTicker(interval)
	c := &Cache{
		CacheEntries: make(map[string]cacheEntry),
	}
	go c.reapLoop(ticker, interval)
	return c
}
