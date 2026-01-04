package sqlcompose

import (
	"fmt"
	"reflect"
)

// ClauseType represents a SQL operation like INSERT or UPDATE.
type ClauseType string

const (
	ClauseInsert    ClauseType = "INSERT"
	ClauseSelect    ClauseType = "SELECT"
	ClauseUpdate    ClauseType = "UPDATE"
	ClauseDelete    ClauseType = "DELETE"
	ClauseWhere     ClauseType = "WHERE"
	ClauseOrderBy   ClauseType = "ORDER BY"
	ClauseLimit     ClauseType = "LIMIT"
	ClauseOffset    ClauseType = "OFFSET"
	ClauseCoalesce  ClauseType = "COALESCE"
	ClauseReturning ClauseType = "RETURNING"
	ClauseDesc      ClauseType = "DESC"
	ClauseAsc       ClauseType = "ASC"
)

// SqlClause represents a SQL statement before rendering.
//
// ModelType retains the generic type used when building the clause so that
// values can later be mapped to columns when executing the statement.
type SqlClause struct {
	Type        ClauseType
	TableName   string
	ColumnNames []string
	ModelType   reflect.Type
	Expr        string
	Args        []any
}

// Write renders an individual SQL clause to a string.
func (c SqlClause) Write() (string, error) {
	sql, _, err := writeClause(c, 1, questionPlaceholder{})
	return sql, err
}

func formatCoalesceValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "NULL"
	case fmt.Stringer:
		return val.String()
	case string:
		return val
	default:
		return fmt.Sprint(val)
	}
}
