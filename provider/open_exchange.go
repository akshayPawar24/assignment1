package provider

import (
	"assignment1/models"
	"encoding/json"
	"net/http"
)

type OpenExchangeProvider struct {
	URL     string
	AppId   string
	Adapter ProviderAdapter
}

type apiResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Time  int64              `json:"timestamp"`
}

func (o *OpenExchangeProvider) fetchRawRates() (*apiResponse, error) {
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
	return &data, nil
}

func (o *OpenExchangeProvider) GetRates() (map[string]models.Rate, error) {
	data, err := o.fetchRawRates()
	if err != nil {
		return nil, err
	}
	return o.Adapter.Adapt(data)
}
