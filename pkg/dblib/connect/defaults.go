package dbstore

import "strings"

// defaultValues maps PostgreSQL types to their default values as string literals.
var defaultValues = map[string]string{
	"integer":       "0",
	"bigint":        "0",
	"smallint":      "0",
	"serial":        "0",
	"bigserial":     "0",
	"boolean":       "false",
	"text":          "''",
	"varchar":       "''",
	"char":          "''",
	"date":          "'1970-01-01'",
	"timestamp":     "'1970-01-01 00:00:00'",
	"timestamptz":   "'1970-01-01 00:00:00+00'",
	"time":          "'00:00:00'",
	"timetz":        "'00:00:00+00'",
	"uuid":          "'00000000-0000-0000-0000-000000000000'",
	"json":          "'{}'",
	"jsonb":         "'{}'",
	"bytea":         "'\\x'",
	"float4":        "0.0",
	"float8":        "0.0",
	"numeric":       "0.0",
	"money":         "0.0",
	"interval":      "'00:00:00'",
	"point":         "'(0,0)'",
	"line":          "'{0,0,0}'",
	"lseg":          "'[(0,0),(0,0)]'",
	"box":           "'((0,0),(0,0))'",
	"path":          "'[]'",
	"polygon":       "'((0,0),(0,0),(0,0))'",
	"circle":        "'<(0,0),0>'",
	"cidr":          "'0.0.0.0/0'",
	"inet":          "'0.0.0.0'",
	"macaddr":       "'00:00:00:00:00:00'",
	"bit":           "B'0'",
	"varbit":        "B'0'",
	"tsvector":      "' '",
	"tsquery":       "' '",
	"xml":           "'<root></root>'",
	"array":         "'{}'",
	"composite":     "'(0,0)'",
	"range":         "'[0,0)'",
	"regproc":       "0",
	"regprocedure":  "0",
	"regoper":       "0",
	"regoperator":   "0",
	"regclass":      "0",
	"regtype":       "0",
	"regrole":       "0",
	"regnamespace":  "0",
	"regconfig":     "0",
	"regdictionary": "0",
}

// pgTypeDefault returns the default value for a PostgreSQL type as a string literal.
// If the column is nullable, returns NULL. If the type is not found, returns NULL.
func pgTypeDefault(pgType string, nullable bool) string {
	if nullable {
		return "NULL"
	}
	pgType = strings.ToLower(pgType)
	if value, exists := defaultValues[pgType]; exists {
		return value
	}
	return "NULL"
}
