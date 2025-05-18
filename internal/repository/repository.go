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

// Used in repository methods to get transaction from context
func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(contextTxKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
