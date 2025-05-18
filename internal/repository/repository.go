// internal/repository/repository.go
package repository

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const (
	contextTxKey contextKey = "db_transaction"
)

func (r *providerRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, contextTxKey, tx)
		return fn(txCtx)
	})
}

// Used in repository methods to get transaction from context
func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(contextTxKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
