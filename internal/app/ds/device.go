package ds

type Device struct {
	ID               int     `gorm:"primaryKey"`
	Title            string  `gorm:"type:varchar(50);not null"`
	AvgMinPower      float64 `gorm:"column:minavgpower;type:numeric(10,2)"`
	AvgMaxPower      float64 `gorm:"column:maxavgpower;type:numeric(10,2)"`
	MinSafeRange     float64 `gorm:"column:minsaferange;type:numeric(10,2)"`
	MaxSafeRange     float64 `gorm:"column:maxsaferange;type:numeric(10,2)"`
	RadiationType    string  `gorm:"column:radiationtype;type:varchar(100)"`
	RadiationSource  string  `gorm:"column:radiationsource;type:varchar(100)"`
	MaxRadiationZone string  `gorm:"column:maxradiationzone;type:varchar(100)"`
	Image            string  `gorm:"column:image;type:varchar(200)"`
}

func (Device) TableName() string {
	return "devices"
}
