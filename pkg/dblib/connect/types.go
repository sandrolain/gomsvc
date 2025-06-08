package dbstore

import "time"

// Column represents a column in a PostgreSQL table.
type Column struct {
	Name     string // Column name
	Type     string // Data type
	Nullable bool   // True if the column is nullable
}

// CacheEntry stores column metadata with expiration for caching purposes.
type CacheEntry struct {
	Columns   []Column  // List of columns
	ExpiresAt time.Time // Expiration timestamp (Unix seconds)
}

// Record represents a generic row as a map of column names to values.
type Record map[string]interface{}

// InsertRecordOnConflict defines the ON CONFLICT behavior for inserts.
type InsertRecordOnConflict string

const (
	DoNothing InsertRecordOnConflict = "DO NOTHING" // Ignore conflicts
	DoUpdate  InsertRecordOnConflict = "DO UPDATE"  // Update on conflict
)

// InsertRecordArgs contains all arguments for the InsertRecord function.
type InsertRecordArgs struct {
	TableName          string                 // Name of the table to insert into
	OtherColumn        string                 // Name of the column for extra fields (as JSON)
	BatchRecords       []Record               // Records to insert
	OnConflict         InsertRecordOnConflict // ON CONFLICT behavior
	ConflictConstraint string                 // Optional constraint name for ON CONFLICT
	ConflictColumns    string                 // Optional column list for ON CONFLICT
	BatchSize          int                    // Maximum number of records per batch
	CacheExpiration    int64                  // How long to cache table metadata (seconds)
}
