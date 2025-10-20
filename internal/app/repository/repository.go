package repository

import (
	"lab1/internal/app/ds"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db          *gorm.DB
	minioClient *minio.Client
	bucketName  string
}

func New(dsn, minioEndpoint, minioAccessKey, minioSecretKey, bucketName string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Авто-миграция моделей
	err = db.AutoMigrate(
		&ds.Device{},
		&ds.Emission{},
		&ds.DeviceEmissionCalculation{},
	)
	if err != nil {
		return nil, err
	}

	// Создаём MinIO клиента
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false, // true, если HTTPS
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}
