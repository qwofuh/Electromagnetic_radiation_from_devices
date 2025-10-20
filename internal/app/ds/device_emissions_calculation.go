package ds

import "database/sql"

// представляет связь "устройства ↔ расчет" (многие ко многим)
type DeviceEmissionCalculation struct {
	EmissionID  int             `gorm:"not null" json:"emission_id"`
	DeviceID    int             `gorm:"not null" json:"device_id"`
	CustomPower sql.NullFloat64 `gorm:"type:numeric(10,2)" json:"custom_power"`

	Device   *Device   `gorm:"foreignKey:DeviceID;references:ID" json:"device,omitempty"`
	Emission *Emission `gorm:"foreignKey:EmissionID;references:ID" json:"emission,omitempty"`
}
