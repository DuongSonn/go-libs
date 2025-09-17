package _gorm_postgres

import (
	"context"
	"database/sql"
	"fmt"

	_postgres "go-libs/pkg/postgres"
)

// BeginTx starts a new transaction
func (c *Connection) BeginTx(ctx context.Context) (_postgres.Transaction, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Transaction{tx: tx}, nil
}

// Exec executes a query
func (c *Connection) Exec(ctx context.Context, query string, args ...any) error {
	if c.db == nil {
		return fmt.Errorf("database not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	return c.db.WithContext(queryCtx).Exec(query, args...).Error
}

// Query executes a query and returns rows
func (c *Connection) Query(ctx context.Context, query string, args ...any) (_postgres.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	rows, err := c.db.WithContext(queryCtx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}

	return &RowsWrapper{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (c *Connection) QueryRow(ctx context.Context, query string, args ...any) _postgres.Row {
	if c.db == nil {
		return &RowWrapper{row: &sql.Row{}}
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	row := c.db.WithContext(queryCtx).Raw(query, args...).Row()
	return &RowWrapper{row: row}
}
