package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// UpdateBuilder builds UPDATE queries
type UpdateBuilder struct {
	conn       ConnectionInterface
	table      table.TableInterface
	sets       map[string]interface{} // Column-value pairs to update
	whereExprs []expr.Expr
	returning  []string
}

// NewUpdate creates a new UPDATE builder
func NewUpdate(conn ConnectionInterface, tbl table.TableInterface) *UpdateBuilder {
	return &UpdateBuilder{
		conn:  conn,
		table: tbl,
		sets:  make(map[string]interface{}),
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
		if !b.conn.Dialect().SupportsReturning() {
			return "", nil, fmt.Errorf("driver does not support RETURNING clause")
		}
		sql.WriteString(" RETURNING ")
		sql.WriteString(strings.Join(b.returning, ", "))
	}

	return sql.String(), args, nil
}

// Exec executes the UPDATE and returns the result
func (b *UpdateBuilder) Exec(ctx context.Context) (sql.Result, error) {
	if len(b.returning) > 0 {
		return nil, fmt.Errorf("Exec cannot be used with RETURNING clause")
	}
	ctx = b.resolveContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sqlStr, args, err := b.ToSQL()
	if err != nil {
		return nil, err
	}

	rawSQL := sqlStr
	sqlStr = FormatPlaceholders(sqlStr, b.conn.Dialect())
	logSQLTransform(b.conn.Logger(), rawSQL, sqlStr, args)

	// Regular update
	return b.conn.ExecuteContext(ctx, sqlStr, args...)
}

// One executes the UPDATE with RETURNING and scans into dest
func (b *UpdateBuilder) One(ctx context.Context, dest interface{}) error {
	if len(b.returning) == 0 {
		return fmt.Errorf("RETURNING clause required for One()")
	}
	ctx = b.resolveContext(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}

	sqlStr, args, err := b.ToSQL()
	if err != nil {
		return err
	}

	rawSQL := sqlStr
	sqlStr = FormatPlaceholders(sqlStr, b.conn.Dialect())
	logSQLTransform(b.conn.Logger(), rawSQL, sqlStr, args)

	rows, err := b.conn.QueryRowsContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanOne(rows, dest)
}

func (b *UpdateBuilder) resolveContext(ctx context.Context) context.Context {
	if ctx == nil {
		return b.conn.Context()
	}
	return ctx
}
