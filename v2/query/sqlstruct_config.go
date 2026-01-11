package query

import "github.com/kisielk/sqlstruct"

func init() {
	sqlstruct.TagName = "sql"
	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
}
