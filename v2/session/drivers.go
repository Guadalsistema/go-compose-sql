package session

import "fmt"

// SQLiteDriver implements the Driver interface for SQLite
type SQLiteDriver struct{}

func (d *SQLiteDriver) Placeholder(position int) string {
	return "?"
}

func (d *SQLiteDriver) SupportsReturning() bool {
	return true // SQLite 3.35.0+ supports RETURNING
}

func (d *SQLiteDriver) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// PostgresDriver implements the Driver interface for PostgreSQL
type PostgresDriver struct{}

func (d *PostgresDriver) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (d *PostgresDriver) SupportsReturning() bool {
	return true
}

func (d *PostgresDriver) Quote(identifier string) string {
	return `"` + identifier + `"`
}

// MySQLDriver implements the Driver interface for MySQL
type MySQLDriver struct{}

func (d *MySQLDriver) Placeholder(position int) string {
	return "?"
}

func (d *MySQLDriver) SupportsReturning() bool {
	return false // MySQL doesn't support RETURNING
}

func (d *MySQLDriver) Quote(identifier string) string {
	return "`" + identifier + "`"
}

// DriverByName returns a driver by name
func DriverByName(name string) (Driver, error) {
	switch name {
	case "sqlite", "sqlite3":
		return &SQLiteDriver{}, nil
	case "postgres", "postgresql":
		return &PostgresDriver{}, nil
	case "mysql":
		return &MySQLDriver{}, nil
	default:
		return nil, fmt.Errorf("unknown driver: %s", name)
	}
}
