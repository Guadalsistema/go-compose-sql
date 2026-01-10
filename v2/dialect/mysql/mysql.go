package mysql

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// MySQLDialect implements the Dialect interface for MySQL.
type MySQLDialect struct {
	registry *typeconv.Registry
}

// NewMySQLDialect creates a new MySQL dialect with type converters configured
func NewMySQLDialect() *MySQLDialect {
	registry := typeconv.NewRegistry()

	// MySQL behavior depends on connection string parameter parseTime=true
	// - With parseTime=true: returns time.Time (like PostgreSQL)
	// - Without parseTime: returns []byte/string (like SQLite)
	//
	// We register converters for both scenarios

	stringType := reflect.TypeOf("")
	bytesType := reflect.TypeOf([]byte{})
	timeType := reflect.TypeOf(time.Time{})
	nullTimeType := reflect.TypeOf(sql.NullTime{})

	// String -> time.Time (for parseTime=false)
	registry.Register(stringType, timeType, typeconv.StringToTime)

	// String -> sql.NullTime
	registry.Register(stringType, nullTimeType, typeconv.StringToNullTime)

	// []byte -> time.Time (MySQL often returns []byte)
	registry.Register(bytesType, timeType, func(source interface{}) (interface{}, error) {
		b := source.([]byte)
		return typeconv.StringToTime(string(b))
	})

	// []byte -> sql.NullTime
	registry.Register(bytesType, nullTimeType, func(source interface{}) (interface{}, error) {
		b := source.([]byte)
		return typeconv.StringToNullTime(string(b))
	})

	// Register default converters (handles multiple source types)
	registry.RegisterDefault(timeType, typeconv.DefaultTimeConverter)
	registry.RegisterDefault(nullTimeType, typeconv.DefaultNullTimeConverter)

	return &MySQLDialect{
		registry: registry,
	}
}

func (d *MySQLDialect) Placeholder(position int) string {
	return "?"
}

func (d *MySQLDialect) SupportsReturning() bool {
	return false // MySQL doesn't support RETURNING
}

func (d *MySQLDialect) Quote(identifier string) string {
	return "`" + identifier + "`"
}

// TypeRegistry returns the type converter registry for this dialect
func (d *MySQLDialect) TypeRegistry() *typeconv.Registry {
	if d.registry == nil {
		// Lazy initialization for backwards compatibility
		d.registry = typeconv.NewRegistry()
	}
	return d.registry
}
