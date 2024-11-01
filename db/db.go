package db

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
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

// Exec executes a SQL query without returning any rows
func (s *Service) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
    return s.pool.Exec(ctx, sql, arguments...)
}

// Query executes a query that returns rows
func (s *Service) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
    return s.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row
func (s *Service) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
    return s.pool.QueryRow(ctx, sql, args...)
}