package cache

import (
	"assignment1/models"
	"time"
)

type RateCache interface {
	Get(key string, expiry time.Duration) (models.Rate, bool)
	Set(key string, data models.Rate, expiry time.Duration)
}

/*type CachedPrice struct {
	Data      models.Rate
	Timestamp int64
}

var (
	cache = make(map[string]CachedPrice)
	mutex sync.RWMutex
)

func Get(currency string, expiry time.Duration) (models.Rate, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	item, found := cache[currency]

	if !found || time.Now().Unix()-item.Timestamp > int64(expiry.Seconds()) {
		return models.Rate{}, false
	}

	return item.Data, true
}

func Set(currency string, data models.Rate) {
	mutex.Lock()

	defer mutex.Unlock()

	cache[currency] = CachedPrice{
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}*/
