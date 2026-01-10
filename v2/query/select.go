package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/expr"
)

// SelectBuilder builds SELECT queries
type SelectBuilder struct {
	session    ConnectionInterface
	table      interface{}
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
	Table     interface{}
	Condition expr.Expr
}

// OrderByClause represents an ORDER BY clause
type OrderByClause struct {
	Column    string
	Direction string // "ASC" or "DESC"
}

// NewSelect creates a new SELECT builder
func NewSelect(session ConnectionInterface, table interface{}) *SelectBuilder {
	return &SelectBuilder{
		session: session,
		table:   table,
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
func (b *SelectBuilder) Join(table interface{}, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "INNER JOIN",
		Table:     table,
		Condition: condition,
	})
	return b
}

// LeftJoin adds a LEFT JOIN
func (b *SelectBuilder) LeftJoin(table interface{}, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "LEFT JOIN",
		Table:     table,
		Condition: condition,
	})
	return b
}

// RightJoin adds a RIGHT JOIN
func (b *SelectBuilder) RightJoin(table interface{}, condition expr.Expr) *SelectBuilder {
	b.joins = append(b.joins, &JoinClause{
		Type:      "RIGHT JOIN",
		Table:     table,
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
	tableName := b.session.GetTableName(b.table)
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString(" FROM ")
	sql.WriteString(tableName)

	// JOINs
	for _, join := range b.joins {
		joinTableName := b.session.GetTableName(join.Table)
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
func (b *SelectBuilder) All(dest interface{}) error {
	sqlStr, args, err := b.ToSQL()
	if err != nil {
		return err
	}

	// Replace placeholders based on driver
	sqlStr = b.replacePlaceholders(sqlStr, args)

	rows, err := b.session.QueryRows(sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column types from database
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types: %w", err)
	}

	// Get expected types from table definition
	expectedTypes, err := b.getExpectedTypes()
	if err != nil {
		return fmt.Errorf("failed to get expected types: %w", err)
	}

	// Ensure we have the same number of expected types as columns
	if len(expectedTypes) != len(columnTypes) {
		return fmt.Errorf("column count mismatch: expected %d, got %d", len(expectedTypes), len(columnTypes))
	}

	// Get type registry from dialect
	registry := b.session.Engine().Dialect().TypeRegistry()

	// Prepare destination slice
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be pointer to slice")
	}
	sliceValue := destValue.Elem()
	elemType := sliceValue.Type().Elem()

	// Scan all rows
	for rows.Next() {
		// Create scan targets with conversion support
		scanTargets := CreateScanTargets(columnTypes, expectedTypes, registry)

		// Scan the row
		err := rows.Scan(scanTargets...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Extract values from scanners
		values := ExtractValues(scanTargets)

		// Create new struct and populate fields
		newElem := reflect.New(elemType).Elem()
		for i, value := range values {
			if i >= newElem.NumField() {
				break
			}
			field := newElem.Field(i)
			if field.CanSet() {
				valueReflect := reflect.ValueOf(value)
				if valueReflect.Type().AssignableTo(field.Type()) {
					field.Set(valueReflect)
				}
			}
		}

		// Append to slice
		sliceValue.Set(reflect.Append(sliceValue, newElem))
	}

	return rows.Err()
}

// One executes the query and returns a single result
func (b *SelectBuilder) One(dest interface{}) error {
	sqlStr, args, err := b.ToSQL()
	if err != nil {
		return err
	}

	// Replace placeholders based on driver
	sqlStr = b.replacePlaceholders(sqlStr, args)

	// Use QueryRows instead of QueryRow to get column types
	rows, err := b.session.QueryRows(sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Check if there's a row
	if !rows.Next() {
		return sql.ErrNoRows
	}

	// Get column types from database
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types: %w", err)
	}

	// Get expected types from table definition
	expectedTypes, err := b.getExpectedTypes()
	if err != nil {
		return fmt.Errorf("failed to get expected types: %w", err)
	}

	// Ensure we have the same number of expected types as columns
	if len(expectedTypes) != len(columnTypes) {
		return fmt.Errorf("column count mismatch: expected %d, got %d", len(expectedTypes), len(columnTypes))
	}

	// Get type registry from dialect
	registry := b.session.Engine().Dialect().TypeRegistry()

	// Create scan targets with conversion support
	scanTargets := CreateScanTargets(columnTypes, expectedTypes, registry)

	// Scan the row
	err = rows.Scan(scanTargets...)
	if err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Extract values from scanners
	values := ExtractValues(scanTargets)

	// Populate dest struct
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	destValue = destValue.Elem()

	for i, value := range values {
		if i >= destValue.NumField() {
			break
		}
		field := destValue.Field(i)
		if field.CanSet() {
			valueReflect := reflect.ValueOf(value)
			if valueReflect.Type().AssignableTo(field.Type()) {
				field.Set(valueReflect)
			}
		}
	}

	return nil
}

// Count returns the count of matching rows
func (b *SelectBuilder) Count() (int64, error) {
	// Create a copy of the builder with COUNT(*)
	countBuilder := &SelectBuilder{
		session:    b.session,
		table:      b.table,
		columns:    []string{"COUNT(*) as count"},
		whereExprs: b.whereExprs,
		joins:      b.joins,
		groupBy:    b.groupBy,
		having:     b.having,
	}

	sql, args, err := countBuilder.ToSQL()
	if err != nil {
		return 0, err
	}

	sql = b.replacePlaceholders(sql, args)

	var count int64
	row := b.session.QueryRow(sql, args...)
	err = row.Scan(&count)
	return count, err
}

// replacePlaceholders converts ? placeholders to driver-specific format
func (b *SelectBuilder) replacePlaceholders(sql string, args []interface{}) string {
	driver := b.session.Engine().Dialect()
	position := 1
	result := ""

	for _, char := range sql {
		if char == '?' {
			result += driver.Placeholder(position)
			position++
		} else {
			result += string(char)
		}
	}

	return result
}

// getExpectedTypes extracts expected column types from the table definition
func (b *SelectBuilder) getExpectedTypes() ([]reflect.Type, error) {
	// Try to get column types from table using reflection
	tableValue := reflect.ValueOf(b.table)

	// Call ColumnTypes() method if available
	columnTypesMethod := tableValue.MethodByName("ColumnTypes")
	if !columnTypesMethod.IsValid() {
		return nil, fmt.Errorf("table does not have ColumnTypes() method")
	}

	results := columnTypesMethod.Call(nil)
	if len(results) == 0 {
		return nil, fmt.Errorf("ColumnTypes() returned no results")
	}

	// Convert result to []reflect.Type
	if types, ok := results[0].Interface().([]reflect.Type); ok {
		return types, nil
	}

	return nil, fmt.Errorf("ColumnTypes() did not return []reflect.Type")
}
