package provider

import (
	"assignment1/models"
	"errors"
	"time"
)

type OpenExchangeAdapter struct{}

func (a *OpenExchangeAdapter) Adapt(rawData interface{}) (map[string]models.Rate, error) {
	data, ok := rawData.(*apiResponse)
	if !ok {
		return nil, errors.New("invalid data type for OpenExchangeAdapter")
	}
	updateTime := time.Now().Unix()
	rates := make(map[string]models.Rate)
	for code, val := range data.Rates {
		key := data.Base + "_" + code
		rates[key] = models.Rate{
			Rate:      val,
			Base:      data.Base,
			Target:    code,
			UpdatedAt: updateTime,
		}
	}
	return rates, nil
}
