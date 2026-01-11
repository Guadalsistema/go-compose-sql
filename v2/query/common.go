package query

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
	"github.com/guadalsistema/go-compose-sql/v2/engine"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// ConnectionInterface defines the methods required by query builders
type ConnectionInterface interface {
	Engine() *engine.Engine
	Context() context.Context
	ExecuteContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRowsContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	GetTableName(tbl interface{}) string
	GetTableColumns(tbl interface{}) []*table.ColumnRef
}

// FormatPlaceholders converts ? placeholders to driver-specific format.
func FormatPlaceholders(sql string, dialect dialect.Dialect) string {
	position := 1
	var b strings.Builder
	b.Grow(len(sql))
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' {
			b.WriteString(dialect.Placeholder(position))
			position++
			continue
		}
		b.WriteByte(sql[i])
	}
	return b.String()
}

func logSQLTransform(logger *slog.Logger, rawSQL string, formattedSQL string, args []interface{}) {
	if logger == nil {
		return
	}
	if rawSQL == formattedSQL {
		logger.Debug("sqlcompose: sql built", "sql", formattedSQL, "args_len", len(args))
		return
	}
	logger.Debug("sqlcompose: sql placeholders formatted", "raw_sql", rawSQL, "sql", formattedSQL, "args_len", len(args))
}
