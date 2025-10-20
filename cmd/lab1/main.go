package main

import (
	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"lab1/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// main инициализирует конфигурацию, репозиторий, хендлеры и запускает приложение
func main() {
	router := gin.Default() // создание нового роутера Gin

	// Загрузка конфигурации приложения
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получение строки подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	logrus.Infof("DSN: %s", postgresString) // ← подробное логирование

	// Инициализация репозитория
	rep, errRep := repository.New(
		postgresString,
		conf.Minio.Endpoint,
		conf.Minio.AccessKey,
		conf.Minio.SecretKey,
		conf.Minio.Bucket,
	)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Создание хендлера с подключённым репозиторием
	hand := handler.NewHandler(rep)

	// Инициализация приложения и запуск сервера
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
