package usecases

import "context"

type TxHandle interface {
	Commit(context.Context) error
	Rollback(context.Context) error
}

type TxManager interface {
	WithTx(parent context.Context) (context.Context, TxHandle, error)
}
