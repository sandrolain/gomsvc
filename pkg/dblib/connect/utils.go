package dbstore

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// valueToPgValue converts a Go value to a PostgreSQL literal string suitable for inline SQL.
// Handles common types, including time.Time, []byte, and nulls.
func valueToPgValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "NULL"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return pq.QuoteLiteral(v)
	case []byte:
		return fmt.Sprintf("'\\x%s'::bytea", hex.EncodeToString(v))
	case time.Time:
		return pq.QuoteLiteral(v.UTC().Format(time.RFC3339))
	default:
		return "NULL"
	}
}
