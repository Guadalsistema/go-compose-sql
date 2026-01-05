package mysql

// MySQLDialect implements the Dialect interface for MySQL.
type MySQLDialect struct{}

func (d *MySQLDialect) Placeholder(position int) string {
	return "?"
}

func (d *MySQLDialect) SupportsReturning() bool {
	return false // MySQL doesn't support RETURNING
}

func (d *MySQLDialect) Quote(identifier string) string {
	return "`" + identifier + "`"
}
