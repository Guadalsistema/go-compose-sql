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
	PrimaryKey bool
	NotNull    bool
	Unique     bool
	AutoIncr   bool
	DefaultVal interface{}
	ForeignKey *ForeignKeyRef
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
