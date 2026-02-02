package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Querier interface {
	Exec(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

type querier struct {
	pool DBTX
}

func NewQuerier(pool DBTX) Querier {
	return &querier{
		pool: pool,
	}
}

func (c *querier) Exec(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return c.pool.Query(ctx, query, args...)
}
