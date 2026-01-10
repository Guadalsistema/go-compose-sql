package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/guadalsistema/go-compose-sql/v2/dialect"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/postgres"
	"github.com/guadalsistema/go-compose-sql/v2/dialect/sqlite"
)

// This example demonstrates how the TypeRegistry handles timestamp conversions
// across different database dialects (PostgreSQL, SQLite, MySQL)

func main() {
	demonstratePostgreSQLTimestamps()
	demonstrateSQLiteTimestamps()
	demonstrateRegistryAPI()
}

func demonstratePostgreSQLTimestamps() {
	fmt.Println("=== PostgreSQL Timestamp Handling ===")

	pgDialect := postgres.NewPostgresDialect()
	registry := pgDialect.TypeRegistry()

	// PostgreSQL driver returns time.Time natively
	dbValue := time.Now()
	targetType := reflect.TypeOf(time.Time{})

	converted, err := registry.Convert(dbValue, targetType)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database value: %v (type: %T)\n", dbValue, dbValue)
	fmt.Printf("Converted value: %v (type: %T)\n", converted, converted)
	fmt.Println("✓ PostgreSQL handles time.Time natively, no conversion needed\n")
}

func demonstrateSQLiteTimestamps() {
	fmt.Println("=== SQLite Timestamp Handling ===")

	sqliteDialect := sqlite.NewSQLiteDialect()
	registry := sqliteDialect.TypeRegistry()

	// SQLite returns timestamps as strings
	dbValue := "2024-01-15 14:30:00"
	targetType := reflect.TypeOf(time.Time{})

	converted, err := registry.Convert(dbValue, targetType)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database value: %v (type: %T)\n", dbValue, dbValue)
	fmt.Printf("Converted value: %v (type: %T)\n", converted, converted)
	fmt.Println("✓ SQLite string automatically converted to time.Time\n")

	// Convert to sql.NullTime
	fmt.Println("--- Converting to sql.NullTime ---")
	nullTimeType := reflect.TypeOf(sql.NullTime{})

	convertedNull, err := registry.Convert(dbValue, nullTimeType)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database value: %v (type: %T)\n", dbValue, dbValue)
	fmt.Printf("Converted value: %v (type: %T)\n", convertedNull, convertedNull)
	fmt.Println("✓ SQLite string automatically converted to sql.NullTime\n")

	// Handle NULL values
	fmt.Println("--- Handling NULL values ---")
	var nilValue interface{} = nil

	convertedNil, err := registry.Convert(nilValue, nullTimeType)
	if err != nil {
		log.Fatal(err)
	}

	nullTime := convertedNil.(sql.NullTime)
	fmt.Printf("NULL value converted to: %+v\n", nullTime)
	fmt.Printf("Valid: %v\n", nullTime.Valid)
	fmt.Println("✓ NULL values handled correctly\n")
}

func demonstrateRegistryAPI() {
	fmt.Println("=== Type Registry API ===")

	// Create a custom registry
	sqliteDialect := sqlite.NewSQLiteDialect()
	registry := sqliteDialect.TypeRegistry()

	// Test different source types
	testCases := []struct {
		name       string
		source     interface{}
		targetType reflect.Type
	}{
		{
			name:       "String to time.Time",
			source:     "2024-01-15 10:30:00",
			targetType: reflect.TypeOf(time.Time{}),
		},
		{
			name:       "Unix timestamp (int64) to time.Time",
			source:     int64(1705318200),
			targetType: reflect.TypeOf(time.Time{}),
		},
		{
			name:       "Int64 to bool (SQLite 0/1)",
			source:     int64(1),
			targetType: reflect.TypeOf(true),
		},
		{
			name:       "Int64 to bool (false)",
			source:     int64(0),
			targetType: reflect.TypeOf(true),
		},
	}

	for _, tc := range testCases {
		fmt.Printf("Test: %s\n", tc.name)
		result, err := registry.Convert(tc.source, tc.targetType)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Source: %v (%T)\n", tc.source, tc.source)
			fmt.Printf("  Result: %v (%T)\n", result, result)
			fmt.Println("  ✓ Success")
		}
		fmt.Println()
	}
}

// Example of how this would be used in a query builder
func demonstrateQueryBuilderUsage() {
	fmt.Println("=== Query Builder Integration (Conceptual) ===")

	// This shows how the SelectBuilder would use the registry

	// 1. Execute query and get rows
	// rows, _ := db.Query("SELECT id, name, created_at FROM users")

	// 2. Get column types from database
	// columnTypes, _ := rows.ColumnTypes()

	// 3. Get expected types from table definition
	expectedTypes := []reflect.Type{
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(""),
		reflect.TypeOf(time.Time{}), // User expects time.Time
	}

	// 4. Determine if conversion is needed
	d, _ := dialect.DialectByName("sqlite")
	registry := d.TypeRegistry()

	// Simulate: database returns string for timestamp
	dbTypes := []reflect.Type{
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(""),
		reflect.TypeOf(""), // SQLite returns string!
	}

	for i, dbType := range dbTypes {
		expectedType := expectedTypes[i]
		needsConversion := registry.NeedsConversion(dbType, expectedType)

		fmt.Printf("Column %d: DB type=%v, Expected=%v, Conversion needed=%v\n",
			i, dbType, expectedType, needsConversion)

		if needsConversion {
			// Use converting scanner
			scanner := registry.CreateScanner(expectedType)
			fmt.Printf("  → Will use converting scanner: %T\n", scanner)
		} else {
			// Direct scan
			fmt.Printf("  → Will scan directly\n")
		}
	}

	fmt.Println("\n✓ Type registry enables transparent conversion in query builders")
}
