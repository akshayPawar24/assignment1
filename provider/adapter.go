package provider

import "assignment1/models"

type ProviderAdapter interface {
	Adapt(rawData interface{}) (map[string]models.Rate, error)
}
