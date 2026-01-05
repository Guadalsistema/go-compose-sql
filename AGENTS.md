## Project Overview

go-compose-sql is a Go library that transforms Go structs into SQL queries using a type-safe, composable API. The library uses reflection to map struct fields to database columns via the `github.com/kisielk/sqlstruct` package.

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -func=coverage.out | tail -n 1

# Run a single test
go test -run TestName ./...
```

### Building
```bash
# Build the project
go build -v ./...
```

### Code Formatting
```bash
# Format all Go files before committing (required)
gofmt -w .
```

### Dependency Management
```bash
# Tidy dependencies after modifications
go mod tidy
```

## Architecture

### Core Components

The library is structured around three main abstractions:

1. **SQLStatement** (`compose.go`): Represents a complete SQL statement as a sequence of clauses. Provides builder methods (`.Where()`, `.OrderBy()`, `.Limit()`, etc.) for chaining clauses together.

2. **SqlClause** (`clause.go`): Represents individual SQL components (SELECT, INSERT, WHERE, ORDER BY, etc.). Each clause stores its type, table name, column names, and the reflected ModelType from the generic type parameter.

3. **Query/Exec Functions** (`query.go`, `exec.go`): Execute statements against a database and map results back to structs using reflection.

### Type Reflection System

The library uses Go generics and reflection to maintain type safety:

- Generic functions `Insert[T]`, `Select[T]`, and `Delete[T]` capture the type `T` at compile time
- The reflected `ModelType` is stored in each clause to preserve type information
- At execution time, `sqlstruct.Scan()` maps database columns back to struct fields
- Field-to-column mapping: struct fields use the `db` tag; if absent, field names are converted to snake_case

### Statement Composition

Statements are built by:
1. Starting with a base clause: `Insert[User](opts)`, `Select[User](opts)`, or `Delete[User](opts)`
2. Chaining modifier methods: `.Where()`, `.OrderBy()`, `.Limit()`, `.Returning()`, etc.
3. Executing with `Exec()` or `Query()` functions

The `Write()` method validates clause ordering (e.g., DESC/ASC must follow ORDER BY, RETURNING must follow INSERT/UPDATE/DELETE) and concatenates clauses into the final SQL string.

### Special Clause Handling

- **RETURNING**: Only valid after INSERT, UPDATE, or DELETE clauses. Requires using `Query()` instead of `Exec()` to retrieve returned values.
- **COALESCE**: Injected into the SELECT column list rather than appended as a separate clause. Must follow a SELECT clause.
- **DESC/ASC**: Must immediately follow an ORDER BY clause.
- **OFFSET**: Should appear at the end of the statement (validation enforced).

### Driver Differences

The `SqlOpts.Driver` field accepts a `Driver` implementation (defaults to `SQLiteDriver{}`):
- `PostgresDriver{}`: Uses `$1`-style placeholders and omits the trailing semicolon
- `SQLiteDriver{}` (default): Uses `?` placeholders and appends a semicolon

### Query Iteration

`Query()` returns a `QueryRowIterator[T]` for efficient row-by-row scanning:
- `.Next()` advances to the next row
- `.Scan(dest)` maps the current row to a struct
- `.Close()` releases resources (should be deferred)

`QueryOne()` is a convenience wrapper that expects exactly one row and returns it directly.
