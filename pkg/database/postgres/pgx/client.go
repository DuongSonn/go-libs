package pgx

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go-libs/pkg/database/postgres/interfaces"

	"github.com/jackc/pgx/v5"
)

// BeginTx starts a new transaction
func (c *Connection) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("database not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	tx, err := c.pool.Begin(queryCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx, ctx: queryCtx}, nil
}

// Exec executes a query
func (c *Connection) Exec(ctx context.Context, query string, args ...any) error {
	if c.pool == nil {
		return fmt.Errorf("database not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	_, err := c.pool.Exec(queryCtx, query, args...)
	return err
}

// Query executes a query and returns rows
func (c *Connection) Query(ctx context.Context, query string, args ...any) (interfaces.Rows, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("database not connected")
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	rows, err := c.pool.Query(queryCtx, query, args...)
	if err != nil {
		return nil, err
	}

	return &RowsWrapper{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (c *Connection) QueryRow(ctx context.Context, query string, args ...any) interfaces.Row {
	if c.pool == nil {
		return &RowWrapper{row: nil}
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	row := c.pool.QueryRow(queryCtx, query, args...)
	return &RowWrapper{row: row}
}

// InsertModel inserts a model into the database
func (c *Connection) InsertModel(ctx context.Context, model any) error {
	if c.pool == nil {
		return fmt.Errorf("database not connected")
	}

	query, args, err := c.buildInsertQuery(model)
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	_, err = c.pool.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert model: %w", err)
	}

	return nil
}

// UpsertModel performs an upsert (insert or update) operation
func (c *Connection) UpsertModel(ctx context.Context, model any, primaryKeys ...string) error {
	if c.pool == nil {
		return fmt.Errorf("database not connected")
	}

	query, args, err := c.buildUpsertQuery(model, primaryKeys...)
	if err != nil {
		return fmt.Errorf("failed to build upsert query: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	_, err = c.pool.Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to upsert model: %w", err)
	}

	return nil
}

// BatchInsertModel performs batch insert of multiple models
func (c *Connection) BatchInsertModel(ctx context.Context, models []any, batchSize int) error {
	if c.pool == nil {
		return fmt.Errorf("database not connected")
	}

	if len(models) == 0 {
		return nil
	}

	// Process in batches
	for i := 0; i < len(models); i += batchSize {
		end := i + batchSize
		if end > len(models) {
			end = len(models)
		}

		batch := models[i:end]
		if err := c.executeBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to execute batch at index %d: %w", i, err)
		}
	}

	return nil
}

// executeBatch executes a batch of inserts
func (c *Connection) executeBatch(ctx context.Context, models []any) error {
	if len(models) == 0 {
		return nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	// Use the first model to determine the structure
	query, err := c.buildBatchInsertQuery(models)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, model := range models {
		args, err := c.extractModelValues(model)
		if err != nil {
			return err
		}
		batch.Queue(query, args...)
	}

	results := c.pool.SendBatch(queryCtx, batch)
	defer results.Close()

	// Process all results
	for i := 0; i < len(models); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch item %d: %w", i, err)
		}
	}

	return nil
}

// Helper methods for building queries
func (c *Connection) buildInsertQuery(model any) (string, []any, error) {
	tableName, columns, values, err := c.analyzeModel(model)
	if err != nil {
		return "", nil, err
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, values, nil
}

func (c *Connection) buildUpsertQuery(model any, primaryKeys ...string) (string, []any, error) {
	tableName, columns, values, err := c.analyzeModel(model)
	if err != nil {
		return "", nil, err
	}

	if len(primaryKeys) == 0 {
		primaryKeys = []string{"id"} // Default primary key
	}

	placeholders := make([]string, len(columns))
	updateClauses := make([]string, 0)

	for i, col := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)

		// Don't update primary keys
		isPrimaryKey := false
		for _, pk := range primaryKeys {
			if col == pk {
				isPrimaryKey = true
				break
			}
		}

		if !isPrimaryKey {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
	}

	conflictColumns := strings.Join(primaryKeys, ", ")
	updateClause := strings.Join(updateClauses, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		conflictColumns,
		updateClause)

	return query, values, nil
}

func (c *Connection) buildBatchInsertQuery(models []any) (string, error) {
	if len(models) == 0 {
		return "", fmt.Errorf("no models provided")
	}

	tableName, columns, _, err := c.analyzeModel(models[0])
	if err != nil {
		return "", err
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, nil
}

func (c *Connection) analyzeModel(model any) (tableName string, columns []string, values []any, err error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", nil, nil, fmt.Errorf("model must be a struct")
	}

	t := v.Type()

	// Get table name from struct name (convert to snake_case)
	tableName = toSnakeCase(t.Name())

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get column name from tag or field name
		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = toSnakeCase(field.Name)
		}

		// Skip fields marked with "-"
		if columnName == "-" {
			continue
		}

		columns = append(columns, columnName)
		values = append(values, fieldValue.Interface())
	}

	return tableName, columns, values, nil
}

func (c *Connection) extractModelValues(model any) ([]any, error) {
	_, _, values, err := c.analyzeModel(model)
	return values, err
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
