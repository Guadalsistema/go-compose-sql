package engine

import (
	"context"
	"database/sql"

	"github.com/guadalsistema/go-compose-sql/v2/query"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// Connection represents a database connection/transaction context.
type Connection struct {
	engine *Engine
	db     *sql.DB
	ctx    context.Context
	tx     *sql.Tx
}

// Begin starts a transaction on the connection.
func (c *Connection) Begin() error {
	if c.tx != nil {
		return ErrAlreadyInTransaction
	}
	tx, err := c.db.BeginTx(c.ctx, nil)
	if err != nil {
		return err
	}
	c.tx = tx
	return nil
}

// ExecuteContext runs a SQL statement with the provided context.
func (c *Connection) ExecuteContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if ctx == nil {
		ctx = c.ctx
	}
	if c.tx != nil {
		return c.tx.ExecContext(ctx, query, args...)
	}
	return c.db.ExecContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row with the provided context.
func (c *Connection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if ctx == nil {
		ctx = c.ctx
	}
	if c.tx != nil {
		return c.tx.QueryRowContext(ctx, query, args...)
	}
	return c.db.QueryRowContext(ctx, query, args...)
}

// QueryRowsContext executes a query that returns multiple rows with the provided context.
func (c *Connection) QueryRowsContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if ctx == nil {
		ctx = c.ctx
	}
	if c.tx != nil {
		return c.tx.QueryContext(ctx, query, args...)
	}
	return c.db.QueryContext(ctx, query, args...)
}


// Commit commits the transaction.
func (c *Connection) Commit() error {
	if c.tx == nil {
		return ErrNotInTransaction
	}
	err := c.tx.Commit()
	c.tx = nil
	return err
}

// Rollback rolls back the transaction.
func (c *Connection) Rollback() error {
	if c.tx == nil {
		return ErrNotInTransaction
	}
	err := c.tx.Rollback()
	c.tx = nil
	return err
}

// Close closes the connection and rolls back if needed.
func (c *Connection) Close() error {
	if c.tx != nil {
		_ = c.Rollback()
	}
	return c.db.Close()
}

// Engine returns the underlying engine.
func (c *Connection) Engine() *Engine {
	return c.engine
}

// Context returns the connection context.
func (c *Connection) Context() context.Context {
	return c.ctx
}

// InTransaction returns true if the connection is in a transaction.
func (c *Connection) InTransaction() bool {
	return c.tx != nil
}

// Query creates a new SELECT query builder.
func (c *Connection) Query(tbl interface{}) *query.SelectBuilder {
	return query.NewSelect(c, tbl)
}

// Insert creates a new INSERT query builder.
func (c *Connection) Insert(tbl interface{}) *query.InsertBuilder {
	return query.NewInsert(c, tbl)
}

// Update creates a new UPDATE query builder.
func (c *Connection) Update(tbl interface{}) *query.UpdateBuilder {
	return query.NewUpdate(c, tbl)
}

// Delete creates a new DELETE query builder.
func (c *Connection) Delete(tbl interface{}) *query.DeleteBuilder {
	return query.NewDelete(c, tbl)
}

// GetTableName extracts the table name from a Table[T] object.
func (c *Connection) GetTableName(tbl interface{}) string {
	if t, ok := tbl.(interface{ Name() string }); ok {
		return t.Name()
	}
	return ""
}

// GetTableColumns extracts column references from a Table[T] object.
func (c *Connection) GetTableColumns(tbl interface{}) []*table.ColumnRef {
	if t, ok := tbl.(interface{ Columns() []*table.ColumnRef }); ok {
		return t.Columns()
	}
	return nil
}
