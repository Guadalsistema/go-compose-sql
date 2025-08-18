package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

func Query[T any](ctx context.Context, db *sql.DB, stmt SQLStatement) ([]T, error) {
	return QueryContext[T](context.Background(), db, stmt)
}

// Query executes the SELECT SQLStatement against the provided database and scans
// the resulting rows into a slice of T.
//
// The first clause must be built using Select[T] so that ModelType and
// ColumnNames match the fields in T. Query returns an error if the first clause
// is not a SELECT clause.
func QueryContext[T any](ctx context.Context, db *sql.DB, stmt SQLStatement) ([]T, error) {
	if len(stmt.Clauses) == 0 || stmt.Clauses[0].Type != ClauseSelect {
		return nil, fmt.Errorf("sqlcompose: Query requires a SELECT clause")
	}

	first := stmt.Clauses[0]

	rows, err := db.QueryContext(ctx, stmt.Write(), stmt.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Determine if T is a pointer type so we can return the correct form.
	isPtr := reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Pointer

	var out []T
	for rows.Next() {
		pv := reflect.New(first.ModelType)
		if err := sqlstruct.Scan(pv.Interface(), rows); err != nil {
			return nil, err
		}
		if isPtr {
			out = append(out, pv.Interface().(T))
		} else {
			out = append(out, pv.Elem().Interface().(T))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
