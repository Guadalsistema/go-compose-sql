package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func demonstrateColumnTypes() {
	// Example with PostgreSQL
	fmt.Println("=== PostgreSQL ===")
	dbPg, _ := sql.Open("postgres", "postgres://user:pass@localhost/db")
	defer dbPg.Close()

	rows, err := dbPg.QueryContext(context.Background(),
		"SELECT id, name, created_at FROM users LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Get column types BEFORE scanning
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Fatal(err)
	}

	for i, ct := range columnTypes {
		fmt.Printf("Column %d: %s\n", i, ct.Name())
		fmt.Printf("  Database Type: %s\n", ct.DatabaseTypeName())
		fmt.Printf("  Go ScanType: %v\n", ct.ScanType())
		fmt.Printf("  Nullable: %v\n", ct.Nullable())
		fmt.Println()
	}

	// Output for PostgreSQL:
	// Column 0: id
	//   Database Type: INT8
	//   Go ScanType: int64
	//   Nullable: false
	//
	// Column 1: name
	//   Database Type: TEXT
	//   Go ScanType: string
	//   Nullable: false
	//
	// Column 2: created_at
	//   Database Type: TIMESTAMP
	//   Go ScanType: time.Time
	//   Nullable: false

	fmt.Println("\n=== SQLite ===")
	dbSqlite, _ := sql.Open("sqlite3", ":memory:")
	defer dbSqlite.Close()

	dbSqlite.Exec("CREATE TABLE users (id INTEGER, name TEXT, created_at DATETIME)")
	rows2, _ := dbSqlite.Query("SELECT id, name, created_at FROM users LIMIT 1")
	defer rows2.Close()

	columnTypes2, _ := rows2.ColumnTypes()

	for i, ct := range columnTypes2 {
		fmt.Printf("Column %d: %s\n", i, ct.Name())
		fmt.Printf("  Database Type: %s\n", ct.DatabaseTypeName())
		fmt.Printf("  Go ScanType: %v\n", ct.ScanType())
		fmt.Println()
	}

	// Output for SQLite:
	// Column 0: id
	//   Database Type: INTEGER
	//   Go ScanType: int64
	//
	// Column 1: name
	//   Database Type: TEXT
	//   Go ScanType: string
	//
	// Column 2: created_at
	//   Database Type: DATETIME
	//   Go ScanType: string  <-- ⚠️ Returns STRING, not time.Time!
}

// This is the KEY insight for your V2 API
func smartScan(rows *sql.Rows, dest interface{}) error {
	// 1. Get column types from the database
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	// 2. Get expected types from your Table definition
	// (you know this from Column[T])
	expectedTypes := getExpectedTypesFromStruct(dest)

	// 3. Create smart scanners based on BOTH pieces of information
	scanTargets := make([]interface{}, len(columnTypes))

	for i, ct := range columnTypes {
		dbType := ct.DatabaseTypeName()      // e.g., "DATETIME"
		scanType := ct.ScanType()            // e.g., reflect.TypeOf(string)
		expectedType := expectedTypes[i]      // e.g., reflect.TypeOf(time.Time{})

		// Decision logic:
		if expectedType == timeType && scanType == stringType {
			// Database returns string, but user expects time.Time
			// → Use a converter scanner
			scanTargets[i] = &timeConverter{}
		} else {
			// Types match, scan directly
			scanTargets[i] = reflect.New(scanType).Interface()
		}
	}

	return rows.Scan(scanTargets...)
}

type timeConverter struct {
	result time.Time
}

func (tc *timeConverter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		tc.result = v
	case string:
		// SQLite returns ISO8601 string
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		tc.result = parsed
	case int64:
		tc.result = time.Unix(v, 0)
	default:
		return fmt.Errorf("cannot convert %T to time.Time", value)
	}

	return nil
}

func getExpectedTypesFromStruct(dest interface{}) []reflect.Type {
	// Implementation would use reflection to extract types
	// from your struct fields
	return nil
}
