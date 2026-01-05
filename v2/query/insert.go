package query

import (
	"fmt"
	"strings"
)

// InsertBuilder builds INSERT queries
type InsertBuilder struct {
	session   ConnectionInterface
	table     interface{}
	values    []map[string]interface{} // Column-value pairs for each row
	returning []string
}

// NewInsert creates a new INSERT builder
func NewInsert(session ConnectionInterface, table interface{}) *InsertBuilder {
	return &InsertBuilder{
		session: session,
		table:   table,
	}
}

// Values adds values to insert (can be called multiple times for batch insert)
func (b *InsertBuilder) Values(data interface{}) *InsertBuilder {
	// TODO: Use reflection to extract column-value pairs from struct
	// For now, this is a placeholder
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
	if len(b.values) == 0 {
		return "", nil, fmt.Errorf("no values to insert")
	}

	var sql strings.Builder
	var args []interface{}

	// INSERT INTO table_name
	tableName := b.session.GetTableName(b.table)
	if tableName == "" {
		return "", nil, fmt.Errorf("invalid table")
	}
	sql.WriteString("INSERT INTO ")
	sql.WriteString(tableName)

	// Get column names from first row
	var columns []string
	for col := range b.values[0] {
		columns = append(columns, col)
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
			args = append(args, row[col])
		}
		sql.WriteString(")")
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

// Exec executes the INSERT and returns the result
func (b *InsertBuilder) Exec() (interface{}, error) {
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

	// Regular insert
	result, err := b.session.Execute(sql, args...)
	return result, err
}

// One executes the INSERT with RETURNING and scans into dest
func (b *InsertBuilder) One(dest interface{}) error {
	if len(b.returning) == 0 {
		return fmt.Errorf("RETURNING clause required for One()")
	}

	sql, args, err := b.ToSQL()
	if err != nil {
		return err
	}

	sql = replacePlaceholders(sql, args, b.session.Engine().Dialect())

	row := b.session.QueryRow(sql, args...)

	// TODO: Scan row into dest using reflection/sqlstruct
	_ = row
	return nil
}
