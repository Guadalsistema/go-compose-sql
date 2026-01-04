# go-compose-sql
Transform Go structs into simple SQL queries.

This project uses [github.com/kisielk/sqlstruct](https://pkg.go.dev/github.com/kisielk/sqlstruct) to map struct fields to database columns when executing queries.

## Future plans
 - Coalesce need a redesign
 - Allow configure how the name to table translations (Take in account the posibilityfor define a interface with methods)

## Drivers

Statements are rendered by a `Driver` implementation. If no driver is provided, `SQLiteDriver{}` is used, which emits `?` placeholders and appends a semicolon. For PostgreSQL, pass `PostgresDriver{}` in `SqlOpts` to generate `$1`-style placeholders and omit the trailing semicolon:

```go
stmt := Select[User](&SqlOpts{Driver: PostgresDriver{}}).Where("id=?", 10)
sql, _ := stmt.Write()
// SELECT id, first_name FROM user WHERE id=$1
```
