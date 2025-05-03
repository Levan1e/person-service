package postgres

import (
	"context"
	"fmt"
	"person-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool представляет пул соединений к Postgres.
type Pool struct {
	*pgxpool.Pool
}

// NewPostgres создаёт новый пул соединений к Postgres на основе конфигурации.
func NewPostgres(ctx context.Context, cfg *config.Database) (*Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать DSN: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Close закрывает пул соединений.
func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
