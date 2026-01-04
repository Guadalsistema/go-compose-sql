package sqlcompose

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/kisielk/sqlstruct"
)

// Exec executes the INSERT, UPDATE, or DELETE statement against the provided database using
// context.Background(). It delegates to ExecContext.
func Exec(db *sql.DB, stmt SQLStatement, models ...any) (sql.Result, error) {
	return ExecContext(context.Background(), db, stmt, models...)
}

// ExecContext executes the INSERT, UPDATE, or DELETE SQLStatement against the provided database
// using the supplied context. The models' exported fields are mapped to column
// names in the first clause and passed as arguments to the INSERT statement.
//
// The first clause must be built using Insert[T] so that ModelType and
// ColumnNames match the fields in the model. ExecContext returns an error if
// the first clause is not an INSERT, UPDATE, or DELETE clause. It executes the statement once for
// each provided model and returns the result of the final execution. For DELETE
// statements, models are optional; if none are provided the statement is executed
// once using only the arguments supplied in the builder (e.g., WHERE clause).
//
// If the statement contains a RETURNING clause, ExecContext returns an error
// because Exec cannot retrieve returned values. Use Query instead.
func ExecContext(ctx context.Context, db *sql.DB, stmt SQLStatement, models ...any) (sql.Result, error) {
	if len(stmt.Clauses) == 0 {
		return nil, fmt.Errorf("sqlcompose: Exec requires an INSERT, UPDATE, or DELETE clause")
	}

	if stmt.Clauses[0].Type != ClauseInsert && stmt.Clauses[0].Type != ClauseUpdate && stmt.Clauses[0].Type != ClauseDelete {
		return nil, fmt.Errorf("sqlcompose: Exec requires an INSERT, UPDATE, or DELETE clause")
	}

	if hasReturningClause(stmt) {
		return nil, fmt.Errorf("sqlcompose: Exec cannot be used with RETURNING clause, use Query instead")
	}

	if len(models) == 0 && stmt.Clauses[0].Type != ClauseDelete {
		return nil, fmt.Errorf("sqlcompose: Exec requires at least one model")
	}

	first := stmt.Clauses[0]
	columns := make(map[string]struct{}, len(first.ColumnNames))
	for _, c := range first.ColumnNames {
		columns[c] = struct{}{}
	}

	sqlStmt, err := stmt.Write()
	if err != nil {
		return nil, err
	}

	var res sql.Result

	if len(models) == 0 && stmt.Clauses[0].Type == ClauseDelete {
		return db.ExecContext(ctx, sqlStmt, stmt.Args()...)
	}

	for _, model := range models {
		val := reflect.ValueOf(model)
		for val.Kind() == reflect.Pointer {
			val = val.Elem()
		}

		if !val.IsValid() || val.Type() != first.ModelType {
			return nil, fmt.Errorf("sqlcompose: model type %T does not match clause type %s", model, first.ModelType)
		}

		args := make([]any, 0, len(columns)+len(stmt.Args()))
		for i := 0; i < first.ModelType.NumField(); i++ {
			f := first.ModelType.Field(i)
			if f.PkgPath != "" || f.Tag.Get(sqlstruct.TagName) == "-" {
				continue
			}
			tag := f.Tag.Get(sqlstruct.TagName)
			if tag == "" {
				tag = sqlstruct.ToSnakeCase(f.Name)
			}
			if _, ok := columns[tag]; !ok {
				continue
			}
			args = append(args, val.Field(i).Interface())
		}
		args = append(args, stmt.Args()...)

		r, err := db.ExecContext(ctx, sqlStmt, args...)
		if err != nil {
			return r, err
		}
		res = r
	}

	return res, nil
}
