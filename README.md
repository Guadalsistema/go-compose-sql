# go-compose-sql
Transform Go structs into simple SQL queries.

This project uses [github.com/kisielk/sqlstruct](https://pkg.go.dev/github.com/kisielk/sqlstruct) to map struct fields to database columns when executing queries.

## Future plans
 - Coalesce need a redesign
 - Allow configure how the name to table translations (Take in account the posibilityfor define a interface with methods)