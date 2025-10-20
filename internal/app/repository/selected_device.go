package repository

import "lab1/internal/app/ds"

func (r *Repository) GetDevice(id int) (ds.Device, error) {
	var device ds.Device
	result := r.db.First(&device, id)
	if result.Error != nil {
		return ds.Device{}, result.Error
	}
	return device, nil
}
