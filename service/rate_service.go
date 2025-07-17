package service

import (
	"assignment1/cache"
	"assignment1/db"
	"assignment1/models"
	"assignment1/provider"
	"fmt"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type RateService struct {
	Provider            provider.RateProvider
	Expiry              time.Duration
	Cache               cache.RateCache
	BackgroundTaskTimer time.Duration
	GlobalBaseCurrency  string
}

func (rs *RateService) GetRate(base string, target string) (models.RateDto, error) {
	var rate models.Rate
	var err error
	rate, found := rs.getRateFromCache(base, target)

	if !found {
		rate, err = rs.getRateFromProvider(base, target)
	}

	if err == nil {
		resp := models.NewRateDto(rate)
		return resp, nil
	} else {
		return models.RateDto{}, err
	}
}

func (rs *RateService) getRateFromCache(base, target string) (models.Rate, bool) {
	pair := base + "_" + target
	rate, found := rs.Cache.Get(pair, rs.Expiry)

	if found {
		return rate, true
	}

	usdToBase, foundBase := rs.Cache.Get(rs.GlobalBaseCurrency+base, rs.Expiry)
	usdToTarget, foundTarget := rs.Cache.Get(rs.GlobalBaseCurrency+target, rs.Expiry)

	if foundBase && foundTarget {
		crossRate := usdToTarget.Rate / usdToBase.Rate
		updatedAt := usdToTarget.UpdatedAt
		if usdToBase.UpdatedAt > updatedAt {
			updatedAt = usdToBase.UpdatedAt
		}
		rate := models.Rate{
			Base:      base,
			Target:    target,
			Rate:      crossRate,
			UpdatedAt: updatedAt,
		}
		rs.Cache.Set(pair, rate, rs.Expiry)
		return rate, true
	}

	return models.Rate{}, false
}

func (rs *RateService) getRateFromProvider(base, target string) (models.Rate, error) {
	rates, err := rs.Provider.GetRates()
	if err != nil {
		return models.Rate{}, err
	}

	// Consider batching or debouncing these syncs
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in syncToDB: %v", r)
			}
		}()
		rs.syncToDB(rates)
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in syncToCache: %v", r)
			}
		}()
		rs.syncToCache(rates)
	}()

	return rs.CalculateCrossRate(rates, base, target)
}

func (rs *RateService) CalculateCrossRate(rates map[string]models.Rate, base, target string) (models.Rate, error) {
	pair := base + "_" + target

	rate, ok := rates[pair]

	if ok {
		return rate, nil
	}
	usdToBaseKey := rs.GlobalBaseCurrency + base
	usdToTargetKey := rs.GlobalBaseCurrency + target
	usdToBase, foundUsdToBase := rates[usdToBaseKey]
	usdToTarget, foundUsdToTarget := rates[usdToTargetKey]

	if !foundUsdToBase || !foundUsdToTarget {
		return models.Rate{}, fmt.Errorf("rate not found for %s_%s", base, target)
	}

	crossRate := usdToTarget.Rate / usdToBase.Rate
	updatedAt := usdToTarget.UpdatedAt

	if usdToBase.UpdatedAt > updatedAt {
		updatedAt = usdToBase.UpdatedAt
	}

	return models.Rate{
		Base:      base,
		Target:    target,
		Rate:      crossRate,
		UpdatedAt: updatedAt,
	}, nil
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
