package ds

// представляет связь "устройства ↔ расчет" (многие ко многим)
type DeviceEmissionCalculation struct {
	EmissionID  int     `gorm:"not null"`
	DeviceID    int     `gorm:"not null"`
	CustomPower float64 `gorm:"type:numeric(10,2)"`

	Device   Device   `gorm:"foreignKey:DeviceID;references:ID"`
	Emission Emission `gorm:"foreignKey:EmissionID;references:ID"`
}
