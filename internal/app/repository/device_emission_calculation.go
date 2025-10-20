package repository

import (
	"lab1/internal/app/ds"
)

func (r *Repository) DeleteDeviceFromOrder(deviceID, orderID int) error {
	return r.db.
		Where("device_id = ? AND emission_id = ?", deviceID, orderID).
		Delete(&ds.DeviceEmissionCalculation{}).Error
}

func (r *Repository) UpdateCustomPower(deviceID, orderID int, customPower float64) error {
	return r.db.
		Model(&ds.DeviceEmissionCalculation{}).
		Where("device_id = ? AND emission_id = ?", deviceID, orderID).
		Update("custom_power", customPower).Error
}
