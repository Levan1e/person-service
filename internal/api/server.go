package api

import (
	"context"
	"fmt"
	"net/http"
	v1 "person-service/internal/api/v1"
	"person-service/internal/config"
	"person-service/internal/service"
	"person-service/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Server представляет HTTP-сервер.
type Server struct {
	httpServer *http.Server
	Router     *chi.Mux
	logger     *zap.Logger
}

// NewServer создаёт новый HTTP-сервер.
func NewServer(cfg *config.Config, service *service.PersonService, logger *zap.Logger) *Server {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API v1
	handler := v1.NewHandler(service, logger)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/persons", handler.CreatePerson)
		r.Get("/persons", handler.ListPersons)
		r.Get("/persons/{id}", handler.GetPerson)
		r.Put("/persons/{id}", handler.UpdatePerson)
		r.Delete("/persons/{id}", handler.DeletePerson)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	return &Server{
		httpServer: httpServer,
		Router:     r,
		logger:     logger,
	}
}

// Start запускает сервер.
func (s *Server) Start() error {
	s.logger.Info("Запуск HTTP-сервера", logger.InfoKV("port", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}
	return nil
}

// Shutdown останавливает сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Остановка HTTP-сервера")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка остановки сервера: %w", err)
	}
	return nil
}
