package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
)

type contextKey struct{}

var txKey = contextKey{}

type PgTxManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) usecases.TxManager {
	return &PgTxManager{db: db}
}

func (m *PgTxManager) WithTx(parent context.Context) (context.Context, usecases.TxHandle, error) {
	tx, err := m.db.Begin(parent)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.WithValue(parent, txKey, tx)
	handle := &PgTxHandle{tx: tx}

	return ctx, handle, nil
}

type PgTxHandle struct {
	tx pgx.Tx
}

func (h *PgTxHandle) Commit(ctx context.Context) error {
	return h.tx.Commit(ctx)
}

func (h *PgTxHandle) Rollback(ctx context.Context) error {
	return h.tx.Rollback(ctx)
}
