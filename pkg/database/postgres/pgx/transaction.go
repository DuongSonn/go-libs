package pgx

import (
	"context"

	"go-libs/pkg/database/postgres/interfaces"

	"github.com/jackc/pgx/v5"
)

// Transaction wraps GORM transaction
type Transaction struct {
	tx  pgx.Tx
	ctx context.Context
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit(t.ctx)
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback(t.ctx)
}

// Exec executes a query within transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...any) error {
	_, err := t.tx.Exec(t.ctx, query, args...)
	return err
}

// Query executes a query and returns rows within transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...any) (interfaces.Rows, error) {
	rows, err := t.tx.Query(t.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &RowsWrapper{rows: rows}, nil
}

// QueryRow executes a query and returns a single row within transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...any) interfaces.Row {
	row := t.tx.QueryRow(t.ctx, query, args...)
	return &RowWrapper{row: row}
}
