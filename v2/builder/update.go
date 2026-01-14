package builder

import (
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// UpdateBuilder builds UPDATE queries
type UpdateBuilder struct {
	dialect    dialect.Dialect
	table      table.TableInterface
	sets       map[string]interface{} // Column-value pairs to update
	whereExprs []expr.Expr
	returning  []string
}

// NewUpdate creates a new UPDATE builder
func NewUpdate(d dialect.Dialect, tbl table.TableInterface) *UpdateBuilder {
	return &UpdateBuilder{
		dialect: d,
		table:   tbl,
		sets:    make(map[string]interface{}),
	}
}

// Set sets a column value
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.sets[column] = value
	return b
}

// Where adds a WHERE condition
func (b *UpdateBuilder) Where(condition expr.Expr) *UpdateBuilder {
	b.whereExprs = append(b.whereExprs, condition)
	return b
}

// Returning specifies which columns to return
func (b *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	b.returning = columns
	return b
}

// ToSQL generates the SQL query and arguments
func (b *UpdateBuilder) ToSQL() (string, []interface{}, error) {
	if len(b.sets) == 0 {
		return "", nil, fmt.Errorf("no columns to update")
	}

	var sql strings.Builder
	var args []interface{}

	// UPDATE table_name
	tableName := b.table.Name()
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString("UPDATE ")
	sql.WriteString(tableName)

	// SET column1 = ?, column2 = ?
	sql.WriteString(" SET ")
	setParts := make([]string, 0, len(b.sets))
	for col, val := range b.sets {
		setParts = append(setParts, col+" = ?")
		args = append(args, val)
	}
	sql.WriteString(strings.Join(setParts, ", "))

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
