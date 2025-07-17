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
	log.Printf("API Call: Getting rate for %s to %s", base, target)

	var rate models.Rate
	var err error

	// Step 1: Try to get from cache
	log.Printf("Step 1: Checking cache for %s_%s", base, target)
	rate, found := rs.getRateFromCache(base, target)

	if found {
		log.Printf("Cache HIT: Found rate %s_%s = %.6f", base, target, rate.Rate)
		resp := models.NewRateDto(rate)
		return resp, nil
	}

	// Step 2: Cache miss, try database
	log.Printf("Cache MISS: Checking database for %s_%s", base, target)
	rate, err = rs.getRateFromDB(base, target)

	if err == nil {
		log.Printf("Database HIT: Found rate %s_%s = %.6f", base, target, rate.Rate)
		// Cache the result for future requests
		pair := base + "_" + target
		rs.Cache.Set(pair, rate, rs.Expiry)
		log.Printf("Cached rate %s_%s for future requests", base, target)

		resp := models.NewRateDto(rate)
		return resp, nil
	} else {
		log.Printf("Database MISS: Rate not found for %s_%s, error: %v", base, target, err)
		return models.RateDto{}, err
	}
}

func (rs *RateService) getRateFromDB(base, target string) (models.Rate, error) {
	log.Printf("DB Query: Attempting to fetch rate from database for %s_%s", base, target)

	if base == rs.GlobalBaseCurrency {
		// Only look in for direct pair if base is the global currency
		var rate models.Rate
		result := db.DB.Where("base = ? AND target = ?", base, target).First(&rate)
		if result.Error != nil {
			log.Printf("DB Error: Direct pair not found for %s_%s, error: %v", base, target, result.Error)
			return models.Rate{}, fmt.Errorf("provided currency %s is currently not supported", target)
		}
		log.Printf("DB Success: Found direct pair %s_%s = %.6f", base, target, rate.Rate)
		return rate, nil
	}

	// Otherwise, calculate cross rate using USD as intermediary
	log.Printf("DB Query: Calculating cross rate for %s_%s using %s as intermediary", base, target, rs.GlobalBaseCurrency)
	var usdToBase, usdToTarget models.Rate

	resBase := db.DB.Where("base = ? AND target = ?", rs.GlobalBaseCurrency, base).First(&usdToBase)
	if resBase.Error != nil {
		log.Printf("DB Error: Could not find %s_%s in database, error: %v", rs.GlobalBaseCurrency, base, resBase.Error)
		return models.Rate{}, fmt.Errorf("provided currency %s is currently not supported", base)
	}

	resTarget := db.DB.Where("base = ? AND target = ?", rs.GlobalBaseCurrency, target).First(&usdToTarget)
	if resTarget.Error != nil {
		log.Printf("DB Error: Could not find %s_%s in database, error: %v", rs.GlobalBaseCurrency, target, resTarget.Error)
		return models.Rate{}, fmt.Errorf("provided currency %s is currently not supported", target)
	}

	log.Printf("DB Success: Found %s_%s = %.6f and %s_%s = %.6f",
		rs.GlobalBaseCurrency, base, usdToBase.Rate,
		rs.GlobalBaseCurrency, target, usdToTarget.Rate)

	rate := rs.calculateCrossRateFromRates(usdToBase, usdToTarget, base, target)
	log.Printf("DB Success: Calculated cross rate %s_%s = %.6f", base, target, rate.Rate)
	return rate, nil
}

func (rs *RateService) getRateFromCache(base, target string) (models.Rate, bool) {
	pair := base + "_" + target

	log.Printf("Cache Lookup: Checking for key '%s'", pair)
	rate, found := rs.Cache.Get(pair, rs.Expiry)
	if found {
		log.Printf("Cache HIT: Found %s = %.6f", pair, rate.Rate)
		return rate, true
	}

	log.Printf("Cache MISS: Direct pair %s not found, trying cross-rate calculation", pair)
	usdToBase, foundBase := rs.Cache.Get(rs.GlobalBaseCurrency+"_"+base, rs.Expiry)
	usdToTarget, foundTarget := rs.Cache.Get(rs.GlobalBaseCurrency+"_"+target, rs.Expiry)

	if foundBase && foundTarget {
		log.Printf("Cache Cross-Rate: Found %s_%s = %.6f and %s_%s = %.6f",
			rs.GlobalBaseCurrency, base, usdToBase.Rate,
			rs.GlobalBaseCurrency, target, usdToTarget.Rate)

		rate := rs.calculateCrossRateFromRates(usdToBase, usdToTarget, base, target)
		log.Printf("Cache Cross-Rate: Calculated and cached %s = %.6f", pair, rate.Rate)
		return rate, true
	}

	log.Printf("Cache MISS: Cross-rate calculation failed for %s_%s", base, target)
	return models.Rate{}, false
}

func (rs *RateService) calculateCrossRateFromRates(usdToBase, usdToTarget models.Rate, baseCode, targetCode string) models.Rate {
	crossRate := usdToTarget.Rate / usdToBase.Rate
	updatedAt := usdToTarget.UpdatedAt
	if usdToBase.UpdatedAt > updatedAt {
		updatedAt = usdToBase.UpdatedAt
	}

	pair := baseCode + "_" + targetCode
	rate := models.Rate{
		Base:      baseCode,
		Target:    targetCode,
		Rate:      crossRate,
		UpdatedAt: updatedAt,
	}

	rs.Cache.Set(pair, rate, rs.Expiry)

	return rate
}

// Use this method if you want to get the rates on demand from the API
func (rs *RateService) getRateFromProvider(base, target string) (models.Rate, error) {
	log.Printf("Provider Call: Fetching rates from external API for %s_%s", base, target)
	rates, err := rs.Provider.GetRates()
	if err != nil {
		log.Printf("Provider Error: Failed to fetch rates from API, error: %v", err)
		return models.Rate{}, err
	}
	log.Printf("Provider Success: Fetched %d rates from external API", len(rates))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in syncToDB: %v", r)
			}
		}()
		log.Printf("Background Task: Starting database sync for %d rates", len(rates))
		rs.syncToDB(rates)
		log.Printf("Background Task: Completed database sync")
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in syncToCache: %v", r)
			}
		}()
		log.Printf("Background Task: Starting cache sync for %d rates", len(rates))
		rs.syncToCache(rates)
		log.Printf("Background Task: Completed cache sync")
	}()

	return rs.calculateCrossRate(rates, base, target)
}

func (rs *RateService) calculateCrossRate(rates map[string]models.Rate, base, target string) (models.Rate, error) {
	pair := base + "_" + target
	log.Printf("Cross-Rate Calculation: Attempting to calculate %s from provider data", pair)

	rate, ok := rates[pair]

	if ok {
		log.Printf("Cross-Rate Success: Direct pair found %s = %.6f", pair, rate.Rate)
		return rate, nil
	}

	log.Printf("Cross-Rate Calculation: Direct pair not found, calculating cross-rate")
	usdToBaseKey := rs.GlobalBaseCurrency + "_" + base
	usdToTargetKey := rs.GlobalBaseCurrency + "_" + target
	usdToBase, foundUsdToBase := rates[usdToBaseKey]
	usdToTarget, foundUsdToTarget := rates[usdToTargetKey]

	if !foundUsdToBase || !foundUsdToTarget {
		log.Printf("Cross-Rate Error: Missing rates for %s or %s", usdToBaseKey, usdToTargetKey)
		return models.Rate{}, fmt.Errorf("rate not found for %s_%s", base, target)
	}

	log.Printf("Cross-Rate Calculation: Found %s = %.6f and %s = %.6f",
		usdToBaseKey, usdToBase.Rate, usdToTargetKey, usdToTarget.Rate)

	rate = rs.calculateCrossRateFromRates(usdToBase, usdToTarget, base, target)
	log.Printf("Cross-Rate Success: Calculated %s = %.6f", pair, rate.Rate)

	return rate, nil
}

func (rs *RateService) syncToDBAndCache() {
	log.Printf("Sync Task: Starting sync to DB and cache")
	rates, err := rs.Provider.GetRates()

	if err != nil {
		log.Printf("Sync Error: Failed to fetch rates from provider, error: %v", err)
		return
	}

	log.Printf("Sync Task: Successfully fetched %d rates, syncing to cache and DB", len(rates))
	rs.syncToCache(rates)
	rs.syncToDB(rates)
	log.Printf("Sync Task: Completed sync to DB and cache")
}

func (rs *RateService) syncToCache(rates map[string]models.Rate) {
	log.Printf("Cache Sync: Syncing %d rates to cache", len(rates))
	for key, rate := range rates {
		rs.Cache.Set(key, rate, rs.Expiry)
	}
	log.Printf("Cache Sync: Successfully synced %d rates to cache", len(rates))
}

func (rs *RateService) syncToDB(rates map[string]models.Rate) {
	log.Printf("DB Sync: Syncing %d rates to database", len(rates))
	for _, rate := range rates {
		//rate.UpdatedAt = time.Now().Unix()
		db.DB.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "base"}, {Name: "target"}},
				DoUpdates: clause.AssignmentColumns([]string{"rate", "updated_at"}),
			},
		).Create(&rate)
	}
	log.Printf("DB Sync: Successfully synced %d rates to database", len(rates))
}

func (rs *RateService) StartBackgroundSync() {
	log.Printf("Background Sync: Starting background sync task with interval %d minutes", rs.BackgroundTaskTimer)
	go func() {
		ticker := time.NewTicker(rs.BackgroundTaskTimer * time.Minute)
		for range ticker.C {
			log.Print("Auto sync rates to the db and cache")
			rs.syncToDBAndCache()
		}
	}()
}
