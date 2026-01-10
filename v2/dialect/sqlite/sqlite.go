package sqlite

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// SQLiteDialect implements the Dialect interface for SQLite.
type SQLiteDialect struct {
	registry *typeconv.Registry
}

// NewSQLiteDialect creates a new SQLite dialect with type converters configured
func NewSQLiteDialect() *SQLiteDialect {
	registry := typeconv.NewRegistry()

	// SQLite stores timestamps as TEXT (ISO8601 strings)
	// Register converters for time types
	stringType := reflect.TypeOf("")
	int64Type := reflect.TypeOf(int64(0))
	timeType := reflect.TypeOf(time.Time{})
	nullTimeType := reflect.TypeOf(sql.NullTime{})
	boolType := reflect.TypeOf(true)

	// String -> time.Time (most common case for SQLite)
	registry.Register(stringType, timeType, typeconv.StringToTime)

	// String -> sql.NullTime
	registry.Register(stringType, nullTimeType, typeconv.StringToNullTime)

	// Int64 -> time.Time (Unix timestamp)
	registry.Register(int64Type, timeType, typeconv.Int64ToTime)

	// Int64 -> sql.NullTime
	registry.Register(int64Type, nullTimeType, typeconv.Int64ToNullTime)

	// Int64 -> bool (SQLite uses 0/1 for booleans)
	registry.Register(int64Type, boolType, typeconv.Int64ToBool)

	// Register default converters (handle multiple source types)
	registry.RegisterDefault(timeType, typeconv.DefaultTimeConverter)
	registry.RegisterDefault(nullTimeType, typeconv.DefaultNullTimeConverter)
	registry.RegisterDefault(boolType, typeconv.DefaultBoolConverter)

	return &SQLiteDialect{
		registry: registry,
	}
}

func (d *SQLiteDialect) Placeholder(position int) string {
	return "?"
}

func (d *SQLiteDialect) SupportsReturning() bool {
	return true // SQLite 3.35.0+ supports RETURNING
}

func (d *SQLiteDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// TypeRegistry returns the type converter registry for this dialect
func (d *SQLiteDialect) TypeRegistry() *typeconv.Registry {
	if d.registry == nil {
		// Lazy initialization for backwards compatibility
		d.registry = typeconv.NewRegistry()
	}
	return d.registry
}
