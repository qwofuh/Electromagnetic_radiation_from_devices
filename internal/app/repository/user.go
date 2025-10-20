package repository

import (
	"errors"
	"lab1/internal/app/ds"

	"gorm.io/gorm"
)

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(userID int) (*ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByLogin возвращает пользователя по логину
func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser обновляет логин и/или пароль пользователя
func (r *Repository) UpdateUser(userID int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("нет данных для обновления")
	}

	if err := r.db.Model(&ds.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}
