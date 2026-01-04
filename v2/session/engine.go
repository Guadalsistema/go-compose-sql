package session

import (
	"context"
	"database/sql"
)

// Driver represents a SQL dialect driver
type Driver interface {
	// Placeholder returns the placeholder format for this driver
	// e.g., "?" for SQLite/MySQL, "$" for Postgres
	Placeholder(position int) string

	// SupportsReturning indicates if the driver supports RETURNING clauses
	SupportsReturning() bool

	// Quote quotes an identifier (table/column name)
	Quote(identifier string) string
}

// Engine manages database connections and sessions
type Engine struct {
	db     *sql.DB
	driver Driver
	config EngineConfig
}

// EngineConfig holds engine configuration
type EngineConfig struct {
	Driver Driver
	Debug  bool // Log SQL statements
}

// EngineOption is a functional option for configuring an Engine
type EngineOption func(*EngineConfig)

// WithDriver sets the SQL driver
func WithDriver(driver Driver) EngineOption {
	return func(c *EngineConfig) {
		c.Driver = driver
	}
}

// WithDebug enables debug logging
func WithDebug(debug bool) EngineOption {
	return func(c *EngineConfig) {
		c.Debug = debug
	}
}

// NewEngine creates a new database engine
func NewEngine(db *sql.DB, opts ...EngineOption) *Engine {
	config := EngineConfig{
		Driver: &SQLiteDriver{}, // Default to SQLite
		Debug:  false,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &Engine{
		db:     db,
		driver: config.Driver,
		config: config,
	}
}

// DB returns the underlying database connection
func (e *Engine) DB() *sql.DB {
	return e.db
}

// Driver returns the SQL driver
func (e *Engine) Driver() Driver {
	return e.driver
}

// NewSession creates a new session for executing queries
func (e *Engine) NewSession() *Session {
	return &Session{
		engine: e,
		ctx:    context.Background(),
	}
}

// NewSessionWithContext creates a new session with a context
func (e *Engine) NewSessionWithContext(ctx context.Context) *Session {
	return &Session{
		engine: e,
		ctx:    ctx,
	}
}

// Begin starts a new transaction and returns a session bound to it
func (e *Engine) Begin() (*Session, error) {
	return e.BeginWithContext(context.Background())
}

// BeginWithContext starts a new transaction with a context
func (e *Engine) BeginWithContext(ctx context.Context) (*Session, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Session{
		engine: e,
		ctx:    ctx,
		tx:     tx,
	}, nil
}

// Close closes the database connection
func (e *Engine) Close() error {
	return e.db.Close()
}
