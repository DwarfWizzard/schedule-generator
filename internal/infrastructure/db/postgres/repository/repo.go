package repository

import (
	"context"
	"database/sql"
	"errors"
	"schedule-generator/internal/infrastructure/db"

	"gorm.io/gorm"
)

type Repository struct {
	client *gorm.DB
	isTxn  bool
}

func NewPostgresRepository(client *gorm.DB) *Repository {
	return &Repository{
		client: client,
	}
}

// AsTransaction returns repository instance initiated by transaction and tx handling functions
func (r *Repository) AsTransaction(ctx context.Context, isoLevel db.IsoLevel) (db.TransactionalRepository, db.RollbackTxnFunc, db.CommitTxnFunc, error) {
	if r.isTxn {
		return nil, nil, nil, errors.New("repository already transactional")
	}

	tx := r.client.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: isoLevel.ToSQLIsolationLevel(),
	})
	if tx.Error != nil {
		return nil, nil, nil, tx.Error
	}

	rollback := func(ctx context.Context) error {
		err := tx.WithContext(ctx).Rollback().Error
		if err != nil {
			return err
		}

		return nil
	}

	commit := func(ctx context.Context) error {
		err := tx.WithContext(ctx).Commit().Error
		if err != nil {
			return err
		}

		return nil
	}

	newRepo := NewPostgresRepository(tx)
	newRepo.isTxn = true

	return newRepo, rollback, commit, nil
}
