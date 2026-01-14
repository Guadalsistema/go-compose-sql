package builder

import (
	"strings"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
)

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
