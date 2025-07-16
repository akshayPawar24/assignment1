package service

import (
	"assignment1/cache"
	"assignment1/db"
	"assignment1/models"
	"assignment1/provider"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type RateService struct {
	Provider            provider.RateProvider
	Expiry              time.Duration
	Cache               cache.RateCache
	BackgroundTaskTimer time.Duration
}

func (rs *RateService) GetRate(pair string) (models.Rate, error) {
	val, ok := rs.Cache.Get(pair, rs.Expiry)

	if ok {
		return val, nil
	}

	rates, err := rs.Provider.GetRates()

	if err != nil {
		return models.Rate{}, err
	}

	for key, rate := range rates {
		rs.Cache.Set(key, rate, rs.Expiry)
	}

	go rs.syncToDB(rates)

	return rates[pair], nil
}

func (rs *RateService) SyncToDBAndCache() {
	rates, err := rs.Provider.GetRates()

	if err != nil {
		return
	}

	rs.syncToCache(rates)
	rs.syncToDB(rates)
}

func (rs *RateService) syncToCache(rates map[string]models.Rate) {
	for key, rate := range rates {
		rs.Cache.Set(key, rate, rs.Expiry)
	}
}

func (rs *RateService) syncToDB(rates map[string]models.Rate) {
	for _, rate := range rates {
		//rate.UpdatedAt = time.Now().Unix()
		db.DB.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "base"}, {Name: "target"}},
				DoUpdates: clause.AssignmentColumns([]string{"rate", "updated_at"}),
			},
		).Create(&rate)
	}
}

func (rs *RateService) StartBackgroundSync() {
	go func() {
		ticker := time.NewTicker(rs.BackgroundTaskTimer * time.Minute)
		for range ticker.C {
			log.Print("Auto sync rates to the db and cache")
			rs.SyncToDBAndCache()
		}
	}()
}
