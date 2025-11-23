package usecases

import "context"

//go:generate mockgen -typed -destination ../../mocks/tx.go -package mocks . TxHandle,TxManager

type TxHandle interface {
	Commit(context.Context) error
	Rollback(context.Context) error
}

type TxManager interface {
	WithTx(parent context.Context) (context.Context, TxHandle, error)
}
