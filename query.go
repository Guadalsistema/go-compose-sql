package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

func Query[T any](db *sql.DB, stmt SQLStatement) (*QueryRowIterator[T], error) {
	return QueryContext[T](context.Background(), db, stmt)
}

// QueryRowIterator allows for iterating over the results of a query one by one.
type QueryRowIterator[T any] struct {
	rows  *sql.Rows
	isPtr bool
	model reflect.Type
}

// Next prepares the next result row for reading.
func (iter *QueryRowIterator[T]) Next() bool {
	return iter.rows.Next()
}

// Check if error happen
func (iter *QueryRowIterator[T]) Err() error {
	return iter.rows.Err()
}

// Scan scans the current row into the given destination.
func (iter *QueryRowIterator[T]) Scan(dest *T) error {
	pv := reflect.New(iter.model)
	if err := sqlstruct.Scan(pv.Interface(), iter.rows); err != nil {
		return err
	}
	if iter.isPtr {
		*dest = pv.Interface().(T)
	} else {
		*dest = pv.Elem().Interface().(T)
	}
	return nil
}

// Close closes the iterator, releasing any underlying resources.
func (iter *QueryRowIterator[T]) Close() error {
	return iter.rows.Close()
}

// QueryContext executes the SELECT SQLStatement against the provided database
// and returns a QueryRowIterator so the caller can iterate over the results.
func QueryContext[T any](ctx context.Context, db *sql.DB, stmt SQLStatement) (*QueryRowIterator[T], error) {
	if len(stmt.Clauses) == 0 || stmt.Clauses[0].Type != ClauseSelect {
		return nil, fmt.Errorf("sqlcompose: Query requires a SELECT clause")
	}

	first := stmt.Clauses[0]

	sqlStmt, err := stmt.Write()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, sqlStmt, stmt.Args()...)
	if err != nil {
		return nil, err
	}

	isPtr := reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Pointer

	return &QueryRowIterator[T]{
		rows:  rows,
		isPtr: isPtr,
		model: first.ModelType,
	}, nil
}

// QueryOne executes the SELECT SQLStatement against the provided database using
// context.Background(). It delegates to QueryOneContext.
func QueryOne[T any](db *sql.DB, stmt SQLStatement) (T, error) {
	return QueryOneContext[T](context.Background(), db, stmt)
}

// QueryOneContext executes the SELECT SQLStatement against the provided database
// using the supplied context and returns exactly one row. If the query returns
// zero or more than one row, it returns an error.
func QueryOneContext[T any](ctx context.Context, db *sql.DB, stmt SQLStatement) (T, error) {
	var zero T

	iter, err := QueryContext[T](ctx, db, stmt)
	if err != nil {
		return zero, err
	}
	defer iter.Close()

	if !iter.Next() {
		if err := iter.Err(); err != nil {
			return zero, err
		}
		return zero, sql.ErrNoRows
	}

	var result T
	if err := iter.Scan(&result); err != nil {
		return zero, err
	}

	if iter.Next() {
		return zero, fmt.Errorf("sqlcompose: QueryOne requires exactly one row")
	}
	if err := iter.Err(); err != nil {
		return zero, err
	}

	return result, nil
}
