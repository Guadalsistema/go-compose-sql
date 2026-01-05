package session

import (
	"context"
	"database/sql"

	"github.com/guadalsistema/go-compose-sql/v2/engine"
	"github.com/guadalsistema/go-compose-sql/v2/query"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// Session represents a database session for executing queries
type Session struct {
	engine *engine.Engine
	ctx    context.Context
	tx     *sql.Tx // nil if not in a transaction
}

// NewSession creates a new session bound to the engine
func NewSession(ctx context.Context, eng *engine.Engine) *Session {
	return &Session{
		engine: eng,
		ctx:    ctx,
	}
}

// Engine returns the underlying engine
func (s *Session) Engine() *engine.Engine {
	return s.engine
}

// Context returns the session context
func (s *Session) Context() context.Context {
	return s.ctx
}

// Begin starts a transaction on the session
func (s *Session) Begin() error {
	if s.tx != nil {
		return ErrAlreadyInTransaction
	}
	tx, err := s.engine.DB().BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	s.tx = tx
	return nil
}

// Query creates a new SELECT query builder
func (s *Session) Query(tbl interface{}) *query.SelectBuilder {
	return query.NewSelect(s, tbl)
}

// Insert creates a new INSERT query builder
func (s *Session) Insert(tbl interface{}) *query.InsertBuilder {
	return query.NewInsert(s, tbl)
}

// Update creates a new UPDATE query builder
func (s *Session) Update(tbl interface{}) *query.UpdateBuilder {
	return query.NewUpdate(s, tbl)
}

// Delete creates a new DELETE query builder
func (s *Session) Delete(tbl interface{}) *query.DeleteBuilder {
	return query.NewDelete(s, tbl)
}

// Exec executes a raw SQL statement
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.ExecContext(s.ctx, query, args...)
	}
	return s.engine.DB().ExecContext(s.ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRowContext(s.ctx, query, args...)
	}
	return s.engine.DB().QueryRowContext(s.ctx, query, args...)
}

// QueryRows executes a query that returns multiple rows
func (s *Session) QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(s.ctx, query, args...)
	}
	return s.engine.DB().QueryContext(s.ctx, query, args...)
}

// Commit commits the transaction (only valid if session is in a transaction)
func (s *Session) Commit() error {
	if s.tx == nil {
		return ErrNotInTransaction
	}
	err := s.tx.Commit()
	s.tx = nil
	return err
}

// Rollback rolls back the transaction (only valid if session is in a transaction)
func (s *Session) Rollback() error {
	if s.tx == nil {
		return ErrNotInTransaction
	}
	err := s.tx.Rollback()
	s.tx = nil
	return err
}

// Close closes the session (rolls back transaction if active)
func (s *Session) Close() error {
	if s.tx != nil {
		return s.Rollback()
	}
	return nil
}

// InTransaction returns true if the session is in a transaction
func (s *Session) InTransaction() bool {
	return s.tx != nil
}

// GetTableName extracts the table name from a Table[T] object
func (s *Session) GetTableName(tbl interface{}) string {
	if t, ok := tbl.(interface{ Name() string }); ok {
		return t.Name()
	}
	return ""
}

// GetTableColumns extracts column references from a Table[T] object
func (s *Session) GetTableColumns(tbl interface{}) []*table.ColumnRef {
	if t, ok := tbl.(interface{ Columns() []*table.ColumnRef }); ok {
		return t.Columns()
	}
	return nil
}
