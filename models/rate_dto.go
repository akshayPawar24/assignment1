package models

import "assignment1/utility"

type RateDto struct {
	Base      string
	Target    string
	Rate      float64
	UpdatedAt int64
}

func NewRateDto(rate Rate) RateDto {
	return RateDto{
		Base:      rate.Base,
		Target:    rate.Target,
		Rate:      utility.Round(rate.Rate, 2),
		UpdatedAt: rate.UpdatedAt,
	}
}
