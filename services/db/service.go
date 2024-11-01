package db

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
    pool *pgxpool.Pool
}

func NewService(cfg Config) (*Service, error) {
    pool, err := pgxpool.New(context.Background(), cfg.URL)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %w", err)
    }

    // Test the connection
    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("unable to ping database: %w", err)
    }

    return &Service{pool: pool}, nil
}

func (s *Service) GetPool() *pgxpool.Pool {
    return s.pool
}

func (s *Service) Close() {
    if s.pool != nil {
        s.pool.Close()
    }
}