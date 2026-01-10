# Using go-compose-sql as a Dependency

This guide shows how to add `go-compose-sql` to your project with support for different database drivers.

## Basic Installation (No Drivers)

The library core can be used without any database drivers:

```bash
go get github.com/guadalsistema/go-compose-sql/v2
```

This gives you access to the V2 API, query builders, and type conversion system without pulling in any database drivers.

## Adding Database Drivers

Database drivers are **not** included by default. You need to explicitly add the drivers you want to use.

### Option 1: SQLite (Pure Go - No CGo)

Use `modernc.org/sqlite` for a pure Go SQLite driver (no CGo required):

```bash
go get github.com/guadalsistema/go-compose-sql/v2
go get modernc.org/sqlite
```

**Example usage:**

```go
package main

import (
    "context"
    "log"

    _ "modernc.org/sqlite"

    "github.com/guadalsistema/go-compose-sql/v2/engine"
    "github.com/guadalsistema/go-compose-sql/v2/table"
)

func main() {
    // Create SQLite engine
    eng, err := engine.NewEngine("sqlite://./mydb.db", engine.EngineOpts{})
    if err != nil {
        log.Fatal(err)
    }

    conn, err := eng.Connect(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Use the connection...
}
```

### Option 2: SQLite (CGo - Faster)

Use `mattn/go-sqlite3` for better performance (requires CGo):

```bash
go get github.com/guadalsistema/go-compose-sql/v2
go get github.com/mattn/go-sqlite3
```

**Example usage:**

```go
package main

import (
    "context"
    "log"

    _ "github.com/mattn/go-sqlite3"

    "github.com/guadalsistema/go-compose-sql/v2/engine"
)

func main() {
    eng, err := engine.NewEngine("sqlite3://./mydb.db", engine.EngineOpts{})
    if err != nil {
        log.Fatal(err)
    }

    // Use the engine...
}
```

### Option 3: PostgreSQL

Use `lib/pq` for PostgreSQL:

```bash
go get github.com/guadalsistema/go-compose-sql/v2
go get github.com/lib/pq
```

**Example usage:**

```go
package main

import (
    "context"
    "log"

    _ "github.com/lib/pq"

    "github.com/guadalsistema/go-compose-sql/v2/engine"
)

func main() {
    // PostgreSQL connection
    eng, err := engine.NewEngine(
        "postgresql://user:password@localhost/dbname?sslmode=disable",
        engine.EngineOpts{},
    )
    if err != nil {
        log.Fatal(err)
    }

    conn, err := eng.Connect(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Use the connection...
}
```

### Option 4: Multiple Drivers (SQLite + PostgreSQL)

You can support multiple databases in the same application:

```bash
go get github.com/guadalsistema/go-compose-sql/v2
go get modernc.org/sqlite
go get github.com/lib/pq
```

**Example go.mod:**

```go
module myapp

go 1.21

require (
    github.com/guadalsistema/go-compose-sql/v2 v2.0.0
    github.com/lib/pq v1.10.9
    modernc.org/sqlite v1.29.1
)
```

**Example usage:**

```go
package main

import (
    "context"
    "log"
    "os"

    _ "github.com/lib/pq"
    _ "modernc.org/sqlite"

    "github.com/guadalsistema/go-compose-sql/v2/engine"
    "github.com/guadalsistema/go-compose-sql/v2/table"
)

// Define your table once
type User struct {
    ID    int64
    Name  string
    Email string
}

type UsersColumns struct {
    ID    *table.Column[int64]
    Name  *table.Column[string]
    Email *table.Column[string]
}

var Users = table.NewTable("users", UsersColumns{
    ID:    table.Col[int64]("id").PrimaryKey(),
    Name:  table.Col[string]("name").NotNull(),
    Email: table.Col[string]("email").NotNull(),
})

func main() {
    // Choose database based on environment
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        // Default to SQLite
        dbURL = "sqlite://./app.db"
    }

    eng, err := engine.NewEngine(dbURL, engine.EngineOpts{})
    if err != nil {
        log.Fatal(err)
    }

    conn, err := eng.Connect(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Same code works with both SQLite and PostgreSQL!
    var users []User
    err = conn.Query(Users).All(&users)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d users", len(users))
}
```

## Complete Example Project

Here's a complete example project structure:

```
myapp/
├── go.mod
├── go.sum
├── main.go
└── models/
    └── user.go
```

**go.mod:**

```go
module github.com/yourusername/myapp

go 1.21

require (
    github.com/guadalsistema/go-compose-sql/v2 v2.0.0
    github.com/lib/pq v1.10.9
    modernc.org/sqlite v1.29.1
)
```

**models/user.go:**

```go
package models

import (
    "database/sql"
    "time"

    "github.com/guadalsistema/go-compose-sql/v2/table"
)

type User struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt sql.NullTime
}

type UsersColumns struct {
    ID        *table.Column[int64]
    Name      *table.Column[string]
    Email     *table.Column[string]
    CreatedAt *table.Column[time.Time]
    UpdatedAt *table.Column[sql.NullTime]
}

var Users = table.NewTable("users", UsersColumns{
    ID:        table.Col[int64]("id").PrimaryKey().AutoIncrement(),
    Name:      table.Col[string]("name").NotNull(),
    Email:     table.Col[string]("email").Unique().NotNull(),
    CreatedAt: table.Col[time.Time]("created_at").NotNull(),
    UpdatedAt: table.Col[sql.NullTime]("updated_at"),
})
```

**main.go:**

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "time"

    _ "github.com/lib/pq"        // PostgreSQL driver
    _ "modernc.org/sqlite"       // SQLite driver

    "github.com/guadalsistema/go-compose-sql/v2/engine"
    "github.com/guadalsistema/go-compose-sql/v2/expr"
    "github.com/yourusername/myapp/models"
)

func main() {
    // Connect to database (supports both SQLite and PostgreSQL)
    eng, err := engine.NewEngine("sqlite://./myapp.db", engine.EngineOpts{})
    if err != nil {
        log.Fatal(err)
    }

    conn, err := eng.Connect(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Insert a user
    _, err = conn.Insert(models.Users).
        Set("name", "Alice").
        Set("email", "alice@example.com").
        Set("created_at", time.Now()).
        Exec()
    if err != nil {
        log.Fatal(err)
    }

    // Query users
    var users []models.User
    err = conn.Query(models.Users).
        Where(expr.Like(models.Users.C.Email, "%@example.com")).
        OrderBy("created_at").
        All(&users)
    if err != nil {
        log.Fatal(err)
    }

    for _, user := range users {
        log.Printf("User: %s <%s>", user.Name, user.Email)
    }
}
```

## Build and Run

```bash
# Initialize module
go mod init github.com/yourusername/myapp

# Add dependencies
go get github.com/guadalsistema/go-compose-sql/v2
go get modernc.org/sqlite
go get github.com/lib/pq

# Build
go build

# Run
./myapp
```

## Switching Between Databases

The same code works with different databases by just changing the connection string:

```go
// SQLite (file)
eng, _ := engine.NewEngine("sqlite://./myapp.db", engine.EngineOpts{})

// SQLite (in-memory)
eng, _ := engine.NewEngine("sqlite://:memory:", engine.EngineOpts{})

// PostgreSQL
eng, _ := engine.NewEngine(
    "postgresql://user:pass@localhost/mydb?sslmode=disable",
    engine.EngineOpts{},
)

// MySQL (requires github.com/go-sql-driver/mysql)
eng, _ := engine.NewEngine(
    "mysql://user:pass@tcp(localhost:3306)/mydb?parseTime=true",
    engine.EngineOpts{},
)
```

## Key Features

### ✅ Automatic Type Conversion

The library automatically handles type differences between databases:

```go
type User struct {
    CreatedAt time.Time      // Works with both!
    UpdatedAt sql.NullTime   // Works with both!
}

// PostgreSQL: Returns time.Time natively
// SQLite: Returns string, automatically converted to time.Time
// Same struct, same code, different databases!
```

### ✅ No CGo Required (with modernc.org/sqlite)

```bash
# Pure Go - works everywhere
go get modernc.org/sqlite

# No need for:
# - gcc/clang
# - Platform-specific compilation
# - CGO_ENABLED=1
```

### ✅ Type-Safe Queries

```go
// Type-safe column references
conn.Query(Users).Where(expr.Eq(Users.C.Email, "alice@example.com"))

// Compile-time type checking
var users []User  // Must match table definition
conn.Query(Users).All(&users)
```

## Testing with Different Databases

You can test against multiple databases in CI/CD:

**Test with SQLite:**
```bash
go test ./...
```

**Test with PostgreSQL:**
```bash
export DATABASE_URL="postgresql://localhost/test?sslmode=disable"
go test ./...
```

## Migration from Other ORMs

### From GORM:

```go
// GORM
db.Where("email LIKE ?", "%@example.com").Find(&users)

// go-compose-sql
conn.Query(Users).Where(expr.Like(Users.C.Email, "%@example.com")).All(&users)
```

### From sqlx:

```go
// sqlx
db.Select(&users, "SELECT * FROM users WHERE email LIKE ?", "%@example.com")

// go-compose-sql
conn.Query(Users).Where(expr.Like(Users.C.Email, "%@example.com")).All(&users)
```

## Summary

1. **Install core library**: `go get github.com/guadalsistema/go-compose-sql/v2`
2. **Add drivers you need**: SQLite (modernc.org/sqlite or mattn/go-sqlite3), PostgreSQL (lib/pq), or MySQL (go-sql-driver/mysql)
3. **Import drivers with `_`**: `import _ "modernc.org/sqlite"`
4. **Use engine.NewEngine()** with appropriate connection string
5. **Same code works across databases** thanks to automatic type conversion

The library is designed to be database-agnostic while handling the subtle differences between database engines automatically.
