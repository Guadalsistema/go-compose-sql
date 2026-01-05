package query

import (
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/expr"
)

// DeleteBuilder builds DELETE queries
type DeleteBuilder struct {
	session    ConnectionInterface
	table      interface{}
	whereExprs []expr.Expr
	returning  []string
}

// NewDelete creates a new DELETE builder
func NewDelete(session ConnectionInterface, table interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		session: session,
		table:   table,
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
	tableName := b.session.GetTableName(b.table)
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
		if !b.session.Engine().Dialect().SupportsReturning() {
			return "", nil, fmt.Errorf("driver does not support RETURNING clause")
		}
		sql.WriteString(" RETURNING ")
		sql.WriteString(strings.Join(b.returning, ", "))
	}

	return sql.String(), args, nil
}

// Exec executes the DELETE and returns the result
func (b *DeleteBuilder) Exec() (interface{}, error) {
	sql, args, err := b.ToSQL()
	if err != nil {
		return nil, err
	}

	sql = replacePlaceholders(sql, args, b.session.Engine().Dialect())

	if len(b.returning) > 0 {
		// Use QueryRow for RETURNING
		row := b.session.QueryRow(sql, args...)
		// TODO: Scan the returned values
		_ = row
		return nil, nil
	}

	// Regular delete
	result, err := b.session.Execute(sql, args...)
	return result, err
}

// All executes the DELETE with RETURNING and returns all deleted rows
func (b *DeleteBuilder) All(dest interface{}) error {
	if len(b.returning) == 0 {
		return fmt.Errorf("RETURNING clause required for All()")
	}

	sql, args, err := b.ToSQL()
	if err != nil {
		return err
	}

	sql = replacePlaceholders(sql, args, b.session.Engine().Dialect())

	rows, err := b.session.QueryRows(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// TODO: Scan rows into dest using reflection/sqlstruct
	_ = rows
	return nil
}
