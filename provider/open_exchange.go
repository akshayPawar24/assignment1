package provider

import (
	"assignment1/models"
	"encoding/json"
	"net/http"
	"time"
)

type OpenExchangeProvider struct {
	URL   string
	AppId string
}

type apiResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Time  int64              `json:"timestamp"`
}

func (o *OpenExchangeProvider) GetRates() (map[string]models.Rate, error) {
	resp, err := http.Get(o.URL + o.AppId)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data apiResponse

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
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
