package repository

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wgorm"
)

type IRepository interface {
	Transaction(ctx context.Context, fc func(txRepo IRepository) error) (err error)
	UserRepo
}

type repo struct {
	db *wgorm.Gorm
}

func New(db *wgorm.Gorm) IRepository {
	return &repo{
		db: db,
	}
}

func (r *repo) Transaction(ctx context.Context, fc func(txRepo IRepository) error) (err error) {
	panicked := true
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Make sure to rollback when panic, Block error or Commit error
		if panicked || err != nil {
			if err := tx.Rollback(); err != nil {
				log.Ctx(ctx).Error().Msgf("rollback failed: %+v", err)
			}
		}
	}()

	txRepo := &repo{db: tx}
	err = fc(txRepo)
	if err == nil {
		err = tx.Commit()
	}

	panicked = false
	return
}
