package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// InsertBuilder builds INSERT queries
type InsertBuilder struct {
	conn      ConnectionInterface
	table     table.TableInterface
	values    []map[string]interface{} // Column-value pairs for each row
	returning []string
	err       error
}

// NewInsert creates a new INSERT builder
func NewInsert(conn ConnectionInterface, tbl table.TableInterface) *InsertBuilder {
	return &InsertBuilder{
		conn:  conn,
		table: tbl,
	}
}

// Values adds values to insert (can be called multiple times for batch insert)
func (b *InsertBuilder) Values(data interface{}) *InsertBuilder {
	if b.err != nil {
		return b
	}

	rows, err := normalizeInsertValues(data, b.table.Columns())
	if err != nil {
		b.err = err
		return b
	}
	b.values = append(b.values, rows...)
	return b
}

// Set sets a specific column value
func (b *InsertBuilder) Set(column string, value interface{}) *InsertBuilder {
	if len(b.values) == 0 {
		b.values = append(b.values, make(map[string]interface{}))
	}
	b.values[0][column] = value
	return b
}

// Returning specifies which columns to return
func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returning = columns
	return b
}

// ToSQL generates the SQL query and arguments
func (b *InsertBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if len(b.values) == 0 {
		return "", nil, fmt.Errorf("no values to insert")
	}

	var sql strings.Builder
	var args []interface{}

	// INSERT INTO table_name
	tableName := b.table.Name()
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString("INSERT INTO ")
	sql.WriteString(tableName)

	// Get column names from first row
	columns := orderedInsertColumns(b.values[0], b.table.Columns())
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no insertable columns found")
	}

	// (column1, column2, ...)
	sql.WriteString(" (")
	sql.WriteString(strings.Join(columns, ", "))
	sql.WriteString(")")

	// VALUES
	sql.WriteString(" VALUES ")

	// Add value rows
	for i, row := range b.values {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString("(")
		for j, col := range columns {
			if j > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString("?")
			val, ok := row[col]
			if ok {
				args = append(args, val)
			} else {
				args = append(args, nil)
			}
		}
		sql.WriteString(")")
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

// Exec executes the INSERT and returns the result
func (b *InsertBuilder) Exec(ctx context.Context) (sql.Result, error) {
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

	// Regular insert
	return b.conn.ExecuteContext(ctx, sqlStr, args...)
}

// One executes the INSERT with RETURNING and scans into dest
func (b *InsertBuilder) One(ctx context.Context, dest interface{}) error {
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

func (b *InsertBuilder) resolveContext(ctx context.Context) context.Context {
	if ctx == nil {
		return b.conn.Context()
	}
	return ctx
}
