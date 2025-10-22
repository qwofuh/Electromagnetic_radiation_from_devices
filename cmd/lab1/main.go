package main

import (
	"context"
	"fmt"
	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	redisclient "lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Device emission Service API
// @version 1.0
// @description API сервиса аутентификации с Redis и JWT. Поддерживает роли пользователей и администраторов.
// @host localhost:8080
// @BasePath /
// @schemes http
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	router := gin.Default()

	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получаем строку подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	fmt.Println("Postgres DSN:", postgresString)

	// Инициализируем репозиторий (Postgres + MinIO)
	rep, err := repository.New(
		postgresString,
		conf.Minio.Endpoint,
		conf.Minio.AccessKey,
		conf.Minio.SecretKey,
		conf.Minio.Bucket,
	)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Создаём хендлер с репозиторием и конфигом
	// Создаём redis client
	redisCli, err := redisclient.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(rep, conf, redisCli)

	// Инициализируем приложение и запускаем сервер
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
