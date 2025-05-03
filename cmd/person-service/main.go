package main

import (
	"context"
	"log"
	"person-service/internal/api"
	"person-service/internal/config"
	"person-service/internal/repository/postgres"
	"person-service/internal/service"
	"person-service/pkg/logger"
	pg "person-service/pkg/postgres"

	_ "person-service/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Person Service API
// @version 1.0
// @description API для управления записями о людях с обогащением данных из внешних источников.
// @host localhost:8081
// @BasePath /api/v1
func main() {
	// Инициализация логгера
	logr, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer func() {
		if err := logr.Sync(); err != nil {
			log.Printf("Ошибка синхронизации логов: %v", err)
		}
	}()
	logr.Info("Логгер инициализирован")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logr.Fatal("Ошибка загрузки конфигурации", logger.ErrorKV("error", err))
	}
	logr.Info("Конфигурация загружена", logger.InfoKV("config", cfg))

	// Подключение к Postgres
	ctx := context.Background()
	db, err := pg.NewPostgres(ctx, &cfg.Database)
	if err != nil {
		logr.Fatal("Ошибка подключения к Postgres", logger.ErrorKV("error", err))
	}
	defer db.Close()
	logr.Info("Подключение к Postgres установлено")

	// Инициализация репозитория
	repo := postgres.NewPersonRepository(db.Pool)

	// Инициализация сервиса
	svc := service.NewPersonService(repo, logr.Logger, &cfg.APIs)

	// Инициализация HTTP-сервера
	server := api.NewServer(cfg, svc, logr.Logger)

	// Добавление Swagger UI
	server.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"), // URL для JSON-документации
	))

	logr.Info("HTTP-сервер инициализирован")

	// Запуск сервера
	if err := server.Start(); err != nil {
		logr.Fatal("Ошибка запуска сервера", logger.ErrorKV("error", err))
	}
}
