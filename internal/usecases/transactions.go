package usecases

import "context"

type TxHandle interface {
	Commit() error
	Rollback() error
}

type TxManager interface {
	WithTx(parent context.Context) (context.Context, TxHandle, error)
}
