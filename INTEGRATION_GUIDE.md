# TypeRegistry Integration Guide

This guide explains how to integrate the TypeRegistry into the V2 query builders for automatic type conversion.

## Overview

The TypeRegistry system handles type conversions between database types and Go types transparently. This is essential for supporting multiple database dialects that return different types for the same SQL type (e.g., SQLite returns timestamps as strings, PostgreSQL returns time.Time).

## Architecture

```
User Code (Column[time.Time])
    ↓
SelectBuilder.All() / .One()
    ↓
rows.ColumnTypes() → Get what DB returns
    ↓
table.Columns() → Get what user expects
    ↓
TypeRegistry.NeedsConversion() → Compare
    ↓
If needed: CreateScanner() → Converting Scanner
If not: Direct scan
    ↓
Scan values → Populate struct
```

## Integration Steps

### Step 1: Enhance Table to Store Column Type Information

Currently, `Column[T]` stores the type, but we need to expose it:

```go
// v2/table/column.go
import "reflect"

// Add method to get the Go type
func (c *Column[T]) Type() reflect.Type {
    var zero T
    return reflect.TypeOf(zero)
}
```

### Step 2: Create Smart Scanning Logic

```go
// v2/query/scanner.go (new file)
package query

import (
    "database/sql"
    "reflect"

    "github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// ScanTarget represents a target for scanning a column value
type ScanTarget struct {
    scanner    sql.Scanner
    targetType reflect.Type
    index      int
}

// CreateScanTargets creates scan targets for all columns
func CreateScanTargets(
    columnTypes []*sql.ColumnType,
    expectedTypes []reflect.Type,
    registry *typeconv.Registry,
) []interface{} {

    targets := make([]interface{}, len(columnTypes))

    for i, ct := range columnTypes {
        dbType := ct.ScanType()
        expectedType := expectedTypes[i]

        // Check if conversion is needed
        if registry.NeedsConversion(dbType, expectedType) {
            // Use converting scanner
            targets[i] = registry.CreateScanner(expectedType)
        } else {
            // Direct scan - create pointer to expected type
            targets[i] = reflect.New(expectedType).Interface()
        }
    }

    return targets
}

// ExtractValues extracts scanned values from scan targets
func ExtractValues(targets []interface{}) []interface{} {
    values := make([]interface{}, len(targets))

    for i, target := range targets {
        // Check if it's a converting scanner
        if cs, ok := target.(*typeconv.ConvertingScanner); ok {
            values[i] = cs.Result()
        } else {
            // Extract value using reflection
            values[i] = reflect.ValueOf(target).Elem().Interface()
        }
    }

    return values
}
```

### Step 3: Update SelectBuilder.All()

```go
// v2/query/select.go

import (
    "reflect"
    "github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// All executes the query and returns all results
func (b *SelectBuilder) All(dest interface{}) error {
    sql, args, err := b.ToSQL()
    if err != nil {
        return err
    }

    // Replace placeholders based on driver
    sql = b.replacePlaceholders(sql, args)

    rows, err := b.session.QueryRows(sql, args...)
    if err != nil {
        return err
    }
    defer rows.Close()

    // Get column types from database
    columnTypes, err := rows.ColumnTypes()
    if err != nil {
        return err
    }

    // Get expected types from table definition
    expectedTypes, err := b.getExpectedTypes()
    if err != nil {
        return err
    }

    // Get type registry from dialect
    registry := b.session.Engine().Dialect().TypeRegistry()

    // Create scan targets with conversion support
    scanTargets := CreateScanTargets(columnTypes, expectedTypes, registry)

    // Prepare slice for results
    destValue := reflect.ValueOf(dest)
    if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
        return fmt.Errorf("dest must be pointer to slice")
    }
    sliceValue := destValue.Elem()
    elemType := sliceValue.Type().Elem()

    // Scan all rows
    for rows.Next() {
        err := rows.Scan(scanTargets...)
        if err != nil {
            return err
        }

        // Extract values from scanners
        values := ExtractValues(scanTargets)

        // Create new struct and populate fields
        newElem := reflect.New(elemType).Elem()
        for i, value := range values {
            field := newElem.Field(i)
            field.Set(reflect.ValueOf(value))
        }

        // Append to slice
        sliceValue.Set(reflect.Append(sliceValue, newElem))
    }

    return rows.Err()
}

// getExpectedTypes extracts expected types from table definition
func (b *SelectBuilder) getExpectedTypes() ([]reflect.Type, error) {
    // This needs to extract Column[T] types from the table
    // Implementation depends on your table structure

    // Example pseudocode:
    // table := b.table.(TableWithColumns)
    // columns := table.Columns()
    // types := make([]reflect.Type, len(columns))
    // for i, col := range columns {
    //     types[i] = col.Type()
    // }
    // return types, nil

    return nil, fmt.Errorf("not implemented")
}
```

### Step 4: Update SelectBuilder.One()

```go
// One executes the query and returns a single result
func (b *SelectBuilder) One(dest interface{}) error {
    sql, args, err := b.ToSQL()
    if err != nil {
        return err
    }

    // Replace placeholders based on driver
    sql = b.replacePlaceholders(sql, args)

    rows, err := b.session.QueryRows(sql, args...)
    if err != nil {
        return err
    }
    defer rows.Close()

    if !rows.Next() {
        return sql.ErrNoRows
    }

    // Get column types from database
    columnTypes, err := rows.ColumnTypes()
    if err != nil {
        return err
    }

    // Get expected types from table definition
    expectedTypes, err := b.getExpectedTypes()
    if err != nil {
        return err
    }

    // Get type registry from dialect
    registry := b.session.Engine().Dialect().TypeRegistry()

    // Create scan targets with conversion support
    scanTargets := CreateScanTargets(columnTypes, expectedTypes, registry)

    // Scan row
    err = rows.Scan(scanTargets...)
    if err != nil {
        return err
    }

    // Extract values
    values := ExtractValues(scanTargets)

    // Populate dest struct
    destValue := reflect.ValueOf(dest)
    if destValue.Kind() != reflect.Ptr {
        return fmt.Errorf("dest must be a pointer")
    }
    destValue = destValue.Elem()

    for i, value := range values {
        field := destValue.Field(i)
        field.Set(reflect.ValueOf(value))
    }

    return nil
}
```

## Usage Example

Once integrated, users can write code like this:

```go
type User struct {
    ID        int64
    Name      string
    CreatedAt time.Time  // Works with both PostgreSQL and SQLite!
}

// PostgreSQL - driver returns time.Time directly
connPG, _ := enginePG.Connect(ctx)
var usersPG []User
connPG.Query(Users).All(&usersPG)  // ✓ Works

// SQLite - driver returns string, registry converts to time.Time
connSQLite, _ := engineSQLite.Connect(ctx)
var usersSQLite []User
connSQLite.Query(Users).All(&usersSQLite)  // ✓ Also works!

// Same code, different dialects, transparent conversion
```

## Testing the Integration

```go
func TestSQLiteTimestampConversion(t *testing.T) {
    // Setup SQLite database
    db, _ := sql.Open("sqlite3", ":memory:")
    defer db.Close()

    db.Exec("CREATE TABLE users (id INTEGER, created_at DATETIME)")
    db.Exec("INSERT INTO users VALUES (1, '2024-01-15 10:30:00')")

    // Query using V2 API
    type User struct {
        ID        int64
        CreatedAt time.Time
    }

    var users []User
    conn.Query(Users).All(&users)

    // Verify conversion worked
    assert.Equal(t, 2024, users[0].CreatedAt.Year())
    assert.Equal(t, time.January, users[0].CreatedAt.Month())
}
```

## Advanced: Custom Type Converters

Users can register custom converters for their own types:

```go
// Custom type
type CustomTimestamp struct {
    time.Time
    Timezone string
}

// Register custom converter
dialect := sqlite.NewSQLiteDialect()
registry := dialect.TypeRegistry()

registry.Register(
    reflect.TypeOf(""),
    reflect.TypeOf(CustomTimestamp{}),
    func(source interface{}) (interface{}, error) {
        s := source.(string)
        t, err := time.Parse("2006-01-02 15:04:05", s)
        if err != nil {
            return CustomTimestamp{}, err
        }
        return CustomTimestamp{Time: t, Timezone: "UTC"}, nil
    },
)
```

## Performance Considerations

1. **Type Registry Lookup**: O(1) - uses map lookup
2. **Reflection overhead**: Minimal - only used during scanning
3. **Memory**: One converter per type pair (typically < 100 converters)
4. **No runtime penalty**: If types match, no conversion is performed

## Migration Guide

For existing code using the V2 API:

1. **No breaking changes**: Existing code continues to work
2. **Automatic conversion**: New dialects automatically use converters
3. **Optional**: Users can register custom converters if needed
4. **Backward compatible**: Old constructors still work with lazy initialization

## Summary

The TypeRegistry integration provides:
- ✓ Transparent type conversion across dialects
- ✓ No m×n if-else complexity
- ✓ Standard Go types work everywhere
- ✓ Extensible for custom types
- ✓ Clean, testable architecture
- ✓ Minimal performance overhead

The integration is complete in the dialect layer. The remaining work is to update the query builders to use `CreateScanTargets()` and `ExtractValues()` when scanning rows.
