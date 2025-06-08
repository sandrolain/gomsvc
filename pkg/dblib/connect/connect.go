package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// columnCache stores column metadata for tables, with expiration.
var (
	columnCache = make(map[string]CacheEntry)
	cacheMutex  sync.RWMutex
)

// ErrInvalidTableName is returned when the table name is empty.
var ErrInvalidTableName = errors.New("table name is required")

// GetTableColumns retrieves columns and their types from the specified PostgreSQL table.
// It uses an in-memory cache to avoid repeated queries for the same table within a 5-minute window.
// If the table name is empty, ErrInvalidTableName is returned.
func GetTableColumns(ctx context.Context, db *sql.DB, tableName string) ([]Column, error) {
	if tableName == "" {
		return nil, ErrInvalidTableName
	}

	now := time.Now()

	if cols, ok := getCachedColumns(tableName, now); ok {
		return cols, nil
	}

	columns, err := fetchTableColumns(ctx, db, tableName)
	if err != nil {
		return nil, err
	}

	cacheMutex.Lock()
	columnCache[tableName] = CacheEntry{
		Columns:   columns,
		ExpiresAt: now.Add(5 * time.Minute),
	}
	cacheMutex.Unlock()

	return columns, nil
}

// getCachedColumns returns columns from cache if valid
func getCachedColumns(tableName string, now time.Time) ([]Column, bool) {
	cacheMutex.RLock()
	entry, found := columnCache[tableName]
	cacheMutex.RUnlock()
	if found && now.Before(entry.ExpiresAt) {
		return entry.Columns, true
	}
	return nil, false
}

// fetchTableColumns queries the DB for column metadata, including default and PK info
func fetchTableColumns(ctx context.Context, db *sql.DB, tableName string) ([]Column, error) {
	query := `
		SELECT c.column_name, c.data_type, c.is_nullable, c.udt_name, c.column_default, (
			SELECT tc.constraint_type
			 FROM information_schema.key_column_usage kcu
			 JOIN information_schema.table_constraints tc
			   ON tc.constraint_name = kcu.constraint_name
			  AND tc.table_name = c.table_name
			WHERE kcu.column_name = c.column_name
			  AND kcu.table_name = c.table_name
			  AND tc.constraint_type = 'PRIMARY KEY'
			LIMIT 1
		) as column_key
		FROM information_schema.columns c
		WHERE c.table_name = $1
	`
	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query table columns: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var columns []Column
	for rows.Next() {
		var col Column
		var isNullable, udtName, columnDefault, columnKey sql.NullString
		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &udtName, &columnDefault, &columnKey); err != nil {
			return nil, fmt.Errorf("failed to scan column data: %w", err)
		}
		col.Nullable = (isNullable.Valid && isNullable.String == "YES")
		if udtName.Valid && strings.HasPrefix(udtName.String, "_") {
			col.Type = fmt.Sprintf("%s[]", strings.TrimPrefix(udtName.String, "_"))
		}
		if columnDefault.Valid {
			col.Type += "|default=" + columnDefault.String
		}
		if columnKey.Valid {
			col.Type += "|key=" + columnKey.String
		}
		columns = append(columns, col)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return columns, nil
}
