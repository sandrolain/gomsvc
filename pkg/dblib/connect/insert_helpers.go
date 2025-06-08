package dbstore

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// buildPreparedStatement constructs the SQL INSERT statement and its parameters for a batch of records.
// It builds the column list, value tuples, and ON CONFLICT clause as needed.
// Returns the SQL string and a slice of parameters (empty, since values are inlined for now).
func buildPreparedStatement(args InsertRecordArgs, columns []Column, records []Record) (string, []interface{}, error) {
	tableName := args.TableName
	otherColumn := args.OtherColumn
	colNames, colMap := buildColNames(columns, otherColumn)
	batchRecordsValues, err := buildBatchValues(columns, otherColumn, colMap, records)
	if err != nil {
		return "", nil, err
	}
	upsert := buildUpsertClause(args, colNames)
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s%s`,
		tableName,
		strings.Join(colNames, ", "),
		strings.Join(batchRecordsValues, ", "),
		func() string {
			if upsert != "" {
				return " " + upsert
			}
			return ""
		}(),
	)
	params := []interface{}{}
	return query, params, nil
}

// buildColNames returns the list of column names and a map of column metadata.
// If otherColumn is set, it is appended at the end of the list.
func buildColNames(columns []Column, otherColumn string) ([]string, map[string]Column) {
	colMap := make(map[string]Column, len(columns))
	colNames := make([]string, 0, len(columns))
	for _, col := range columns {
		colMap[col.Name] = col
		if col.Name == otherColumn {
			continue
		}
		colNames = append(colNames, col.Name)
	}
	if otherColumn != "" {
		colNames = append(colNames, otherColumn)
	}
	return colNames, colMap
}

// buildBatchValues builds the value tuples for all records in the batch.
// Returns a slice of value strings, one for each record.
func buildBatchValues(columns []Column, otherColumn string, colMap map[string]Column, batchRecords []Record) ([]string, error) {
	batchRecordsValues := make([]string, 0, len(batchRecords))
	for _, record := range batchRecords {
		val, err := buildSingleRecordValues(columns, otherColumn, colMap, record)
		if err != nil {
			return nil, err
		}
		batchRecordsValues = append(batchRecordsValues, val)
	}
	return batchRecordsValues, nil
}

// buildSingleRecordValues builds the value tuple for a single record.
// It serializes extra fields as JSON if otherColumn is set.
func buildSingleRecordValues(columns []Column, otherColumn string, colMap map[string]Column, record Record) (string, error) {
	columnsMap, otherColumns := splitRecordColumns(record, colMap)
	recordValues := make([]string, 0, len(columns))
	for _, col := range columns {
		if col.Name == otherColumn {
			continue
		}
		value, ok := columnsMap[col.Name]
		if !ok {
			value = pgTypeDefault(col.Type, col.Nullable)
		}
		recordValues = append(recordValues, value)
	}
	if otherColumn != "" {
		otherJSON, err := json.Marshal(otherColumns)
		if err != nil {
			return "", err
		}
		recordValues = append(recordValues, pq.QuoteLiteral(string(otherJSON)))
	}
	return fmt.Sprintf("(%s)", strings.Join(recordValues, ", ")), nil
}

// splitRecordColumns separates known columns from extra fields in a record.
// Returns a map of known column values and a map of extra fields.
func splitRecordColumns(record Record, colMap map[string]Column) (map[string]string, map[string]interface{}) {
	columnsMap := make(map[string]string)
	otherColumns := make(map[string]interface{})
	for name, value := range record {
		if _, ok := colMap[name]; ok {
			columnsMap[name] = valueToPgValue(value)
		} else {
			otherColumns[name] = value
		}
	}
	return columnsMap, otherColumns
}

// buildUpsertClause builds the ON CONFLICT clause for the insert statement.
// Handles both DO NOTHING and DO UPDATE cases, with support for constraints or column lists.
func buildUpsertClause(args InsertRecordArgs, colNames []string) string {
	if args.OnConflict == "" {
		return ""
	}
	upsert := "ON CONFLICT "
	if args.ConflictConstraint != "" {
		upsert += fmt.Sprintf("ON CONSTRAINT %s ", args.ConflictConstraint)
	} else if args.ConflictColumns != "" {
		upsert += fmt.Sprintf("(%s) ", args.ConflictColumns)
	}
	upsert += string(args.OnConflict)
	if args.OnConflict == DoUpdate {
		upsert += " SET "
		excludedCols := make([]string, len(colNames))
		for i, col := range colNames {
			excludedCols[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
		}
		upsert += strings.Join(excludedCols, ", ")
	}
	return upsert
}
