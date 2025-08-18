package sqlcompose

import "github.com/kisielk/sqlstruct"

func init() {
	sqlstruct.TagName = "db"
	sqlstruct.NameMapper = sqlstruct.ToSnakeCase
}
