package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// This demonstrates the EXACT problem you're facing

func main() {
	testPostgreSQL()
	testSQLite()
}

func testPostgreSQL() {
	fmt.Println("=== PostgreSQL ===")
	db, err := sql.Open("postgres", "postgres://localhost/testdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE IF NOT EXISTS test_time (created_at TIMESTAMP)")
	db.Exec("INSERT INTO test_time VALUES (NOW())")

	rows, _ := db.Query("SELECT created_at FROM test_time")
	defer rows.Close()

	columnTypes, _ := rows.ColumnTypes()
	ct := columnTypes[0]

	fmt.Printf("Database Type: %s\n", ct.DatabaseTypeName()) // "TIMESTAMP"
	fmt.Printf("Go ScanType: %v\n", ct.ScanType())           // time.Time

	if rows.Next() {
		var t time.Time
		err := rows.Scan(&t) // ✅ Works! PostgreSQL returns time.Time
		fmt.Printf("Scan into time.Time: %v (error: %v)\n", t, err)

		rows.Next()
		var nt sql.NullTime
		rows.Scan(&nt) // ✅ Also works!
		fmt.Printf("Scan into sql.NullTime: %v (error: %v)\n", nt, err)
	}
}

func testSQLite() {
	fmt.Println("\n=== SQLite ===")
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Exec("CREATE TABLE test_time (created_at DATETIME)")
	db.Exec("INSERT INTO test_time VALUES (datetime('now'))")

	rows, _ := db.Query("SELECT created_at FROM test_time")
	defer rows.Close()

	columnTypes, _ := rows.ColumnTypes()
	ct := columnTypes[0]

	fmt.Printf("Database Type: %s\n", ct.DatabaseTypeName()) // "DATETIME"
	fmt.Printf("Go ScanType: %v\n", ct.ScanType())           // string ⚠️

	if rows.Next() {
		// Attempt 1: Scan into time.Time
		var t time.Time
		err := rows.Scan(&t) // ❌ ERROR: "sql: Scan error on column index 0, name \"created_at\": unsupported Scan, storing driver.Value type string into type *time.Time"
		fmt.Printf("Scan into time.Time: %v (error: %v)\n", t, err)
	}

	rows, _ = db.Query("SELECT created_at FROM test_time")
	defer rows.Close()

	if rows.Next() {
		// Attempt 2: Scan into string first, then parse
		var s string
		err := rows.Scan(&s) // ✅ Works! Returns "2024-01-10 12:34:56"
		fmt.Printf("Scan into string: %v (error: %v)\n", s, err)

		// Manual conversion
		t, err := time.Parse("2006-01-02 15:04:05", s)
		fmt.Printf("After parsing: %v (error: %v)\n", t, err)
	}

	rows, _ = db.Query("SELECT created_at FROM test_time")
	defer rows.Close()

	if rows.Next() {
		// Attempt 3: Scan into sql.NullTime
		var nt sql.NullTime
		err := rows.Scan(&nt) // ❌ ERROR: same problem as time.Time
		fmt.Printf("Scan into sql.NullTime: %v (error: %v)\n", nt, err)
	}
}

// THE SOLUTION: Create a smart scanner that knows about the mismatch

type SmartTimeScanner struct {
	Time  time.Time
	Valid bool
}

func (s *SmartTimeScanner) Scan(value interface{}) error {
	if value == nil {
		s.Valid = false
		return nil
	}

	// Handle different types the driver might return
	switch v := value.(type) {
	case time.Time:
		// PostgreSQL returns this
		s.Time = v
		s.Valid = true
		return nil

	case string:
		// SQLite returns this
		// Try multiple formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.999999Z",
			time.RFC3339,
		}

		var err error
		for _, format := range formats {
			s.Time, err = time.Parse(format, v)
			if err == nil {
				s.Valid = true
				return nil
			}
		}
		return fmt.Errorf("cannot parse time string: %s", v)

	case int64:
		// Unix timestamp
		s.Time = time.Unix(v, 0)
		s.Valid = true
		return nil

	default:
		return fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

func testSmartScanner() {
	fmt.Println("\n=== Smart Scanner Solution ===")

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	db.Exec("CREATE TABLE test_time (created_at DATETIME)")
	db.Exec("INSERT INTO test_time VALUES (datetime('now'))")

	rows, _ := db.Query("SELECT created_at FROM test_time")
	defer rows.Close()

	if rows.Next() {
		var scanner SmartTimeScanner
		err := rows.Scan(&scanner) // ✅ Works with both PostgreSQL and SQLite!
		fmt.Printf("Smart scan: %v (valid: %v, error: %v)\n", scanner.Time, scanner.Valid, err)
	}
}
