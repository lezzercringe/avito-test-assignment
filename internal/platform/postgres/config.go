package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	ConnString  string        `yaml:"conn_string"`
	InitTimeout time.Duration `yaml:"initial_timeout"`
}

func SetupPool(config Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.InitTimeout)
	defer cancel()
	return pgxpool.New(ctx, config.ConnString)
}
