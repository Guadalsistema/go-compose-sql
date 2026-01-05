package sqlite

// SQLiteDialect implements the Dialect interface for SQLite.
type SQLiteDialect struct{}

func (d *SQLiteDialect) Placeholder(position int) string {
	return "?"
}

func (d *SQLiteDialect) SupportsReturning() bool {
	return true // SQLite 3.35.0+ supports RETURNING
}

func (d *SQLiteDialect) Quote(identifier string) string {
	return `"` + identifier + `"`
}
