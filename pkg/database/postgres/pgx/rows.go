package pgx

import "github.com/jackc/pgx/v5"

// RowsWrapper wraps pgx.Rows to implement interfaces.Rows
type RowsWrapper struct {
	rows pgx.Rows
}

func (r *RowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *RowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *RowsWrapper) Close() error {
	r.rows.Close()
	return nil
}

func (r *RowsWrapper) Err() error {
	return r.rows.Err()
}

// RowWrapper wraps pgx.Row to implement interfaces.Row
type RowWrapper struct {
	row pgx.Row
}

func (r *RowWrapper) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}
