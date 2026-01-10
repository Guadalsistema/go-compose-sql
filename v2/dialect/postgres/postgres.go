package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// PostgresDialect implements the Dialect interface for PostgreSQL.
type PostgresDialect struct {
	registry *typeconv.Registry
}

// NewPostgresDialect creates a new PostgreSQL dialect with type converters configured
func NewPostgresDialect() *PostgresDialect {
	registry := typeconv.NewRegistry()

	// PostgreSQL driver (lib/pq) handles most types natively
	// We only need converters for edge cases or alternative formats

	timeType := reflect.TypeOf(time.Time{})
	nullTimeType := reflect.TypeOf(sql.NullTime{})

	// Register default converters for flexibility (handles string/int64 if needed)
	registry.RegisterDefault(timeType, typeconv.DefaultTimeConverter)
	registry.RegisterDefault(nullTimeType, typeconv.DefaultNullTimeConverter)

	return &PostgresDialect{
		registry: registry,
	}
}

func (d *PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (d *PostgresDialect) SupportsReturning() bool {
	return true
}

func (d *PostgresDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// TypeRegistry returns the type converter registry for this dialect
func (d *PostgresDialect) TypeRegistry() *typeconv.Registry {
	if d.registry == nil {
		// Lazy initialization for backwards compatibility
		d.registry = typeconv.NewRegistry()
	}
	return d.registry
}
