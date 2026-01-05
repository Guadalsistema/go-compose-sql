package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/mysql"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/postgres"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/sqlite"
)

// Engine manages database configuration and connections.
type Engine struct {
	dialect dialect.Dialect
	config  EngineOpts
	info    *connectionInfo
}

// EngineOpts holds engine configuration.
// Logger is optional and can be used by higher layers to trace SQL statements.
type EngineOpts struct {
	Logger     *slog.Logger
	Autocommit bool
}

// NewEngine creates a new database engine from a SQLAlchemy-style connection URL,
// e.g. "sqlite+pysqlite:///:memory:" or "postgresql+psycopg2://user:pass@host/db".
// It opens the underlying database with sql.Open and selects the dialect driver
// (placeholder/quoting behaviour) based on the URL scheme.
func NewEngine(connectionURL string, opts EngineOpts) (*Engine, error) {
	parsed, err := parseConnectionURL(connectionURL)
	if err != nil {
		return nil, err
	}

	dialectDriver, err := dialectForScheme(parsed.dialect)
	if err != nil {
		return nil, err
	}

	return &Engine{
		dialect: dialectDriver,
		config:  opts,
		info:    parsed,
	}, nil
}

// Dialect returns the configured SQL dialect (placeholder/quoting behaviour).
func (e *Engine) Dialect() dialect.Dialect {
	return e.dialect
}

// Logger returns the configured logger (may be nil).
func (e *Engine) Logger() *slog.Logger {
	return e.config.Logger
}

// Autocommit returns whether the engine defaults to autocommit connections.
func (e *Engine) Autocommit() bool {
	return e.config.Autocommit
}

// ConnectionInfo returns the parsed connection information for the engine.
func (e *Engine) ConnectionInfo() *connectionInfo {
	return e.info
}

// Connect creates a new database connection using the engine configuration.
func (e *Engine) Connect(ctx context.Context) (*Connection, error) {
	db, err := sql.Open(e.info.sqlDriverName, e.info.dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		engine: e,
		db:     db,
		ctx:    ctx,
	}, nil
}

type connectionInfo struct {
	dialect       string
	driverHint    string
	sqlDriverName string
	dsn           string
}

// SQLDriverName returns the Go SQL driver name.
func (c *connectionInfo) SQLDriverName() string {
	return c.sqlDriverName
}

// DSN returns the driver-specific DSN.
func (c *connectionInfo) DSN() string {
	return c.dsn
}

func parseConnectionURL(raw string) (*connectionInfo, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid connection URL: %w", err)
	}

	baseScheme, driverHint := splitScheme(u.Scheme)
	if baseScheme == "" {
		return nil, fmt.Errorf("invalid connection URL: missing scheme")
	}
	if driverHint == "" {
		driverHint = defaultDriverHint(baseScheme)
	}

	sqlDriverName := goSQLDriverName(baseScheme, driverHint)
	if sqlDriverName == "" {
		return nil, fmt.Errorf("unsupported driver for scheme %q", baseScheme)
	}

	dsn, err := buildDSN(baseScheme, u)
	if err != nil {
		return nil, err
	}

	return &connectionInfo{
		dialect:       baseScheme,
		driverHint:    driverHint,
		sqlDriverName: sqlDriverName,
		dsn:           dsn,
	}, nil
}

func splitScheme(scheme string) (string, string) {
	parts := strings.SplitN(strings.ToLower(scheme), "+", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func defaultDriverHint(dialect string) string {
	switch strings.ToLower(dialect) {
	case "sqlite":
		return "pysqlite"
	case "postgres", "postgresql":
		return "psycopg2"
	case "mysql":
		return "pymysql"
	default:
		return ""
	}
}

func goSQLDriverName(dialect, driverHint string) string {
	switch strings.ToLower(driverHint) {
	case "pysqlite":
		return "sqlite3"
	case "psycopg2", "pg8000":
		return "postgres"
	case "pymysql":
		return "mysql"
	}

	switch strings.ToLower(dialect) {
	case "sqlite":
		return "sqlite3"
	case "postgres", "postgresql":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return ""
	}
}

func buildDSN(dialect string, u *url.URL) (string, error) {
	switch strings.ToLower(dialect) {
	case "sqlite":
		path := strings.TrimPrefix(u.Path, "/")
		if path == "" {
			path = ":memory:"
		}
		if u.Host != "" {
			path = strings.TrimPrefix(u.Host+"/"+path, "/")
		}
		if u.RawQuery != "" {
			return path + "?" + u.RawQuery, nil
		}
		return path, nil
	case "postgres", "postgresql":
		normalized := *u
		normalized.Scheme = "postgres"
		return normalized.String(), nil
	case "mysql":
		normalized := *u
		normalized.Scheme = "mysql"
		return normalized.String(), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", dialect)
	}
}

func dialectForScheme(scheme string) (dialect.Dialect, error) {
	switch strings.ToLower(scheme) {
	case "sqlite":
		return &sqlite.SQLiteDialect{}, nil
	case "postgres", "postgresql":
		return &postgres.PostgresDialect{}, nil
	case "mysql":
		return &mysql.MySQLDialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported dialect %q", scheme)
	}
}
