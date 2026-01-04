package sqlcompose

import "fmt"

// PostgresDriver renders SQL using dollar-prefixed placeholders.
type PostgresDriver struct{}

// Write renders the clause using dollar-prefixed placeholders.
func (PostgresDriver) Write(clause SqlClause, argPosition int) (string, int, error) {
	return writeClause(clause, argPosition, dollarPlaceholder{})
}

type dollarPlaceholder struct{}

func (dollarPlaceholder) Placeholder(idx int) string { return fmt.Sprintf("$%d", idx) }
