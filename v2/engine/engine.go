package engine

import (
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

// Engine manages database connections and sessions
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
	config  EngineConfig
}

// EngineConfig holds engine configuration.
// Logger is optional and can be used by higher layers to trace SQL statements.
type EngineConfig struct {
	Logger *slog.Logger
}

// NewEngine creates a new database engine from a SQLAlchemy-style connection URL,
// e.g. "sqlite+pysqlite:///:memory:" or "postgresql+psycopg2://user:pass@host/db".
// It opens the underlying database with sql.Open and selects the dialect driver
// (placeholder/quoting behaviour) based on the URL scheme.
func NewEngine(connectionURL string, cfg EngineConfig) (*Engine, error) {
	parsed, err := parseConnectionURL(connectionURL)
	if err != nil {
		return nil, err
	}

	dialectDriver, err := dialectForScheme(parsed.dialect)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(parsed.sqlDriverName, parsed.dsn)
	if err != nil {
		return nil, err
	}

	return &Engine{
		db:      db,
		dialect: dialectDriver,
		config:  cfg,
	}, nil
}

// DB returns the underlying database connection.
func (e *Engine) DB() *sql.DB {
	return e.db
}

// Dialect returns the configured SQL dialect (placeholder/quoting behaviour).
func (e *Engine) Dialect() dialect.Dialect {
	return e.dialect
}

// Logger returns the configured logger (may be nil).
func (e *Engine) Logger() *slog.Logger {
	return e.config.Logger
}

// Close closes the database connection.
func (e *Engine) Close() error {
	return e.db.Close()
}

type connectionInfo struct {
	dialect       string
	driverHint    string
	sqlDriverName string
	dsn           string
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
