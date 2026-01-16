package db

import (
	"context"
	"github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
)

type TxCreator interface {
	CreateTx(ctx context.Context) (tx pgx.Tx, err error)
}

type txCreator struct {
	conn pgxv5.Tr
}

func NewTxCreator(conn pgxv5.Tr) TxCreator {
	return &txCreator{conn: conn}
}

func (c *txCreator) CreateTx(ctx context.Context) (tx pgx.Tx, err error) {
	return c.conn.Begin(ctx)
}
