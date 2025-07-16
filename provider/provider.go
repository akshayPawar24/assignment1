package provider

import "assignment1/models"

type RateProvider interface {
	GetRates() (map[string]models.Rate, error)
}
