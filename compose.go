package sqlcompose

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kisielk/sqlstruct"
)

// SqlOpts contains optional settings for building SQL clauses.
type SqlOpts struct {
	TableName string
	Fields    []string
	// Driver chooses the SQL dialect for rendering; defaults to DefaultDriver when nil.
	Driver Driver
}

// SQLStatement represents a sequence of SQL clauses forming a statement.
type SQLStatement struct {
	Clauses []SqlClause
	Driver  Driver
}

// Write renders the complete SQL statement by concatenating all clauses using the configured Driver (or DefaultDriver).
func (s SQLStatement) Write() (string, error) {
	driver := s.Driver
	if driver == nil {
		driver = DefaultDriver
	}

	renderer := rendererForDriver(driver)
	sql, _, err := renderClauses(s, driver, renderer, 1)
	if err != nil {
		return "", err
	}
	if sql == "" {
		return "", nil
	}
	if needsSemicolon(driver) {
		sql += ";"
	}
	return sql, nil
}

func needsSemicolon(driver Driver) bool {
	switch driver.(type) {
	case PostgresDriver, *PostgresDriver:
		return false
	default:
		return true
	}
}

// Args returns the collected arguments from all clauses in the statement.
func (s SQLStatement) Args() []any {
	var out []any
	for _, c := range s.Clauses {
		if c.Type == ClauseJoin {
			out = append(out, c.JoinStatement.Args()...)
		}
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

// Coalesce appends a COALESCE expression to the SELECT list.
func (s SQLStatement) Coalesce(values ...any) SQLStatement {
	formatted := make([]string, 0, len(values))
	for _, v := range values {
		formatted = append(formatted, formatCoalesceValue(v))
	}
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseCoalesce, ColumnNames: formatted})
	return s
}

// Desc appends a DESC clause ensuring it follows an ORDER BY clause.
func (s SQLStatement) Desc() SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseDesc})
	return s
}

// Returning appends a RETURNING clause to INSERT, UPDATE, or DELETE statements.
func (s SQLStatement) Returning(columns ...string) SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseReturning, ColumnNames: columns})
	return s
}

// Asc appends an ASC clause ensuring it follows an ORDER BY clause.
func (s SQLStatement) Asc() SQLStatement {
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseAsc})
	return s
}

// Join appends a JOIN clause with a nested statement and identifier.
func (s SQLStatement) Join(stmt SQLStatement, identifier string, on string, args ...any) SQLStatement {
	if stmt.Driver == nil {
		stmt.Driver = s.Driver
	}
	s.Clauses = append(s.Clauses, SqlClause{
		Type:          ClauseJoin,
		Identifier:    identifier,
		JoinStatement: stmt,
		Expr:          on,
		Args:          args,
	})
	return s
}

// Update builds an UPDATE statement for type T using the provided options.
//
// Column names and table name follow the same rules as Insert. The reflected
// type is stored in the resulting clause.
func Update[T any](opts *SqlOpts) SQLStatement {
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
		Type:        ClauseUpdate,
		TableName:   tableName,
		ColumnNames: names,
		ModelType:   typ,
	}
	driver := DefaultDriver
	if opts != nil && opts.Driver != nil {
		driver = opts.Driver
	}
	return SQLStatement{Clauses: []SqlClause{clause}, Driver: driver}
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
	driver := DefaultDriver
	if opts != nil && opts.Driver != nil {
		driver = opts.Driver
	}
	return SQLStatement{Clauses: []SqlClause{clause}, Driver: driver}
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
	driver := DefaultDriver
	if opts != nil && opts.Driver != nil {
		driver = opts.Driver
	}
	return SQLStatement{Clauses: []SqlClause{clause}, Driver: driver}
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
	driver := DefaultDriver
	if opts != nil && opts.Driver != nil {
		driver = opts.Driver
	}
	return SQLStatement{Clauses: []SqlClause{clause}, Driver: driver}
}

func renderClauses(stmt SQLStatement, driver Driver, renderer placeholderRenderer, argPosition int) (string, int, error) {
	var parts []string
	var usedTotal int
	for i, c := range stmt.Clauses {
		if (c.Type == ClauseDesc || c.Type == ClauseAsc) && (i == 0 || stmt.Clauses[i-1].Type != ClauseOrderBy) {
			return "", 0, NewErrMisplacedClause(string(c.Type))
		}
		if c.Type == ClauseReturning {
			switch stmt.Clauses[0].Type {
			case ClauseInsert, ClauseUpdate, ClauseDelete:
			default:
				return "", 0, NewErrMisplacedClause(string(c.Type))
			}
		}
		if c.Type == ClauseJoin {
			joinSQL, used, err := renderJoinClause(stmt, c, driver, renderer, argPosition, i)
			if err != nil {
				return "", 0, err
			}
			argPosition += used
			usedTotal += used
			parts = append(parts, joinSQL)
			continue
		}
		p, used, err := driver.Write(c, argPosition)
		if err != nil {
			return "", 0, err
		}
		argPosition += used
		usedTotal += used
		if c.Type == ClauseCoalesce {
			if i == 0 || stmt.Clauses[i-1].Type != ClauseSelect {
				return "", 0, NewErrMisplacedClause(string(c.Type))
			}
			sel := parts[len(parts)-1]
			idx := strings.Index(sel, " FROM ")
			if idx == -1 {
				return "", 0, fmt.Errorf("sqlcompose: malformed SELECT clause")
			}
			parts[len(parts)-1] = sel[:idx] + ", " + p + sel[idx:]
			continue
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, " "), usedTotal, nil
}

func renderJoinClause(stmt SQLStatement, clause SqlClause, driver Driver, renderer placeholderRenderer, argPosition, index int) (string, int, error) {
	if len(stmt.Clauses) == 0 || stmt.Clauses[0].Type != ClauseSelect {
		return "", 0, NewErrMisplacedClause(string(ClauseJoin))
	}
	if index == 0 {
		return "", 0, NewErrMisplacedClause(string(ClauseJoin))
	}
	switch stmt.Clauses[index-1].Type {
	case ClauseWhere, ClauseOrderBy, ClauseLimit, ClauseOffset, ClauseReturning:
		return "", 0, NewErrMisplacedClause(string(ClauseJoin))
	}
	if len(clause.JoinStatement.Clauses) == 0 {
		return "", 0, NewErrMisplacedClause(string(ClauseJoin))
	}

	nestedDriver := clause.JoinStatement.Driver
	if nestedDriver == nil {
		nestedDriver = driver
	}
	nestedRenderer := rendererForDriver(nestedDriver)

	switch clause.JoinStatement.Clauses[0].Type {
	case ClauseSelect:
	case ClauseInsert, ClauseUpdate, ClauseDelete:
		if !hasReturningClause(clause.JoinStatement) {
			return "", 0, NewErrMisplacedClause(string(ClauseJoin))
		}
	default:
		return "", 0, NewErrMisplacedClause(string(ClauseJoin))
	}

	innerSQL, usedInner, err := renderClauses(clause.JoinStatement, nestedDriver, nestedRenderer, argPosition)
	if err != nil {
		return "", 0, err
	}

	onExpr, usedOn := replacePlaceholders(clause.Expr, argPosition+usedInner, renderer)

	joinSQL := fmt.Sprintf("JOIN (%s) %s ON %s", innerSQL, clause.Identifier, onExpr)
	return joinSQL, usedInner + usedOn, nil
}
