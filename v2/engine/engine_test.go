package engine

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/guadalsistema/go-compose-sql/v2/dialect/postgres"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/sqlite"
)

func TestNewEngineFromConnectionURL(t *testing.T) {
	registerTestDrivers()

	tests := []struct {
		name        string
		url         string
		expectedDrv interface{}
	}{
		{
			name:        "sqlite memory",
			url:         "sqlite+pysqlite:///:memory:",
			expectedDrv: &sqlite.SQLiteDialect{},
		},
		{
			name:        "postgres psycopg2",
			url:         "postgresql+psycopg2://scott:tiger@localhost:5432/mydatabase",
			expectedDrv: &postgres.PostgresDialect{},
		},
		{
			name:        "postgres pg8000",
			url:         "postgresql+pg8000://dbuser:kx%40jj5%2Fg@pghost10/appdb",
			expectedDrv: &postgres.PostgresDialect{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewEngine(tt.url, EngineConfig{})
			if err != nil {
				t.Fatalf("NewEngine(%q) error = %v", tt.url, err)
			}
			if eng.DB() == nil {
				t.Fatalf("NewEngine(%q) returned nil DB", tt.url)
			}

			switch tt.expectedDrv.(type) {
			case *sqlite.SQLiteDialect:
				if _, ok := eng.Dialect().(*sqlite.SQLiteDialect); !ok {
					t.Fatalf("expected SQLite dialect, got %T", eng.Dialect())
				}
			case *postgres.PostgresDialect:
				if _, ok := eng.Dialect().(*postgres.PostgresDialect); !ok {
					t.Fatalf("expected Postgres dialect, got %T", eng.Dialect())
				}
			default:
				t.Fatalf("unexpected driver type in test table %T", tt.expectedDrv)
			}

			if err := eng.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		})
	}
}

// registerTestDrivers ensures sql.Open can succeed without pulling real database drivers.
func registerTestDrivers() {
	registerDriverOnce("sqlite3")
	registerDriverOnce("postgres")
	registerDriverOnce("mysql")
}

func registerDriverOnce(name string) {
	for _, existing := range sql.Drivers() {
		if existing == name {
			return
		}
	}
	sql.Register(name, &noopDriver{})
}

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) { return &noopConn{}, nil }

type noopConn struct{}

func (c *noopConn) Prepare(string) (driver.Stmt, error) { return &noopStmt{}, nil }
func (c *noopConn) Close() error                        { return nil }
func (c *noopConn) Begin() (driver.Tx, error)           { return &noopTx{}, nil }
func (c *noopConn) Ping(context.Context) error          { return nil }

type noopStmt struct{}

func (s *noopStmt) Close() error                               { return nil }
func (s *noopStmt) NumInput() int                              { return -1 }
func (s *noopStmt) Exec([]driver.Value) (driver.Result, error) { return noopResult(0), nil }
func (s *noopStmt) Query([]driver.Value) (driver.Rows, error)  { return &noopRows{}, nil }
func (s *noopStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return noopResult(0), nil
}
func (s *noopStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	return &noopRows{}, nil
}

type noopTx struct{}

func (noopTx) Commit() error   { return nil }
func (noopTx) Rollback() error { return nil }

type noopRows struct{}

func (r *noopRows) Columns() []string              { return []string{} }
func (r *noopRows) Close() error                   { return nil }
func (r *noopRows) Next(dest []driver.Value) error { return io.EOF }

type noopResult int64

func (r noopResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r noopResult) RowsAffected() (int64, error) { return int64(r), nil }
