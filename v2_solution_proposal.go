package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

// PROPOSAL FOR V2 API

// 1. Add a converter interface that knows about dialect differences
type TypeConverter interface {
	// NeedsConversion checks if conversion is needed
	// dbType: what the database returns (from ColumnTypes().ScanType())
	// targetType: what the user expects (from Column[T])
	NeedsConversion(dbType, targetType reflect.Type) bool

	// CreateScanner creates a scanner that can handle the conversion
	CreateScanner(targetType reflect.Type) sql.Scanner
}

// 2. Implement converter for each dialect

type PostgreSQLConverter struct{}

func (c *PostgreSQLConverter) NeedsConversion(dbType, targetType reflect.Type) bool {
	// PostgreSQL driver handles time.Time natively
	return false
}

func (c *PostgreSQLConverter) CreateScanner(targetType reflect.Type) sql.Scanner {
	// No conversion needed, return standard scanner
	return nil
}

type SQLiteConverter struct{}

func (c *SQLiteConverter) NeedsConversion(dbType, targetType reflect.Type) bool {
	// Check if database returns string but user expects time.Time
	if dbType.Kind() == reflect.String {
		if targetType == reflect.TypeOf(time.Time{}) {
			return true
		}
		if targetType == reflect.TypeOf(sql.NullTime{}) {
			return true
		}
	}
	return false
}

func (c *SQLiteConverter) CreateScanner(targetType reflect.Type) sql.Scanner {
	if targetType == reflect.TypeOf(time.Time{}) {
		return &TimeScanner{}
	}
	if targetType == reflect.TypeOf(sql.NullTime{}) {
		return &NullTimeScanner{}
	}
	return nil
}

// 3. Create smart scanners

type TimeScanner struct {
	Result time.Time
}

func (s *TimeScanner) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		s.Result = v
		return nil
	case string:
		// SQLite format
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		s.Result = parsed
		return nil
	case int64:
		s.Result = time.Unix(v, 0)
		return nil
	default:
		return fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

type NullTimeScanner struct {
	Result sql.NullTime
}

func (s *NullTimeScanner) Scan(value interface{}) error {
	if value == nil {
		s.Result = sql.NullTime{Valid: false}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		s.Result = sql.NullTime{Time: v, Valid: true}
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		s.Result = sql.NullTime{Time: parsed, Valid: true}
		return nil
	case int64:
		s.Result = sql.NullTime{Time: time.Unix(v, 0), Valid: true}
		return nil
	default:
		return fmt.Errorf("cannot convert %T to sql.NullTime", value)
	}
}

// 4. Usage in SelectBuilder.One() / .All()

func exampleQueryWithConversion(rows *sql.Rows, conn Connection) error {
	// Step 1: Get what the database will return
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	// Step 2: Get what the user expects (from table definition)
	// This comes from your Column[T] definitions
	expectedColumns := []ColumnInfo{
		{Name: "id", Type: reflect.TypeOf(int64(0))},
		{Name: "name", Type: reflect.TypeOf("")},
		{Name: "created_at", Type: reflect.TypeOf(time.Time{})},
	}

	// Step 3: Create scanners
	converter := conn.Engine().Dialect().GetConverter()
	scanTargets := make([]interface{}, len(columnTypes))

	for i, ct := range columnTypes {
		dbType := ct.ScanType()
		expectedType := expectedColumns[i].Type

		if converter.NeedsConversion(dbType, expectedType) {
			// Use smart scanner
			scanner := converter.CreateScanner(expectedType)
			scanTargets[i] = scanner
		} else {
			// Scan directly
			scanTargets[i] = reflect.New(expectedType).Interface()
		}
	}

	// Step 4: Scan with conversion
	if rows.Next() {
		err := rows.Scan(scanTargets...)
		if err != nil {
			return err
		}

		// Extract values from scanners
		for i, target := range scanTargets {
			switch scanner := target.(type) {
			case *TimeScanner:
				// Now you have scanner.Result as time.Time
				fmt.Println(scanner.Result)
			case *NullTimeScanner:
				// Now you have scanner.Result as sql.NullTime
				fmt.Println(scanner.Result)
			default:
				// Direct scan, extract with reflection
				val := reflect.ValueOf(target).Elem().Interface()
				fmt.Println(val)
			}
		}
	}

	return nil
}

type ColumnInfo struct {
	Name string
	Type reflect.Type
}

type Connection interface {
	Engine() Engine
}

type Engine interface {
	Dialect() Dialect
}

type Dialect interface {
	GetConverter() TypeConverter
}

// 5. The beauty of this approach:

func demonstrateTransparency() {
	// User code remains the same:
	type User struct {
		ID        int64
		Name      string
		CreatedAt time.Time // Works with both PostgreSQL and SQLite!
	}

	// The V2 API handles conversion transparently
	// - PostgreSQL: Direct scan
	// - SQLite: String → time.Time conversion

	// Users don't need custom types!
	// They can use standard time.Time and sql.NullTime
}
