package dialect

import (
	"fmt"

	"github.com/guadalsistema/go-compose-sql/v2/dialect/mysql"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/postgres"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/sqlite"
	"github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// Dialect represents a SQL dialect (placeholder/quoting behavior).
type Dialect interface {
	// Placeholder returns the placeholder format for this driver
	// e.g., "?" for SQLite/MySQL, "$" for Postgres
	Placeholder(position int) string

	// SupportsReturning indicates if the driver supports RETURNING clauses
	SupportsReturning() bool

	// Quote quotes an identifier (table/column name)
	Quote(identifier string) string

	// TypeRegistry returns the type converter registry for this dialect
	// Used to handle type conversions between database and Go types
	TypeRegistry() *typeconv.Registry
}

// DialectByName returns a dialect by name
func DialectByName(name string) (Dialect, error) {
	switch name {
	case "sqlite", "sqlite3":
		return sqlite.NewSQLiteDialect(), nil
	case "postgres", "postgresql":
		return postgres.NewPostgresDialect(), nil
	case "mysql":
		return mysql.NewMySQLDialect(), nil
	default:
		return nil, fmt.Errorf("unknown driver: %s", name)
	}
}
