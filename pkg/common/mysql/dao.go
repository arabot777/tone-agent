package mysql

import (
	"context"

	"gorm.io/gorm"
)

type Datastore struct {
	db *gorm.DB
}

func (ds *Datastore) Close() error {
	sqlDb, err := ds.db.DB()
	if err != nil {
		return err
	}
	return sqlDb.Close()
}

func (ds *Datastore) DBIns() *gorm.DB {
	return ds.db
}

type contextTxKey struct{}

func (ds *Datastore) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

func (ds *Datastore) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return ds.db
}

func (ds *Datastore) DBWithContext(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx.WithContext(ctx)
	}
	return ds.db.WithContext(ctx)
}
