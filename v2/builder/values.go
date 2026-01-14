package builder

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/guadalsistema/go-compose-sql/v2/table"
	"github.com/kisielk/sqlstruct"
)

// normalizeInsertValues converts input values (struct/map/slice) into row maps.
// The optional column list filters out fields not present on the table.
func normalizeInsertValues(data interface{}, cols []*table.ColumnRef) ([]map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("values cannot be nil")
	}

	// Build a fast lookup set for allowed columns.
	colSet := make(map[string]struct{}, len(cols))
	for _, col := range cols {
		colSet[col.Name] = struct{}{}
	}

	val := reflect.ValueOf(data)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("values cannot be nil")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		if val.Len() == 0 {
			return nil, fmt.Errorf("values cannot be empty")
		}
		// Collect one map per element.
		rows := make([]map[string]interface{}, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			row, err := extractRow(val.Index(i), colSet)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		return rows, nil
	default:
		row, err := extractRow(val, colSet)
		if err != nil {
			return nil, err
		}
		return []map[string]interface{}{row}, nil
	}
}

// extractRow normalizes a single value into a row map using struct tags or map keys.
func extractRow(val reflect.Value, colSet map[string]struct{}) (map[string]interface{}, error) {
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("values cannot be nil")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		return mapFromMap(val, colSet)
	case reflect.Struct:
		// Build a column map from exported struct fields and tags.
		row := make(map[string]interface{})
		if err := mapFromStruct(val, colSet, row); err != nil {
			return nil, err
		}
		if len(row) == 0 {
			return nil, fmt.Errorf("no insertable columns found")
		}
		return row, nil
	default:
		return nil, fmt.Errorf("unsupported values type: %s", val.Kind())
	}
}

// mapFromMap copies string-keyed values into a row map while applying column filters.
func mapFromMap(val reflect.Value, colSet map[string]struct{}) (map[string]interface{}, error) {
	if val.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("map keys must be strings")
	}

	row := make(map[string]interface{})
	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		// Skip keys not present in the table schema.
		if len(colSet) > 0 {
			if _, ok := colSet[key]; !ok {
				continue
			}
		}
		row[key] = iter.Value().Interface()
	}

	if len(row) == 0 {
		return nil, fmt.Errorf("no insertable columns found")
	}

	return row, nil
}

// mapFromStruct walks exported fields (including embedded structs) and fills row.
func mapFromStruct(val reflect.Value, colSet map[string]struct{}, row map[string]interface{}) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}

		// Inline embedded structs to match sqlstruct behavior.
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := mapFromStruct(val.Field(i), colSet, row); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get(sqlstruct.TagName)
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = sqlstruct.ToSnakeCase(field.Name)
		}

		// Respect the table column filter if present.
		if len(colSet) > 0 {
			if _, ok := colSet[tag]; !ok {
				continue
			}
		}

		row[tag] = val.Field(i).Interface()
	}
	return nil
}

// orderedInsertColumns chooses a stable column order for INSERT statements.
// It prefers table column order when available, otherwise alphabetical order.
func orderedInsertColumns(values map[string]interface{}, cols []*table.ColumnRef) []string {
	if len(values) == 0 {
		return nil
	}

	if len(cols) == 0 {
		// No schema ordering, so sort keys for deterministic SQL output.
		columns := make([]string, 0, len(values))
		for col := range values {
			columns = append(columns, col)
		}
		sort.Strings(columns)
		return columns
	}

	columns := make([]string, 0, len(values))
	for _, col := range cols {
		if _, ok := values[col.Name]; ok {
			columns = append(columns, col.Name)
		}
	}
	if len(columns) == 0 {
		// Fallback for mismatched schema and data.
		for col := range values {
			columns = append(columns, col)
		}
		sort.Strings(columns)
	}
	return columns
}
