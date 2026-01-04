package query

import (
	"database/sql"

	"github.com/guadalsistema/go-compose-sql/v2/session"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// SessionInterface defines the methods required by query builders
type SessionInterface interface {
	Engine() *session.Engine
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRows(query string, args ...interface{}) (*sql.Rows, error)
	GetTableName(tbl interface{}) string
	GetTableColumns(tbl interface{}) []*table.ColumnRef
}

// replacePlaceholders converts ? placeholders to driver-specific format
func replacePlaceholders(sql string, args []interface{}, driver session.Driver) string {
	position := 1
	result := ""

	for _, char := range sql {
		if char == '?' {
			result += driver.Placeholder(position)
			position++
		} else {
			result += string(char)
		}
	}

	return result
}
