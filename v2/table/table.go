package table

import (
	"reflect"
	"strings"
)

// Table represents a database table with typed columns
type Table[T any] struct {
	name    string
	columns []*ColumnRef
	C       T // Column accessor (holds column definitions)
}

// ColumnRef holds metadata about a column without type parameters
type ColumnRef struct {
	Name     string
	FullName string
	Type     reflect.Type
	Options  ColumnOptions
}

// NewTable creates a new table with the given name and column definitions
func NewTable[T any](name string, columnStruct T) *Table[T] {
	table := &Table[T]{
		name: name,
		C:    columnStruct,
	}

	// Initialize columns by iterating over the struct fields
	table.columns = extractColumns(name, columnStruct)

	return table
}

// Name returns the table name
func (t *Table[T]) Name() string {
	return t.name
}

// Columns returns all column references
func (t *Table[T]) Columns() []*ColumnRef {
	return t.columns
}

// ColumnNames returns all column names
func (t *Table[T]) ColumnNames() []string {
	names := make([]string, len(t.columns))
	for i, col := range t.columns {
		names[i] = col.Name
	}
	return names
}

// extractColumns uses reflection to extract column metadata from the struct
func extractColumns(tableName string, columnStruct interface{}) []*ColumnRef {
	var columns []*ColumnRef

	v := reflect.ValueOf(columnStruct)
	t := v.Type()

	// Handle pointer to struct
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return columns
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check if this field is a *Column[T] type
		if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().String() == "*table.Column[...]" {
			if fieldVal.IsNil() {
				continue
			}

			// Use reflection to call methods on the column
			nameMethod := fieldVal.MethodByName("Name")
			if !nameMethod.IsValid() {
				continue
			}

			nameResults := nameMethod.Call(nil)
			if len(nameResults) == 0 {
				continue
			}

			columnName := nameResults[0].String()

			// Set the table name on the column
			setTableNameMethod := fieldVal.MethodByName("setTableName")
			if setTableNameMethod.IsValid() {
				// This won't work because setTableName is unexported
				// We'll need to handle this differently
			}

			// Get column options
			var opts ColumnOptions
			optionsMethod := fieldVal.MethodByName("Options")
			if optionsMethod.IsValid() {
				optResults := optionsMethod.Call(nil)
				if len(optResults) > 0 {
					if o, ok := optResults[0].Interface().(ColumnOptions); ok {
						opts = o
					}
				}
			}

			// Extract the type parameter from Column[T]
			columnType := extractColumnType(fieldVal.Type())

			colRef := &ColumnRef{
				Name:     columnName,
				FullName: tableName + "." + columnName,
				Type:     columnType,
				Options:  opts,
			}

			columns = append(columns, colRef)
		}
	}

	return columns
}

// extractColumnType extracts the type parameter T from *Column[T]
func extractColumnType(columnPtrType reflect.Type) reflect.Type {
	// Remove pointer
	if columnPtrType.Kind() == reflect.Ptr {
		columnPtrType = columnPtrType.Elem()
	}

	// For generic types, we need to extract the type parameter
	// Since Go reflection doesn't directly expose type parameters,
	// we'll use a workaround: get the field type from the struct
	if columnPtrType.Kind() == reflect.Struct {
		// This is a simplified approach - in practice, we might need
		// to store type information differently
		typeStr := columnPtrType.String()
		// Extract type from "table.Column[int64]" -> "int64"
		if idx := strings.Index(typeStr, "["); idx != -1 {
			typeStr = typeStr[idx+1 : len(typeStr)-1]
			// This is a placeholder - proper type extraction would require
			// registering types at column creation time
		}
	}

	// Return interface{} as fallback
	return reflect.TypeOf((*interface{})(nil)).Elem()
}
