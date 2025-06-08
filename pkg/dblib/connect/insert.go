package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// InsertRecord inserts records into the specified table in batches.
// It retrieves column metadata, splits the records into batches, and calls insertBatch for each batch.
// If TableName is empty, returns ErrInvalidTableName. If BatchRecords is empty, does nothing.
// BatchSize defaults to 100 if not specified.
func InsertRecord(ctx context.Context, db *sql.DB, args InsertRecordArgs) error {
	if args.TableName == "" {
		return ErrInvalidTableName
	}
	if len(args.BatchRecords) == 0 {
		return nil
	}
	if args.BatchSize <= 0 {
		args.BatchSize = 100
	}

	// Retrieve column metadata for the table
	columns, err := GetTableColumns(ctx, db, args.TableName)
	if err != nil {
		return fmt.Errorf("failed to get table columns: %w", err)
	}

	filteredColumns := filterInsertableColumns(columns)

	// Process records in batches
	for i := 0; i < len(args.BatchRecords); i += args.BatchSize {
		end := i + args.BatchSize
		if end > len(args.BatchRecords) {
			end = len(args.BatchRecords)
		}
		if err := insertBatch(ctx, db, args, filteredColumns, args.BatchRecords[i:end]); err != nil {
			return fmt.Errorf("failed to insert batch: %w", err)
		}
	}
	return nil
}

// filterInsertableColumns excludes auto-increment (serial/identity) primary key columns from insert
func filterInsertableColumns(columns []Column) []Column {
	filteredColumns := make([]Column, 0, len(columns))
	for _, col := range columns {
		typeLower := strings.ToLower(col.Type)
		isPK := strings.Contains(typeLower, "|key=primary key")
		isAuto := strings.Contains(typeLower, "|default=nextval(")
		if isPK && isAuto {
			// Exclude PK with nextval default (serial/identity)
			continue
		}
		filteredColumns = append(filteredColumns, col)
	}
	return filteredColumns
}

// insertBatch builds and executes a prepared statement for a batch of records.
// It uses buildPreparedStatement to generate the SQL and parameters, prepares the statement,
// executes it, and returns any error encountered.
func insertBatch(ctx context.Context, db *sql.DB, args InsertRecordArgs, columns []Column, records []Record) error {
	query, params, err := buildPreparedStatement(args, columns, records)
	if err != nil {
		return fmt.Errorf("failed to build prepared statement: %w", err)
	}
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
			fmt.Printf("failed to close statement: %v\n", err)
		}
	}()
	_, err = stmt.ExecContext(ctx, params...)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}
