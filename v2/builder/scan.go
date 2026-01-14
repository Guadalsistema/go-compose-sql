package builder

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

// scanAll reads every row and appends it to the destination slice.
// dest must be a pointer to a slice of structs, pointers to structs, or basic types.
func scanAll(rows *sql.Rows, dest interface{}) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer to a slice")
	}

	sliceVal := rv.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	elemType := sliceVal.Type().Elem()

	for rows.Next() {
		// Allocate a new element and pick an addressable scan target.
		elemVal, scanTarget := newScanTarget(elemType)
		if err := scanRow(rows, scanTarget); err != nil {
			return err
		}

		// Preserve pointer element types; otherwise append the value.
		if elemType.Kind() == reflect.Ptr {
			sliceVal = reflect.Append(sliceVal, elemVal)
		} else {
			sliceVal = reflect.Append(sliceVal, elemVal.Elem())
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	rv.Elem().Set(sliceVal)
	return nil
}

// scanOne reads exactly one row into dest, erroring on zero or multiple rows.
// dest must be a non-nil pointer to a struct, pointer-to-struct, or basic type.
func scanOne(rows *sql.Rows, dest interface{}) error {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	if err := scanRow(rows, dest); err != nil {
		return err
	}

	if rows.Next() {
		return fmt.Errorf("expected exactly one row")
	}

	return rows.Err()
}

// scanRow routes scanning based on the destination type.
// Structs use sqlstruct to map columns; non-structs fall back to rows.Scan.
func scanRow(rows *sql.Rows, dest interface{}) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() == reflect.Struct {
		return sqlstruct.Scan(dest, rows)
	}

	if elem.Kind() == reflect.Ptr && elem.Type().Elem().Kind() == reflect.Struct {
		// Ensure the pointer is initialized before scanning.
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		return sqlstruct.Scan(elem.Interface(), rows)
	}

	return rows.Scan(dest)
}

// newScanTarget allocates a value compatible with elemType and returns both the
// value and the interface pointer to pass into scanRow.
func newScanTarget(elemType reflect.Type) (reflect.Value, interface{}) {
	if elemType.Kind() == reflect.Ptr {
		elemVal := reflect.New(elemType.Elem())
		return elemVal, elemVal.Interface()
	}

	elemVal := reflect.New(elemType)
	return elemVal, elemVal.Interface()
}
