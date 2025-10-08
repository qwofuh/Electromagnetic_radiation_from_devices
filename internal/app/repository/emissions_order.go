package repository

import (
	"fmt"
	"lab1/internal/app/ds"
)

func (r *Repository) GetOrderByID(id int) (ds.Emission, []ds.DeviceEmissionCalculation, error) {
	var order ds.Emission
	if err := r.db.First(&order, id).Error; err != nil {
		return ds.Emission{}, nil, fmt.Errorf("расчет с ID=%d не найден", id)
	}

	var dec []ds.DeviceEmissionCalculation
	if err := r.db.Preload("Device").Where("emission_id = ?", id).Find(&dec).Error; err != nil {
		return ds.Emission{}, nil, err
	}

	return order, dec, nil
}
