package builder

import (
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// DeleteBuilder builds DELETE queries
type DeleteBuilder struct {
	dialect    dialect.Dialect
	table      table.TableInterface
	whereExprs []expr.Expr
	returning  []string
}

// NewDelete creates a new DELETE builder
func NewDelete(d dialect.Dialect, tbl table.TableInterface) *DeleteBuilder {
	return &DeleteBuilder{
		dialect: d,
		table:   tbl,
	}
}

// Where adds a WHERE condition
func (b *DeleteBuilder) Where(condition expr.Expr) *DeleteBuilder {
	b.whereExprs = append(b.whereExprs, condition)
	return b
}

// Returning specifies which columns to return
func (b *DeleteBuilder) Returning(columns ...string) *DeleteBuilder {
	b.returning = columns
	return b
}

// ToSQL generates the SQL query and arguments
func (b *DeleteBuilder) ToSQL() (string, []interface{}, error) {
	var sql strings.Builder
	var args []interface{}

	// DELETE FROM table_name
	tableName := b.table.Name()
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString("DELETE FROM ")
	sql.WriteString(tableName)

	// WHERE
	if len(b.whereExprs) > 0 {
		sql.WriteString(" WHERE ")
		for i, whereExpr := range b.whereExprs {
			if i > 0 {
				sql.WriteString(" AND ")
			}
			whereSQL, whereArgs := whereExpr.ToSQL()
			sql.WriteString(whereSQL)
			args = append(args, whereArgs...)
		}
	}

	// RETURNING
	if len(b.returning) > 0 {
		if !b.dialect.SupportsReturning() {
			return "", nil, fmt.Errorf("driver does not support RETURNING clause")
		}
		sql.WriteString(" RETURNING ")
		sql.WriteString(strings.Join(b.returning, ", "))
	}

	return sql.String(), args, nil
}
