package builder

// Builder is the interface that all query builders must implement.
// It provides a method to generate SQL queries with their arguments.
type Builder interface {
	// ToSQL generates the SQL query string and arguments
	ToSQL() (string, []interface{}, error)
}
