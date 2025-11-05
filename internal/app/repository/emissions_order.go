package repository

import (
	"context"
	"fmt"
	"lab1/internal/app/ds"
	"math"
	"strings"
	"time"
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

func (r *Repository) GetOrdersFiltered(status, start, end string) ([]ds.EmissionResponse, error) {
	var orders []ds.EmissionResponse

	query := r.db.
		Table("emissions").
		Select(`emissions.id, 
		        emissions.status, 
		        emissions.create_at,
				emissions.update_at, 
		        emissions.finish_at,
				emissions.distance,
				emissions.total_emission, 
		        u1.login as moderator, 
		        u2.login as creator`).
		Joins("LEFT JOIN users u1 ON u1.id = emissions.moderator_id").
		Joins("LEFT JOIN users u2 ON u2.id = emissions.creator_id")

	// фильтр по статусу
	if status != "" {
		statuses := []string{}
		for _, s := range strings.Split(status, ",") {
			s = strings.TrimSpace(s)
			// Проверяем, что не пытаются фильтровать по черновику или удалённому
			if s == "черновик" || s == "удален" {
				return nil, fmt.Errorf("not_found")
			}
			statuses = append(statuses, s)
		}
		query = query.Where("emissions.status IN ?", statuses)
	}

	// фильтр по диапазону дат
	if start != "" && end != "" {
		query = query.Where("emissions.create_at BETWEEN ? AND ?", start, end)
	}

	// исключаем черновик и удалённые
	query = query.Where("emissions.status NOT IN ?", []string{"черновик", "удален"})

	if err := query.Scan(&orders).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении расчетов: %w", err)
	}

	return orders, nil
}

func (r *Repository) UpdateDeviceOrder(orderID int, req ds.EmissionUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Distance != nil {
		updates["distance"] = *req.Distance
	}

	if len(updates) == 0 {
		return nil // ничего менять не нужно
	}

	return r.db.Model(&ds.Emission{}).Where("id = ?", orderID).Updates(updates).Error
}

func (r *Repository) FormDeviceOrder(orderID int) error {
	// Проверяем, что все custom_power заполнены
	var count int64
	if err := r.db.Model(&ds.DeviceEmissionCalculation{}).
		Where("emission_id = ? AND custom_power IS NULL", orderID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("ошибка проверки custom_power: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("нельзя сформировать заказ: не все custom_power заполнены")
	}

	// Обновляем заказ: статус и update_at
	updates := map[string]interface{}{
		"status":    "сформирован",
		"update_at": time.Now(),
	}

	// Обновляем только если текущий статус черновик
	if err := r.db.Model(&ds.Emission{}).
		Where("id = ? AND status = ?", orderID, "черновик").
		Updates(updates).Error; err != nil {
		return fmt.Errorf("ошибка обновления расчета: %w", err)
	}

	return nil
}

func (r *Repository) CompleteOrRejectOrder(orderID int, req ds.EmissionCompleteRequest) error {
	var order ds.Emission
	if err := r.db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("расчет с ID=%d не найден", orderID)
	}

	// Обновляем статус, модератора и дату завершения
	updates := map[string]interface{}{
		"status":       req.Status,
		"moderator_id": req.ModeratorID,
		"finish_at":    time.Now(),
	}

	if err := r.db.Model(&ds.Emission{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
		return fmt.Errorf("не удалось обновить расчет: %w", err)
	}

	// Если заказ отклонён — прекращаем выполнение, не считаем
	if req.Status == "отклонен" {
		return nil
	}

	var dec []ds.DeviceEmissionCalculation
	if err := r.db.Preload("Device").Where("emission_id = ?", orderID).Find(&dec).Error; err != nil {
		return err
	}

	var totalEmission float64

	for _, deviceCalc := range dec {
		if order.Distance > 0 && deviceCalc.CustomPower.Valid {

			currentEmission := deviceCalc.CustomPower.Float64 / (4 * math.Pi * order.Distance * order.Distance)

			totalEmission += currentEmission
		}
	}

	// Сохраняем ОБЩУЮ сумму в основной расчет Emission
	if err := r.db.Model(&ds.Emission{}).
		Where("id = ?", orderID).
		Update("total_emission", totalEmission).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) SoftDeleteOrder(orderID int) error {
	updates := map[string]interface{}{
		"status":    "удален",
		"update_at": time.Now(), // дата завершения
	}

	return r.db.Model(&ds.Emission{}).Where("id = ?", orderID).Updates(updates).Error
}

// GetOrdersFilteredForUser возвращает заказы указанного пользователя (creator_id = userID)
func (r *Repository) GetOrdersFilteredForUser(ctx context.Context, status, start, end string, userID int) ([]ds.EmissionResponse, error) {
	var orders []ds.EmissionResponse

	// Разрешённые статусы для выдачи
	allowedStatuses := map[string]bool{
		"сформирован": true,
		"завершен":    true,
		"отклонен":    true,
	}

	query := r.db.WithContext(ctx).
		Table("emissions").
		Select(`emissions.id, 
				emissions.status, 
				emissions.create_at, 
				emissions.update_at, 
				emissions.finish_at, 
				u1.login as moderator, 
				u2.login as creator`).
		Joins("LEFT JOIN users u1 ON u1.id = mo.moderator_id").
		Joins("LEFT JOIN users u2 ON u2.id = mo.creator_id").
		Where("emissions.creator_id = ?", userID)

	// фильтр по статусу
	if status != "" {
		statuses := []string{}
		for _, s := range strings.Split(status, ",") {
			s = strings.TrimSpace(s)
			if allowedStatuses[s] { // оставляем только разрешённые
				statuses = append(statuses, s)
			}
		}
		if len(statuses) == 0 {
			return []ds.EmissionResponse{}, nil
		}
		query = query.Where("emissions.status IN ?", statuses)
	} else {
		// Если статус не указан — выдаём все разрешённые статусы
		query = query.Where("emissions.status IN ?", []string{"сформирован", "завершен", "отклонен"})
	}

	// фильтр по диапазону дат
	if start != "" && end != "" {
		query = query.Where("emissions.create_at BETWEEN ? AND ?", start, end)
	}

	if err := query.Scan(&orders).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении заказов: %w", err)
	}

	return orders, nil
}
