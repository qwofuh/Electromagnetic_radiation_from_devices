package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"mime/multipart"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/minio/minio-go/v7"
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
func (r *Repository) GetDraftOrder(ctx context.Context, userID int) (*ds.Emission, error) {
	var order ds.Emission
	err := r.db.WithContext(ctx).Where("creator_id = ? AND status = ?", userID, "черновик").First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // черновик отсутствует
		}
		return nil, err
	}
	return &order, nil
}

// Создаём новый черновой заказ
func (r *Repository) CreateDraftOrder(ctx context.Context, userID int) (*ds.Emission, error) {
	order := ds.Emission{
		CreatorID:   userID,
		ModeratorID: nil,
		Status:      "черновик",
	}
	if err := r.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// Добавляем устройсктво в расчет, если его там нет
func (r *Repository) AddDeviceToOrder(ctx context.Context, orderID int, deviceID int) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&ds.DeviceEmissionCalculation{}).Where("emission_id = ? AND device_id = ?", orderID, deviceID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	item := ds.DeviceEmissionCalculation{
		DeviceID:   deviceID,
		EmissionID: orderID,
		CustomPower: sql.NullFloat64{
			Float64: 0,
			Valid:   false,
		},
	}
	return r.db.WithContext(ctx).Create(&item).Error
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

// Получаем одно устройство по ID
func (r *Repository) GetDeviceByID(id int) (*ds.Device, error) {
	var device ds.Device
	err := r.db.First(&device, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// Получаем список устройств с опциональной фильтрацией по названию
func (r *Repository) GetDevicesFiltered(title string) ([]ds.Device, error) {
	var device []ds.Device
	query := r.db.Model(&ds.Device{}).Where("visability = ?", true)
	if title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	err := query.Find(&device).Error
	if err != nil {
		return nil, err
	}
	return device, nil
}

// Создаём новое устройство
func (r *Repository) CreateDevice(device *ds.Device) error {
	return r.db.Create(device).Error
}

// Обновляем устройство по ID
func (r *Repository) UpdateDevice(id int, updated *ds.Device) error {
	var device ds.Device
	if err := r.db.First(&device, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// Обновляем все поля
	return r.db.Model(&device).Updates(updated).Error
}

// Удаление устройства
func (r *Repository) DeleteDevice(id int) error {
	result := r.db.Model(&ds.Device{}).
		Where("id = ?", id).
		Update("visability", false)
	return result.Error
}

// UploadDeviceImage загружает новое изображение устройства в MinIO, удаляет старое, обновляет image_url
func (r *Repository) UploadDeviceImage(id int, fileHeader *multipart.FileHeader) error {
	var device ds.Device
	if err := r.db.First(&device, id).Error; err != nil {
		return err
	}

	// Удаляем старое изображение из MinIO, если есть
	if device.Image != "" {
		parts := strings.Split(device.Image, "/")
		objectName := parts[len(parts)-1]
		_ = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
	}
	// Открываем новый файл
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Расширение файла
	ext := filepath.Ext(fileHeader.Filename)
	base := strings.TrimSuffix(fileHeader.Filename, ext)

	// Переводим в латиницу
	latinBase := toLatin(base)

	// Генерация имени файла
	objectName := fmt.Sprintf("device-%s%s", latinBase, ext)

	// Загружаем в MinIO
	_, err = r.minioClient.PutObject(
		context.Background(),
		r.bucketName,
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")},
	)
	if err != nil {
		return err
	}

	// Формируем URL
	imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioClient.EndpointURL().Host, r.bucketName, objectName)

	// Обновляем поле ImageURL в БД
	return r.db.Model(&ds.Device{}).Where("id = ?", id).Update("image", imageURL).Error
}

// toLatin переводит строку в латиницу, оставляет только ASCII буквы и цифры
func toLatin(s string) string {
	var out strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII {
			out.WriteRune(unicode.ToLower(r))
		} else if unicode.IsDigit(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}
