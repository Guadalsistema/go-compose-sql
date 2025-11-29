package sqlcompose

import (
	"fmt"
	"reflect"
	"strings"
)

// ClauseType represents a SQL operation like INSERT or UPDATE.
type ClauseType string

const (
	ClauseInsert   ClauseType = "INSERT"
	ClauseSelect   ClauseType = "SELECT"
	ClauseUpdate   ClauseType = "UPDATE"
	ClauseDelete   ClauseType = "DELETE"
	ClauseWhere    ClauseType = "WHERE"
	ClauseOrderBy  ClauseType = "ORDER BY"
	ClauseLimit    ClauseType = "LIMIT"
	ClauseOffset   ClauseType = "OFFSET"
	ClauseCoalesce ClauseType = "COALESCE"
	ClauseDesc     ClauseType = "DESC"
	ClauseAsc      ClauseType = "ASC"
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
	switch c.Type {
	case ClauseInsert:
		cols := strings.Join(c.ColumnNames, ", ")
		placeholders := strings.TrimRight(strings.Repeat("?, ", len(c.ColumnNames)), ", ")
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", c.TableName, cols, placeholders), nil
	case ClauseSelect:
		cols := strings.Join(c.ColumnNames, ", ")
		return fmt.Sprintf("SELECT %s FROM %s", cols, c.TableName), nil
	case ClauseDelete:
		return fmt.Sprintf("DELETE FROM %s", c.TableName), nil
	case ClauseWhere:
		return fmt.Sprintf("WHERE %s", c.Expr), nil
	case ClauseOrderBy:
		cols := strings.Join(c.ColumnNames, ", ")
		return fmt.Sprintf("ORDER BY %s", cols), nil
	case ClauseLimit:
		return "LIMIT ?", nil
	case ClauseOffset:
		return "OFFSET ?", nil
	case ClauseCoalesce:
		if len(c.ColumnNames) < 2 {
			return "", NewErrInvalidCoalesceArgs(len(c.ColumnNames))
		}
		return fmt.Sprintf("COALESCE(%s)", strings.Join(c.ColumnNames, ", ")), nil
	case ClauseDesc:
		return "DESC", nil
	case ClauseAsc:
		return "ASC", nil
	default:
		return "", NewErrInvalidClause(string(c.Type))
	}
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
