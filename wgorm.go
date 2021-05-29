package wgorm

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Option func(g *Gorm) *Gorm

type Gorm struct {
	*gorm.DB
	conn connection
}

type connection struct {
	db  *gorm.DB
	cfg *Config
}

func New(cfg *Config) (*Gorm, error) {
	config, err := cfg.clone()
	if err != nil {
		return nil, err
	}
	db, err := config.newGormDB()
	if err != nil {
		return nil, err
	}

	conn := connection{
		db:  db,
		cfg: config,
	}

	return &Gorm{
		conn: conn,
	}, nil
}

func (g *Gorm) WithContext(ctx context.Context) *Gorm {
	var db *gorm.DB
	if g.DB != nil {
		return g
	}

	db = g.conn.db.WithContext(ctx)
	return &Gorm{
		DB:   db,
		conn: g.conn,
	}
}

func (g *Gorm) Begin(ctx context.Context, opts ...*sql.TxOptions) (*Gorm, error) {
	tx := g.WithContext(ctx).GormDB().Begin(opts...)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Gorm{
		DB:   tx,
		conn: g.conn,
	}, nil
}

func (g *Gorm) Commit() error {
	return g.GormDB().Commit().Error
}

func (g *Gorm) Rollback() error {
	return g.GormDB().Rollback().Error
}

func (g *Gorm) clone() *Gorm {
	return &Gorm{
		DB:   g.DB,
		conn: g.conn,
	}
}

func (g *Gorm) setDB(tx *gorm.DB) *Gorm {
	g.DB = tx
	return g
}

func (g *Gorm) GormDB() *gorm.DB {
	return g.DB
}

func (g *Gorm) Options(opts ...Option) *Gorm {
	for _, opt := range opts {
		g = opt(g)
	}
	return g
}

func SetForUpdate() Option {
	return func(g *Gorm) *Gorm {
		tx := g.DB.Clauses(clause.Locking{Strength: "UPDATE"})
		return g.clone().setDB(tx)
	}
}
