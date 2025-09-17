package _gorm_postgres

import (
	"context"

	_postgres "go-libs/pkg/postgres"

	"gorm.io/gorm"
)

// Transaction wraps GORM transaction
type Transaction struct {
	tx *gorm.DB
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}

// Exec executes a query within transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...any) error {
	return t.tx.WithContext(ctx).Exec(query, args...).Error
}

// Query executes a query and returns rows within transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...any) (_postgres.Rows, error) {
	rows, err := t.tx.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	return &RowsWrapper{rows: rows}, nil
}

// QueryRow executes a query and returns a single row within transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...any) _postgres.Row {
	row := t.tx.WithContext(ctx).Raw(query, args...).Row()
	return &RowWrapper{row: row}
}
