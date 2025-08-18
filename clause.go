package sqlcompose

import (
	"fmt"
	"reflect"
	"strings"
)

// ClauseType represents a SQL operation like INSERT or UPDATE.
type ClauseType string

const (
	ClauseInsert ClauseType = "INSERT"
	ClauseSelect ClauseType = "SELECT"
	ClauseUpdate ClauseType = "UPDATE"
	ClauseDelete ClauseType = "DELETE"
	ClauseWhere  ClauseType = "WHERE"
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
func (c SqlClause) Write() string {
	switch c.Type {
	case ClauseInsert:
		cols := strings.Join(c.ColumnNames, ", ")
		placeholders := strings.TrimRight(strings.Repeat("?, ", len(c.ColumnNames)), ", ")
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", c.TableName, cols, placeholders)
	case ClauseSelect:
		cols := strings.Join(c.ColumnNames, ", ")
		return fmt.Sprintf("SELECT %s FROM %s", cols, c.TableName)
	case ClauseDelete:
		return fmt.Sprintf("DELETE FROM %s", c.TableName)
	case ClauseWhere:
		return fmt.Sprintf("WHERE %s", c.Expr)
	default:
		return ""
	}
}
