package cache

import (
	"assignment1/models"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx: context.Background(),
	}
}

func (r *RedisCache) Get(key string, expiry time.Duration) (models.Rate, bool) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return models.Rate{}, false
	}
	var rate models.Rate

	err = json.Unmarshal([]byte(val), &rate)

	if err != nil {
		return models.Rate{}, false
	}

	return rate, true
}

func (r *RedisCache) Set(key string, data models.Rate, expiry time.Duration) {
	b, _ := json.Marshal(data)
	r.client.Set(r.ctx, key, b, expiry)
}
