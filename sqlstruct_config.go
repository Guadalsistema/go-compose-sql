package sqlcompose

import "github.com/kisielk/sqlstruct"

// The init function sets up the sqlstruct package with custom configurations.
// It changes the struct field tag used for SQL mapping to "sql" and modifies
// the default name mapping function to convert struct field names to snake_case.
func init() {
	sqlstruct.TagName = "sql"
	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
}
