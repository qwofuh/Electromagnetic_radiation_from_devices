package repository

import (
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"strings"

	"gorm.io/gorm"
)

func (r *Repository) GetDevices() ([]ds.Device, error) {
	var devices []ds.Device

	result := r.db.Find(&devices)

	if result.Error != nil {
		fmt.Printf("ERROR: %v\n", result.Error)
		return nil, result.Error
	}
	return devices, nil
}

func (r *Repository) GetDeviceByTitle(title string) ([]ds.Device, error) {
	var devices []ds.Device
	result := r.db.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%").Find(&devices)
	if result.Error != nil {
		return nil, result.Error
	}
	return devices, nil
}

// Получаем черновой заказ пользователя
func (r *Repository) GetDraftOrder(userID int) (*ds.Emission, error) {
	var order ds.Emission
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // черновик отсутствует
		}
		return nil, err
	}
	return &order, nil
}

// Создаём новый черновой заказ
func (r *Repository) CreateDraftOrder(userID int) (*ds.Emission, error) {
	order := ds.Emission{
		CreatorID:   userID,
		ModeratorID: 1,
		Status:      "черновик",
	}
	if err := r.db.Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// Добавляем устройсктво в расчет, если его там нет
func (r *Repository) AddDeviceToOrder(orderID int, deviceID int) error {
	var count int64
	err := r.db.Model(&ds.DeviceEmissionCalculation{}).Where("emission_id = ? AND device_id = ?", orderID, deviceID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	item := ds.DeviceEmissionCalculation{
		DeviceID:   deviceID,
		EmissionID: orderID,
	}
	return r.db.Create(&item).Error
}

// Получаем количество устройств в заказе
func (r *Repository) GetOrderDevicesCount(orderID int) (int64, error) {
	var count int64
	err := r.db.Model(&ds.DeviceEmissionCalculation{}).Where("emission_id = ?", orderID).Count(&count).Error
	return count, err
}

// обновляем статус заказа по его ID
func (r *Repository) SetOrderStatus(orderID int, status string) error {
	return r.db.Model(&ds.Emission{}).Where("id = ?", orderID).Update("status", status).Error
}
