# go-compose-sql v2

A SQLAlchemy-inspired SQL query builder for Go with type-safe expressions and composable queries.

## Overview

Version 2 of go-compose-sql is a complete rewrite inspired by SQLAlchemy's query builder pattern. It provides:

- **Type-safe column expressions** - Define columns with type parameters for compile-time safety
- **Rich expression language** - Build complex WHERE clauses with `Eq`, `Gt`, `Like`, `In`, `Between`, etc.
- **Connection pattern** - Manage database connections and transactions elegantly
- **Composable queries** - Chain methods to build queries incrementally
- **Multiple SQL dialects** - Support for PostgreSQL, SQLite, and MySQL
- **Full SQL feature set** - GROUP BY, HAVING, DISTINCT, JOINs, and more

## Key Differences from v1

| Feature | v1 | v2 |
|---------|----|----|
| WHERE clauses | String-based, injection-prone | Type-safe expression language |
| Table definition | Implicit via generics | Explicit Table objects |
| Column access | String literals | Typed Column objects |
| Query building | Statement chaining | Connection + builder pattern |
| GROUP BY/HAVING | Not supported | ✅ Supported |
| JOIN types | Generic JOIN only | INNER, LEFT, RIGHT, FULL |
| Transactions | Manual | Connection-based with Begin/Commit |
| DISTINCT | Not supported | ✅ Supported |

## Installation

```bash
go get github.com/guadalsistema/go-compose-sql/v2
```

## Quick Start

### 1. Define Your Table

```go
import (
    "github.com/guadalsistema/go-compose-sql/v2/table"
)

// Define column structure
type UsersColumns struct {
    ID        *table.Column[int64]
    Name      *table.Column[string]
    Email     *table.Column[string]
    Age       *table.Column[int]
    CreatedAt *table.Column[time.Time]
}

// Create table definition
var Users = table.NewTable("users", UsersColumns{
    ID:        table.Col[int64]("id").PrimaryKey().AutoIncrement(),
    Name:      table.Col[string]("name").NotNull(),
    Email:     table.Col[string]("email").Unique().NotNull(),
    Age:       table.Col[int]("age"),
    CreatedAt: table.Col[time.Time]("created_at").NotNull(),
})
```

### 2. Create an Engine and Connection

```go
import (
    "context"

    "github.com/guadalsistema/go-compose-sql/v2/engine"
)

// Create engine from SQLAlchemy-style URL
eng, _ := engine.NewEngine(
    "postgresql+psycopg2://user:pass@localhost:5432/mydatabase",
    engine.EngineOpts{Autocommit: true},
)

// Create connection
conn, _ := eng.Connect(context.Background())
defer conn.Close()
```

### 3. Build and Execute Queries

```go
import "github.com/guadalsistema/go-compose-sql/v2/expr"

// SELECT with WHERE
query := conn.Query(Users).
    Where(expr.Eq(Users.C.ID, int64(1)))
// SQL: SELECT * FROM users WHERE users.id = $1

// Multiple conditions
query := conn.Query(Users).
    Where(expr.Gt(Users.C.Age, 18)).
    Where(expr.Like(Users.C.Email, "%@example.com")).
    OrderByDesc("created_at").
    Limit(10)
// SQL: SELECT * FROM users WHERE users.age > $1 AND users.email LIKE $2
//      ORDER BY created_at DESC LIMIT 10

// Complex OR conditions
query := conn.Query(Users).
    Where(expr.Or(
        expr.Eq(Users.C.Name, "John"),
        expr.Eq(Users.C.Name, "Jane"),
    ))
// SQL: SELECT * FROM users WHERE ((users.name = $1) OR (users.name = $2))

// INSERT
result, _ := conn.Insert(Users).
    Set("name", "John Doe").
    Set("email", "john@example.com").
    Set("age", 30).
    Exec()

// UPDATE
result, _ := conn.Update(Users).
    Set("age", 31).
    Where(expr.Eq(Users.C.ID, int64(1))).
    Exec()

// DELETE
result, _ := conn.Delete(Users).
    Where(expr.Lt(Users.C.Age, 18)).
    Exec()
```

## Expression Language

The v2 API provides a rich set of type-safe expressions:

### Comparison Operators

```go
expr.Eq(Users.C.Age, 25)           // age = 25
expr.Ne(Users.C.Age, 25)           // age != 25
expr.Lt(Users.C.Age, 25)           // age < 25
expr.Le(Users.C.Age, 25)           // age <= 25
expr.Gt(Users.C.Age, 25)           // age > 25
expr.Ge(Users.C.Age, 25)           // age >= 25
```

### NULL Checks

```go
expr.IsNull(Users.C.Email)         // email IS NULL
expr.IsNotNull(Users.C.Email)      // email IS NOT NULL
```

### IN Clauses

```go
expr.In(Users.C.ID, 1, 2, 3)       // id IN (1, 2, 3)
expr.NotIn(Users.C.ID, 1, 2, 3)    // id NOT IN (1, 2, 3)
```

### Pattern Matching

```go
expr.Like(Users.C.Email, "%@example.com")       // email LIKE '%@example.com'
expr.NotLike(Users.C.Email, "%spam%")           // email NOT LIKE '%spam%'
expr.ILike(Users.C.Email, "%@EXAMPLE.COM")      // email ILIKE '%@EXAMPLE.COM'
```

### Range Checks

```go
expr.Between(Users.C.Age, 18, 65)      // age BETWEEN 18 AND 65
expr.NotBetween(Users.C.Age, 0, 17)    // age NOT BETWEEN 0 AND 17
```

### Logical Operators

```go
expr.And(
    expr.Gt(Users.C.Age, 18),
    expr.Lt(Users.C.Age, 65),
)  // (age > 18 AND age < 65)

expr.Or(
    expr.Eq(Users.C.Status, "active"),
    expr.Eq(Users.C.Status, "pending"),
)  // (status = 'active' OR status = 'pending')
```

### Raw SQL

```go
expr.Raw("age * 2 > ?", 50)  // age * 2 > 50
```

## Advanced Features

### GROUP BY and HAVING

```go
query := sess.Query(Users).
    Select("age", "COUNT(*) as count").
    GroupBy("age").
    Having(expr.Raw("COUNT(*) > ?", 5))
// SQL: SELECT age, COUNT(*) as count FROM users
//      GROUP BY age HAVING COUNT(*) > $1
```

### JOINs

```go
// Define Orders table
var Orders = table.NewTable("orders", OrdersColumns{...})

// INNER JOIN
query := sess.Query(Users).
    Join(Orders, expr.Eq(Users.C.ID, Orders.C.UserID)).
    Where(expr.Gt(Orders.C.Total, 100))

// LEFT JOIN
query := sess.Query(Users).
    LeftJoin(Orders, expr.Eq(Users.C.ID, Orders.C.UserID))
```

### DISTINCT

```go
query := sess.Query(Users).
    Select("email").
    Distinct()
// SQL: SELECT DISTINCT email FROM users
```

### RETURNING Clause

```go
var user User
err := sess.Insert(Users).
    Set("name", "John").
    Set("email", "john@example.com").
    Returning("id", "created_at").
    One(&user)
// SQL: INSERT INTO users (name, email) VALUES ($1, $2)
//      RETURNING id, created_at
```

### Transactions

```go
// Begin transaction
tx, err := eng.Connect(context.Background())
if err != nil {
    log.Fatal(err)
}
if err := tx.Begin(); err != nil {
    log.Fatal(err)
}

// Perform operations
_, err = tx.Insert(Users).Set("name", "John").Exec()
if err != nil {
    tx.Rollback()
    return err
}

// Commit
err = tx.Commit()
```

## Supported Drivers

### PostgreSQL

```go
eng, _ := engine.NewEngine(
    "postgresql+psycopg2://user:pass@localhost:5432/mydatabase",
    engine.EngineOpts{Autocommit: true},
)
// Uses $1, $2, ... placeholders
// Supports RETURNING
```

### SQLite

```go
eng, _ := engine.NewEngine(
    "sqlite+pysqlite:///:memory:",
    engine.EngineOpts{Autocommit: true},
)
// Uses ? placeholders
// Supports RETURNING (SQLite 3.35.0+)
```

### MySQL

```go
eng, _ := engine.NewEngine(
    "mysql+pymysql://user:pass@localhost:3306/mydatabase",
    engine.EngineOpts{Autocommit: true},
)
// Uses ? placeholders
// Does not support RETURNING
```

## Architecture

### Package Structure

```
v2/
├── table/          # Table and Column definitions
├── expr/           # Expression language for WHERE/HAVING
├── query/          # Query builders (Select, Insert, Update, Delete)
├── engine/         # Engine and Connection implementations
└── examples/       # Usage examples
```

### Core Concepts

1. **Table** - Represents a database table with typed columns
2. **Column** - Type-safe column definition with metadata
3. **Expression** - SQL expression (WHERE, HAVING conditions)
4. **Connection** - Database connection/transaction context
5. **Builder** - Fluent query construction (Select, Insert, Update, Delete)
6. **Driver** - SQL dialect abstraction (Postgres, SQLite, MySQL)

## Migration from v1

The v2 API is a complete rewrite and not backwards compatible with v1. Key migration steps:

1. **Define tables explicitly** instead of using generic type parameters
2. **Replace string WHERE clauses** with expression language
3. **Use Connection pattern** instead of direct database access
4. **Update column references** from string literals to Column objects

### Before (v1)

```go
stmt := Select[User](nil).
    Where("age > ?", 18).
    OrderBy("name")

users, _ := Query[User](db, stmt)
```

### After (v2)

```go
// Define table once
var Users = table.NewTable("users", UsersColumns{
    Age:  table.Col[int]("age"),
    Name: table.Col[string]("name"),
    // ...
})

// Use connection
eng, _ := engine.NewEngine("sqlite+pysqlite:///:memory:", engine.EngineOpts{Autocommit: true})
conn, _ := eng.Connect(context.Background())
defer conn.Close()

query := conn.Query(Users).
    Where(expr.Gt(Users.C.Age, 18)).
    OrderBy("name")

var users []User
query.All(&users)
```

## Roadmap

- [ ] Implement struct scanning (currently uses TODO placeholders)
- [ ] Add aggregate functions (COUNT, SUM, AVG, MAX, MIN)
- [ ] Support for subqueries in SELECT/WHERE
- [ ] UPSERT support (ON CONFLICT / ON DUPLICATE KEY)
- [ ] Schema migration tools
- [ ] Query result caching
- [ ] Relationship mapping (like SQLAlchemy's relationships)

## Contributing

Contributions are welcome! Please see CONTRIBUTING.md for guidelines.

## License

MIT License - see LICENSE file for details
