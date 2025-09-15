package gorm

import "database/sql"

// RowsWrapper wraps sql.Rows to implement interfaces.Rows
type RowsWrapper struct {
	rows *sql.Rows
}

func (r *RowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *RowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *RowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *RowsWrapper) Err() error {
	return r.rows.Err()
}

// RowWrapper wraps sql.Row to implement interfaces.Row
type RowWrapper struct {
	row *sql.Row
}

func (r *RowWrapper) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}
