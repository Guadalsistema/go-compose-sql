package dialect

import (
	"fmt"

	"github.com/guadalsistema/go-compose-sql/v2/dialect/mysql"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/postgres"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/sqlite"
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

	// FormatIgnoreConflict returns the SQL fragment for ignoring conflicts
	// Returns empty string if not supported by the dialect
	FormatIgnoreConflict() string
}

// DialectByName returns a dialect by name
func DialectByName(name string) (Dialect, error) {
	switch name {
	case "sqlite", "sqlite3":
		return &sqlite.SQLiteDialect{}, nil
	case "postgres", "postgresql":
		return &postgres.PostgresDialect{}, nil
	case "mysql":
		return &mysql.MySQLDialect{}, nil
	default:
		return nil, fmt.Errorf("unknown driver: %s", name)
	}
}
