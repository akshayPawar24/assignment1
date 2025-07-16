package models

type Rate struct {
	ID        uint   `gorm:"primaryKey"`
	Base      string `gorm:"index:idx_base_target,unique"`
	Target    string `gorm:"index:idx_base_target,unique"`
	Rate      float64
	UpdatedAt int64
}
