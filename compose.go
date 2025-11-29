package sqlcompose

import (
	"reflect"
	"strings"

	"github.com/kisielk/sqlstruct"
)

// SqlOpts contains optional settings for building SQL clauses.
type SqlOpts struct {
	TableName string
	Fields    []string
}

// SQLStatement represents a sequence of SQL clauses forming a statement.
type SQLStatement struct {
	Clauses []SqlClause
}

// Write renders the complete SQL statement by concatenating all clauses.
func (s SQLStatement) Write() (string, error) {
	var parts []string
	for i, c := range s.Clauses {
		if (c.Type == ClauseDesc || c.Type == ClauseAsc) && (i == 0 || s.Clauses[i-1].Type != ClauseOrderBy) {
			return "", NewErrMisplacedClause(string(c.Type))
		}
		p, err := c.Write()
		if err != nil {
			return "", err
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " ") + ";", nil
}

// Args returns the collected arguments from all clauses in the statement.
func (s SQLStatement) Args() []any {
	var out []any
	for _, c := range s.Clauses {
		out = append(out, c.Args...)
	}
	return out
}

// Where appends a WHERE clause to the statement.
func (s SQLStatement) Where(expr string, args ...any) SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseWhere, Expr: expr, Args: args})
	return s
}

// OrderBy appends an ORDER BY clause to the statement.
func (s SQLStatement) OrderBy(columns ...string) SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseOrderBy, ColumnNames: columns})
	return s
}

// Limit appends a LIMIT clause to the statement.
func (s SQLStatement) Limit(n int) SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseLimit, Args: []any{n}})
	return s
}

// Offset appends an OFFSET clause to the statement.
func (s SQLStatement) Offset(n int) SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseOffset, Args: []any{n}})
	return s
}

// Desc appends a DESC clause ensuring it follows an ORDER BY clause.
func (s SQLStatement) Desc() SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseDesc})
	return s
}

// Asc appends an ASC clause ensuring it follows an ORDER BY clause.
func (s SQLStatement) Asc() SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseAsc})
	return s
}

func getTableName(def string, opts *SqlOpts) string {
	tableName := def
	if opts != nil && opts.TableName != "" {
		tableName = opts.TableName

	}
	return tableName
}

// Insert builds an INSERT statement for type T using the provided options.
//
// Fields are mapped to column names using the `db` struct tag; if absent, the
// field name is converted to snake_case. The table name defaults to the struct
// type name converted to snake_case when opts.TableName is empty. The reflected
// type is stored in the resulting clause.
func Insert[T any](opts *SqlOpts) SQLStatement {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := getTableName(sqlstruct.ToSnakeCase(typ.Name()), opts)

	var names []string
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		// Skip unexported fields
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get(sqlstruct.TagName)
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = sqlstruct.ToSnakeCase(f.Name)
		}
		names = append(names, tag)
	}

	clause := SqlClause{
		Type:        ClauseInsert,
		TableName:   tableName,
		ColumnNames: names,
		ModelType:   typ,
	}
	return SQLStatement{Clauses: []SqlClause{clause}}
}

// Select builds a SELECT statement listing all exported fields of type T.
//
// Column names and table name follow the same rules as Insert. The reflected
// type is stored in the resulting clause.
func Select[T any](opts *SqlOpts) SQLStatement {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := getTableName(sqlstruct.ToSnakeCase(typ.Name()), opts)

	var names []string
	var fieldFilter map[string]struct{}
	if opts != nil && len(opts.Fields) > 0 {
		fieldFilter = make(map[string]struct{}, len(opts.Fields))
		for _, f := range opts.Fields {
			fieldFilter[f] = struct{}{}
		}
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get(sqlstruct.TagName)
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = sqlstruct.ToSnakeCase(f.Name)
		}
		if fieldFilter != nil {
			if _, ok := fieldFilter[tag]; !ok {
				continue
			}
		}
		names = append(names, tag)
	}

	clause := SqlClause{
		Type:        ClauseSelect,
		TableName:   tableName,
		ColumnNames: names,
		ModelType:   typ,
	}
	return SQLStatement{Clauses: []SqlClause{clause}}
}

// Delete builds a DELETE statement for type T.
//
// The table name defaults to the struct type name converted to snake_case when
// opts.TableName is empty. The reflected type is stored in the resulting clause.
func Delete[T any](opts *SqlOpts) SQLStatement {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := getTableName(sqlstruct.ToSnakeCase(typ.Name()), opts)

	clause := SqlClause{
		Type:      ClauseDelete,
		TableName: tableName,
		ModelType: typ,
	}
	return SQLStatement{Clauses: []SqlClause{clause}}
}
