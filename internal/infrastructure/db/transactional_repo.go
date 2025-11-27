package db

import (
	"context"
	"database/sql"
)

type IsoLevel int

const (
	IsoLevelDefault IsoLevel = iota
	IsoLevelReadUncommitted
	IsoLevelReadCommitted
	IsoLevelWriteCommitted
	IsoLevelRepeatableRead
	IsoLevelSnapshot
	IsoLevelSerializable
	IsoLevelLinearizable
)

func (l IsoLevel) ToSQLIsolationLevel() sql.IsolationLevel {
	return sql.IsolationLevel(l)
}

type CommitTxnFunc func(context.Context) error
type RollbackTxnFunc func(context.Context) error

type TransactionalRepository interface {
	AsTransaction(ctx context.Context, isoLevel IsoLevel) (TransactionalRepository, RollbackTxnFunc, CommitTxnFunc, error)
}
