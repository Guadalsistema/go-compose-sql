package sqlcompose

import (
	"fmt"
	"strings"
)

// Driver renders SQL clauses into dialect-specific strings.
// Implementations should render placeholders starting at the provided
// argument position and return how many placeholders were consumed.
type Driver interface {
	Write(SqlClause, int) (string, int, error)
}

type placeholderRenderer interface {
	Placeholder(int) string
}

func writeClause(clause SqlClause, argPosition int, placeholders placeholderRenderer) (string, int, error) {
	switch clause.Type {
	case ClauseInsert:
		cols := strings.Join(clause.ColumnNames, ", ")
		placeholdersList := make([]string, len(clause.ColumnNames))
		for i := range placeholdersList {
			placeholdersList[i] = placeholders.Placeholder(argPosition + i)
		}
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", clause.TableName, cols, strings.Join(placeholdersList, ", ")), len(placeholdersList), nil
	case ClauseSelect:
		cols := strings.Join(clause.ColumnNames, ", ")
		return fmt.Sprintf("SELECT %s FROM %s", cols, clause.TableName), 0, nil
	case ClauseUpdate:
		assignments := make([]string, len(clause.ColumnNames))
		for i, col := range clause.ColumnNames {
			assignments[i] = fmt.Sprintf("%s=%s", col, placeholders.Placeholder(argPosition+i))
		}
		return fmt.Sprintf("UPDATE %s SET %s", clause.TableName, strings.Join(assignments, ", ")), len(assignments), nil
	case ClauseDelete:
		return fmt.Sprintf("DELETE FROM %s", clause.TableName), 0, nil
	case ClauseWhere:
		expr, count := replacePlaceholders(clause.Expr, argPosition, placeholders)
		return fmt.Sprintf("WHERE %s", expr), count, nil
	case ClauseOrderBy:
		cols := strings.Join(clause.ColumnNames, ", ")
		return fmt.Sprintf("ORDER BY %s", cols), 0, nil
	case ClauseLimit:
		return fmt.Sprintf("LIMIT %s", placeholders.Placeholder(argPosition)), 1, nil
	case ClauseOffset:
		return fmt.Sprintf("OFFSET %s", placeholders.Placeholder(argPosition)), 1, nil
	case ClauseCoalesce:
		if len(clause.ColumnNames) < 2 {
			return "", 0, NewErrInvalidCoalesceArgs(len(clause.ColumnNames))
		}
		return fmt.Sprintf("COALESCE(%s)", strings.Join(clause.ColumnNames, ", ")), 0, nil
	case ClauseDesc:
		return "DESC", 0, nil
	case ClauseAsc:
		return "ASC", 0, nil
	case ClauseReturning:
		cols := "*"
		if len(clause.ColumnNames) > 0 {
			cols = strings.Join(clause.ColumnNames, ", ")
		}
		return fmt.Sprintf("RETURNING %s", cols), 0, nil
	default:
		return "", 0, NewErrInvalidClause(string(clause.Type))
	}
}

func replacePlaceholders(expr string, argPosition int, placeholders placeholderRenderer) (string, int) {
	var b strings.Builder
	count := 0
	for i := 0; i < len(expr); i++ {
		if expr[i] == '?' {
			b.WriteString(placeholders.Placeholder(argPosition + count))
			count++
			continue
		}
		b.WriteByte(expr[i])
	}
	if count == 0 {
		return expr, 0
	}
	return b.String(), count
}

// DriverByName returns a Driver instance matching the provided name.
// Recognized names: "postgres"/"postgresql" for PostgresDriver.
// Any other value (including empty) returns DefaultDriver.
func DriverByName(name string) (Driver, error) {
	switch strings.ToLower(name) {
	case "postgres", "postgresql":
		return PostgresDriver{}, nil
	case "sqlite", "sqlite3":
		return SQLiteDriver{}, nil
	default:
		return nil, fmt.Errorf("unknown driver name: %s", name)
	}
}
