package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	CacheExpiry         int
	ExchangeURL         string
	ExchangeAppId       string
	DBUrl               string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	BackgroundTaskTimer time.Duration
	GlobalBaseCurrency  string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	taskTimer, _ := strconv.Atoi(os.Getenv("BACKGROUND_TASK_TIMER"))

	cacheExpiry, _ := strconv.Atoi(os.Getenv("CACHE_EXPIRY_SECONDS"))

	var config = &Config{
		Port:                os.Getenv("PORT"),
		ExchangeURL:         os.Getenv("OPENEXCHANGE_URL"),
		ExchangeAppId:       os.Getenv("OPENEXCHANGE_APP_ID"),
		CacheExpiry:         cacheExpiry,
		DBUrl:               os.Getenv("DATABASE_URL"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             redisDB,
		BackgroundTaskTimer: time.Duration(taskTimer),
		GlobalBaseCurrency:  os.Getenv("GLOBAL_BASE_CURRENCY"),
	}

	return config
}
