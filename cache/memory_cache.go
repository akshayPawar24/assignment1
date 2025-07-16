package cache

import (
	"assignment1/models"
	"sync"
	"time"
)

type InMemoryCache struct {
	cache map[string]CachedPrice
	mutex sync.RWMutex
}

type CachedPrice struct {
	Data      models.Rate
	Timestamp int64 // epoch
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]CachedPrice),
	}
}

func (c *InMemoryCache) Get(key string, expiry time.Duration) (models.Rate, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, found := c.cache[key]
	if !found || time.Now().Unix()-item.Timestamp > int64(expiry.Seconds()) {
		return models.Rate{}, false
	}
	return item.Data, true
}

func (c *InMemoryCache) Set(key string, data models.Rate, expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[key] = CachedPrice{
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}
