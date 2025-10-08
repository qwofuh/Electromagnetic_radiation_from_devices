package repository

import (
	"lab1/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Включаем логи SQL
	})
	if err != nil {
		return nil, err
	}

	// Авто-миграция моделей
	err = db.AutoMigrate(&ds.Device{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}
