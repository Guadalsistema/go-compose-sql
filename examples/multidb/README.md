# Multi-Database Example

This example demonstrates using go-compose-sql with multiple database drivers (SQLite and PostgreSQL).

## Setup

1. **Install dependencies:**

```bash
go mod init example-multidb
go get github.com/guadalsistema/go-compose-sql/v2
go get modernc.org/sqlite
go get github.com/lib/pq
```

2. **Run with SQLite:**

```bash
go run main.go
```

3. **Run with PostgreSQL:**

First, start a PostgreSQL instance:

```bash
docker run -d \
  --name postgres-example \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:15
```

Then run:

```bash
DATABASE_URL="postgresql://postgres:password@localhost/testdb?sslmode=disable" go run main.go
```

## What This Example Shows

✅ **Single table definition** works with both databases
✅ **Automatic timestamp conversion** (SQLite string → time.Time)
✅ **Same query code** for different databases
✅ **Environment-based** database selection
✅ **Type-safe queries** with compile-time checking

## Key Features Demonstrated

1. **Type Conversion**: `time.Time` and `sql.NullTime` work seamlessly
2. **Query Building**: Type-safe WHERE, ORDER BY, LIMIT
3. **Insert/Update**: RETURNING clause (PostgreSQL and SQLite 3.35+)
4. **Transactions**: Begin/Commit/Rollback
5. **Multiple Drivers**: Switch between databases without code changes
