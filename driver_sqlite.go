package sqlcompose

// DefaultDriver is used when no driver is provided on a statement.
var DefaultDriver Driver = SQLiteDriver{}

// SQLiteDriver renders SQL using question mark placeholders.
type SQLiteDriver struct{}

// Write renders the clause using positional question mark placeholders.
func (SQLiteDriver) Write(clause SqlClause, argPosition int) (string, int, error) {
	return writeClause(clause, argPosition, questionPlaceholder{})
}

type questionPlaceholder struct{}

func (questionPlaceholder) Placeholder(_ int) string { return "?" }
