package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

// Exec executes the INSERT statement against the provided database using
// context.Background(). It delegates to ExecContext.
func Exec(db *sql.DB, stmt SQLStatement, model any) (sql.Result, error) {
	return ExecContext(context.Background(), db, stmt, model)
}

// ExecContext executes the INSERT SQLStatement against the provided database
// using the supplied context. The model's exported fields are mapped to
// column names in the first clause and passed as arguments to the INSERT
// statement.
//
// The first clause must be built using Insert[T] so that ModelType and
// ColumnNames match the fields in the model. ExecContext returns an error if
// the first clause is not an INSERT clause.
func ExecContext(ctx context.Context, db *sql.DB, stmt SQLStatement, model any) (sql.Result, error) {
	if len(stmt.Clauses) == 0 || stmt.Clauses[0].Type != ClauseInsert {
		return nil, fmt.Errorf("sqlcompose: Exec requires an INSERT clause")
	}

	first := stmt.Clauses[0]

	val := reflect.ValueOf(model)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if !val.IsValid() || val.Type() != first.ModelType {
		return nil, fmt.Errorf("sqlcompose: model type %T does not match clause type %s", model, first.ModelType)
	}

	var args []any
	for i := 0; i < first.ModelType.NumField(); i++ {
		f := first.ModelType.Field(i)
		if f.PkgPath != "" || f.Tag.Get(sqlstruct.TagName) == "-" {
			continue
		}
		args = append(args, val.Field(i).Interface())
	}

	return db.ExecContext(ctx, stmt.Write(), args...)
}
