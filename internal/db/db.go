package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DbUrl           string
	MaxOpenConns    int32         // pgx uses int32 for connection limits
	MaxIdleConns    int32         // Note: pgx pool manages idles automatically via MinConns
	ConnMaxLifetime time.Duration
}

// NewConnectionPool returns a native pgx connection pool (*pgxpool.Pool)
func NewConnectionPool(config Config) (*pgxpool.Pool, error) {
	// Parse the connection string into a structured config object
	poolConfig, err := pgxpool.ParseConfig(config.DbUrl)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = config.MaxOpenConns
	poolConfig.MaxConnLifetime = config.ConnMaxLifetime

	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// NewPool creates the pool AND immediately pings the database for you
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}