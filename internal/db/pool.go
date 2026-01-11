package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
)

// NewPool creates a pgx connection pool to PostgreSQL/TimescaleDB and pings it.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, sslmode,
	)

	pcfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Sensible defaults
	if cfg.MaxConns > 0 {
		pcfg.MaxConns = cfg.MaxConns
	} else {
		pcfg.MaxConns = 10
	}

	if cfg.MinConns > 0 {
		pcfg.MinConns = cfg.MinConns
	} else {
		pcfg.MinConns = 2
	}

	if cfg.MaxConnLifetime.Duration > 0 {
		pcfg.MaxConnLifetime = cfg.MaxConnLifetime.Duration
	} else {
		pcfg.MaxConnLifetime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
