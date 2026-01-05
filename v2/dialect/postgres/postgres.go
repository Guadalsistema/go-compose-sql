package postgres

import "fmt"

// PostgresDialect implements the Dialect interface for PostgreSQL.
type PostgresDialect struct{}

func (d *PostgresDialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (d *PostgresDialect) SupportsReturning() bool {
	return true
}

func (d *PostgresDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}
