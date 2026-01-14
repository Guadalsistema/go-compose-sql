package builder

import (
	"context"
	"fmt"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// SelectBuilder builds SELECT queries
type SelectBuilder struct {
	conn       ConnectionInterface
	table      table.TableInterface
	columns    []string
	whereExprs []expr.Expr
	joins      []*JoinClause
	orderBy    []OrderByClause
	groupBy    []string
	having     []expr.Expr
	limit      *int
	offset     *int
	distinct   bool
}

// JoinClause represents a JOIN operation
type JoinClause struct {
	Type      string // "INNER", "LEFT", "RIGHT", "FULL"
	Table     table.TableInterface
	Condition expr.Expr
}

// OrderByClause represents an ORDER BY clause
type OrderByClause struct {
	Column    string
	Direction string // "ASC" or "DESC"
}

// NewSelect creates a new SELECT builder
func NewSelect(conn ConnectionInterface, tbl table.TableInterface) *SelectBuilder {
	return &SelectBuilder{
		conn:  conn,
		table: tbl,
	}
}

// Select specifies which columns to select (defaults to all)
func (b *SelectBuilder) Select(columns ...string) *SelectBuilder {
	b.columns = columns
	return b
}

// Where adds a WHERE condition
func (b *SelectBuilder) Where(condition expr.Expr) *SelectBuilder {
	b.whereExprs = append(b.whereExprs, condition)
	return b
}

// Join adds an INNER JOIN
func (b *SelectBuilder) Join(tbl table.TableInterface, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "INNER JOIN",
		Table:     tbl,
		Condition: condition,
	})
	return b
}

// LeftJoin adds a LEFT JOIN
func (b *SelectBuilder) LeftJoin(tbl table.TableInterface, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "LEFT JOIN",
		Table:     tbl,
		Condition: condition,
	})
	return b
}

// RightJoin adds a RIGHT JOIN
func (b *SelectBuilder) RightJoin(tbl table.TableInterface, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "RIGHT JOIN",
		Table:     tbl,
		Condition: condition,
	})
	return b
}

// OrderBy adds an ORDER BY clause (default ASC)
func (b *SelectBuilder) OrderBy(column string) *SelectBuilder {
	b.orderBy = append(b.orderBy, OrderByClause{
		Column:    column,
		Direction: "ASC",
	})
	return b
}

// OrderByDesc adds an ORDER BY DESC clause
func (b *SelectBuilder) OrderByDesc(column string) *SelectBuilder {
	b.orderBy = append(b.orderBy, OrderByClause{
		Column:    column,
		Direction: "DESC",
	})
	return b
}

// GroupBy adds a GROUP BY clause
func (b *SelectBuilder) GroupBy(columns ...string) *SelectBuilder {
	b.groupBy = append(b.groupBy, columns...)
	return b
}

// Having adds a HAVING condition
func (b *SelectBuilder) Having(condition expr.Expr) *SelectBuilder {
	b.having = append(b.having, condition)
	return b
}

// Limit sets the LIMIT
func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
	b.limit = &limit
	return b
}

// Offset sets the OFFSET
func (b *SelectBuilder) Offset(offset int) *SelectBuilder {
	b.offset = &offset
	return b
}

// Distinct enables DISTINCT
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.distinct = true
	return b
}

// ToSQL generates the SQL query and arguments
func (b *SelectBuilder) ToSQL() (string, []interface{}, error) {
	var sql strings.Builder
	var args []interface{}

	// SELECT [DISTINCT]
	sql.WriteString("SELECT")
	if b.distinct {
		sql.WriteString(" DISTINCT")
	}
	sql.WriteString(" ")

	// Columns
	if len(b.columns) > 0 {
		sql.WriteString(strings.Join(b.columns, ", "))
	} else {
		sql.WriteString("*")
	}

	// FROM
	tableName := b.table.Name()
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString(" FROM ")
	sql.WriteString(tableName)

	// JOINs
	for _, join := range b.joins {
		joinTableName := join.Table.Name()
		sql.WriteString(" ")
		sql.WriteString(join.Type)
		sql.WriteString(" ")
		sql.WriteString(joinTableName)
		sql.WriteString(" ON ")

		joinSQL, joinArgs := join.Condition.ToSQL()
		sql.WriteString(joinSQL)
		args = append(args, joinArgs...)
	}

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

	// GROUP BY
	if len(b.groupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(b.groupBy, ", "))
	}

	// HAVING
	if len(b.having) > 0 {
		sql.WriteString(" HAVING ")
		for i, havingExpr := range b.having {
			if i > 0 {
				sql.WriteString(" AND ")
			}
			havingSQL, havingArgs := havingExpr.ToSQL()
			sql.WriteString(havingSQL)
			args = append(args, havingArgs...)
		}
	}

	// ORDER BY
	if len(b.orderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(b.orderBy))
		for i, order := range b.orderBy {
			orderParts[i] = order.Column + " " + order.Direction
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if b.limit != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *b.limit))
	}

	// OFFSET
	if b.offset != nil {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", *b.offset))
	}

	return sql.String(), args, nil
}

// All executes the query and returns all results
func (b *SelectBuilder) All(ctx context.Context, dest interface{}) error {
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

	return scanAll(rows, dest)
}

// One executes the query and returns a single result
func (b *SelectBuilder) One(ctx context.Context, dest interface{}) error {
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

// Count returns the count of matching rows
func (b *SelectBuilder) Count(ctx context.Context) (int64, error) {
	ctx = b.resolveContext(ctx)
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Create a copy of the builder with COUNT(*)
	countBuilder := &SelectBuilder{
		conn:       b.conn,
		table:      b.table,
		columns:    []string{"COUNT(*) as count"},
		whereExprs: b.whereExprs,
		joins:      b.joins,
		groupBy:    b.groupBy,
		having:     b.having,
	}

	sqlStr, args, err := countBuilder.ToSQL()
	if err != nil {
		return 0, err
	}

	rawSQL := sqlStr
	sqlStr = FormatPlaceholders(sqlStr, b.conn.Dialect())
	logSQLTransform(b.conn.Logger(), rawSQL, sqlStr, args)

	rows, err := b.conn.QueryRowsContext(ctx, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("no rows")
	}

	var count int64
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}
	if rows.Next() {
		return 0, fmt.Errorf("expected exactly one row")
	}
	return count, rows.Err()
}

func (b *SelectBuilder) resolveContext(ctx context.Context) context.Context {
	if ctx == nil {
		return b.conn.Context()
	}
	return ctx
}
