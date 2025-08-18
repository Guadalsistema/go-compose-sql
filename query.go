package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

// Query executes the SELECT SqlClause against the provided database and scans
// the resulting rows into a slice of T.
//
// The SqlClause must be built using Select[T] so that ModelType and ColumnNames
// match the fields in T. Query returns an error if the clause is not a SELECT
// clause.
func Query[T any](ctx context.Context, db *sql.DB, clause SqlClause) ([]T, error) {
	if clause.Type != ClauseSelect {
		return nil, fmt.Errorf("sqlcompose: Query requires a SELECT clause")
	}

	rows, err := db.QueryContext(ctx, clause.Write())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Determine if T is a pointer type so we can return the correct form.
	isPtr := reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Pointer

	var out []T
	for rows.Next() {
		// Create a new value of the model type and prepare destinations for Scan.
		pv := reflect.New(clause.ModelType)
		val := pv.Elem()
		var dest []any
		for i := 0; i < clause.ModelType.NumField(); i++ {
			f := clause.ModelType.Field(i)
			if f.PkgPath != "" {
				continue // skip unexported fields
			}
			if f.Tag.Get("db") == "-" {
				continue
			}
			dest = append(dest, val.Field(i).Addr().Interface())
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		if isPtr {
			out = append(out, pv.Interface().(T))
		} else {
			out = append(out, val.Interface().(T))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
