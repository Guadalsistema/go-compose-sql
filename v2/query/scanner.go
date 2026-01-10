package query

import (
	"database/sql"
	"reflect"

	"github.com/guadalsistema/go-compose-sql/v2/typeconv"
)

// CreateScanTargets creates scan targets for all columns with automatic type conversion
// It compares what the database returns (from columnTypes) with what the user expects
// (from expectedTypes) and creates appropriate scanners using the dialect's TypeRegistry.
func CreateScanTargets(
	columnTypes []*sql.ColumnType,
	expectedTypes []reflect.Type,
	registry *typeconv.Registry,
) []interface{} {
	targets := make([]interface{}, len(columnTypes))

	for i, ct := range columnTypes {
		dbType := ct.ScanType()
		expectedType := expectedTypes[i]

		// Handle nil check for databases that don't provide ScanType
		if dbType == nil {
			// Fallback: create a pointer to expected type
			targets[i] = reflect.New(expectedType).Interface()
			continue
		}

		// Check if conversion is needed
		if registry.NeedsConversion(dbType, expectedType) {
			// Use converting scanner from registry
			targets[i] = registry.CreateScanner(expectedType)
		} else {
			// Direct scan - create pointer to expected type
			targets[i] = reflect.New(expectedType).Interface()
		}
	}

	return targets
}

// ExtractValues extracts scanned values from scan targets
// Handles both regular pointers and converting scanners
func ExtractValues(targets []interface{}) []interface{} {
	values := make([]interface{}, len(targets))

	for i, target := range targets {
		// Check if it's a converting scanner
		if scanner, ok := target.(interface{ Result() interface{} }); ok {
			// Extract result from converting scanner
			values[i] = scanner.Result()
		} else {
			// Extract value using reflection (regular pointer)
			targetValue := reflect.ValueOf(target)
			if targetValue.Kind() == reflect.Ptr && !targetValue.IsNil() {
				values[i] = targetValue.Elem().Interface()
			} else {
				values[i] = nil
			}
		}
	}

	return values
}
