package table

import "fmt"

// Column represents a database column with type safety
type Column[T any] struct {
	name        string
	tableName   string
	options     ColumnOptions
	parentTable interface{}
}

// ColumnOptions holds column metadata
type ColumnOptions struct {
	PrimaryKey             bool
	NotNull                bool
	Unique                 bool
	AutoIncr               bool
	DefaultVal             interface{}
	ForeignKey             *ForeignKeyRef
	CreatedAtTimestamp     bool // Automatically set timestamp on INSERT
	UpdatedAtTimestamp     bool // Automatically update timestamp on UPDATE
	DefaultCurrentTimestamp bool // Use database CURRENT_TIMESTAMP as default
}

// ForeignKeyRef represents a foreign key relationship
type ForeignKeyRef struct {
	Table  string
	Column string
}

// NewColumn creates a new column
func NewColumn[T any](name string) *Column[T] {
	return &Column[T]{
		name:    name,
		options: ColumnOptions{},
	}
}

// Col is a shorthand for NewColumn
func Col[T any](name string) *Column[T] {
	return NewColumn[T](name)
}

// Name returns the column name
func (c *Column[T]) Name() string {
	return c.name
}

// TableName returns the table name this column belongs to
func (c *Column[T]) TableName() string {
	return c.tableName
}

// FullName returns the fully qualified column name (table.column)
func (c *Column[T]) FullName() string {
	if c.tableName != "" {
		return fmt.Sprintf("%s.%s", c.tableName, c.name)
	}
	return c.name
}

// setTableName sets the parent table name (called during table initialization)
func (c *Column[T]) setTableName(tableName string) {
	c.tableName = tableName
}

// setParentTable sets the parent table reference
func (c *Column[T]) setParentTable(table interface{}) {
	c.parentTable = table
}

// Options returns the column options
func (c *Column[T]) Options() ColumnOptions {
	return c.options
}

// Builder methods for column options

// PrimaryKey marks this column as a primary key
func (c *Column[T]) PrimaryKey() *Column[T] {
	c.options.PrimaryKey = true
	return c
}

// NotNull marks this column as NOT NULL
func (c *Column[T]) NotNull() *Column[T] {
	c.options.NotNull = true
	return c
}

// Unique marks this column as UNIQUE
func (c *Column[T]) Unique() *Column[T] {
	c.options.Unique = true
	return c
}

// AutoIncrement marks this column as auto-incrementing
func (c *Column[T]) AutoIncrement() *Column[T] {
	c.options.AutoIncr = true
	return c
}

// Default sets a default value for this column
func (c *Column[T]) Default(val T) *Column[T] {
	c.options.DefaultVal = val
	return c
}

// ForeignKey sets a foreign key reference
func (c *Column[T]) ForeignKey(table, column string) *Column[T] {
	c.options.ForeignKey = &ForeignKeyRef{
		Table:  table,
		Column: column,
	}
	return c
}

// CreatedAtTimestamp marks this column to be automatically set to current timestamp on INSERT
func (c *Column[T]) CreatedAtTimestamp() *Column[T] {
	c.options.CreatedAtTimestamp = true
	c.options.NotNull = true // created_at should typically be NOT NULL
	return c
}

// UpdatedAtTimestamp marks this column to be automatically updated to current timestamp on UPDATE
func (c *Column[T]) UpdatedAtTimestamp() *Column[T] {
	c.options.UpdatedAtTimestamp = true
	c.options.CreatedAtTimestamp = true // updated_at should also be set on INSERT
	c.options.NotNull = true // updated_at should typically be NOT NULL
	return c
}

// CurrentTimestamp sets the database default value to CURRENT_TIMESTAMP
func (c *Column[T]) CurrentTimestamp() *Column[T] {
	c.options.DefaultCurrentTimestamp = true
	return c
}
