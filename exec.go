package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

// Exec executes the INSERT SqlClause against the provided database using
// context.Background(). It delegates to ExecContext.
func Exec(db *sql.DB, clause SqlClause, model any) (sql.Result, error) {
	return ExecContext(context.Background(), db, clause, model)
}

// ExecContext executes the INSERT SqlClause against the provided database
// using the supplied context. The model's exported fields are mapped to
// column names in the clause and passed as arguments to the INSERT statement.
//
// The SqlClause must be built using Insert[T] so that ModelType and
// ColumnNames match the fields in the model. ExecContext returns an error if
// the clause is not an INSERT clause.
func ExecContext(ctx context.Context, db *sql.DB, clause SqlClause, model any) (sql.Result, error) {
	if clause.Type != ClauseInsert {
		return nil, fmt.Errorf("sqlcompose: Exec requires an INSERT clause")
	}

	val := reflect.ValueOf(model)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if !val.IsValid() || val.Type() != clause.ModelType {
		return nil, fmt.Errorf("sqlcompose: model type %T does not match clause type %s", model, clause.ModelType)
	}

	var args []any
	for i := 0; i < clause.ModelType.NumField(); i++ {
		f := clause.ModelType.Field(i)
		if f.PkgPath != "" || f.Tag.Get(sqlstruct.TagName) == "-" {
			continue
		}
		args = append(args, val.Field(i).Interface())
	}

	return db.ExecContext(ctx, clause.Write(), args...)
}
