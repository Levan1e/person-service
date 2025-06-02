package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"person-service/internal/api"
	"person-service/internal/config"
	"person-service/internal/repository/postgres"
	"person-service/internal/service"
	"person-service/pkg/logger"
	pg "person-service/pkg/postgres"
	"syscall"
	"time"

	_ "person-service/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Person Service API
// @version 1.0
// @description API для управления записями о людях с обогащением данных из внешних источников.
// @host localhost:8081
// @BasePath /api/v1
func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера
	logr, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logr.Sync()
	logr.Info("Сервис запущен")

	// Подключение к Postgres
	ctx := context.Background()
	db, err := pg.NewPostgres(ctx, &cfg.Database)
	if err != nil {
		logr.Fatal("Ошибка подключения к Postgres", logger.ErrorKV("error", err))
	}
	defer db.Close()

	// Инициализация репозитория
	repo, err := postgres.NewPersonRepository(db.Pool)
	if err != nil {
		logr.Fatal("Ошибка инициализации репозитория", logger.ErrorKV("error", err))
	}

	// Инициализация сервиса
	svc := service.NewPersonService(repo, logr.Logger, &cfg.APIs)

	// Инициализация HTTP-сервера
	server := api.NewServer(cfg, svc, logr.Logger)

	// Добавление Swagger UI
	server.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"),
	))

	// Запуск сервера в отдельной горутине
	go func() {
		if err := server.Start(); err != nil {
			logr.Fatal("Ошибка запуска сервера", logger.ErrorKV("error", err))
		}
	}()

	// Ожидание сигнала для завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logr.Info("Инициируется graceful shutdown")

	// Создание контекста с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Выполнение graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logr.Fatal("Ошибка graceful shutdown", logger.ErrorKV("error", err))
	}
	logr.Info("Сервис успешно остановлен")
}
