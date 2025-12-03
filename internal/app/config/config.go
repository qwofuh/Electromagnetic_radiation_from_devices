package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// userSingleton — структура с фиксированными ID
type userSingleton struct {
	CreatorID   int
	ModeratorID int
}

// Глобальная переменная — единственный экземпляр
var user = &userSingleton{
	CreatorID:   1,
	ModeratorID: 2,
}

// GetUserSingleton возвращает "единственный" экземпляр пользователя
func GetUserSingleton() *userSingleton {
	return user
}

// MinioConfig хранит настройки для MinIO
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type JWTConfig struct {
	AccessSecret    string
	RefreshSecret   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

// Config хранит настройки сервиса
type Config struct {
	ServiceHost string
	ServicePort int
	Minio       MinioConfig
	Redis       RedisConfig
	JWT         JWTConfig
}

// NewConfig загружает и возвращает конфигурацию приложения
func NewConfig() (*Config, error) {
	// Подгрузка переменных окружения из .env
	_ = godotenv.Load()

	configName := os.Getenv("CONFIG_NAME")
	if configName == "" {
		configName = "config"
	}

	// Настройка Viper
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	// Чтение конфигурации
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Преобразование данных в структуру Config
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	log.Info("Config successfully parsed")

	// Дополняем конфиг значениями из .env
	cfg.Minio = MinioConfig{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
	}

	cfg.JWT = JWTConfig{
		AccessSecret:   os.Getenv("JWT_ACCESS_SECRET"),
		AccessTokenTTL: 15 * time.Minute,
	}

	cfg.Redis = RedisConfig{
		Host:        viper.GetString("redis.host"),
		Port:        viper.GetInt("redis.port"),
		User:        viper.GetString("redis.user"),
		Password:    os.Getenv("REDIS_PASSWORD"),
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
	}

	return cfg, nil
}
