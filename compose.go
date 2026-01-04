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

// Values appends a VALUES clause to INSERT or UPDATE statements with explicit values.
// This allows specifying values directly instead of passing models to Exec.
//
// If the first argument is a struct matching the statement's ModelType, its field
// values are automatically extracted in the same order as the clause's ColumnNames.
// Otherwise, all arguments are used as-is.
func (s SQLStatement) Values(values ...any) SQLStatement {
	if len(values) == 0 {
		s.Clauses = append(s.Clauses, SqlClause{Type: ClauseValues, Args: values})
		return s
	}

	// Check if we have a first clause with a ModelType
	if len(s.Clauses) == 0 {
		s.Clauses = append(s.Clauses, SqlClause{Type: ClauseValues, Args: values})
		return s
	}

	first := s.Clauses[0]
	if first.ModelType == nil {
		s.Clauses = append(s.Clauses, SqlClause{Type: ClauseValues, Args: values})
		return s
	}

	// Check if first value is a struct matching the ModelType
	val := reflect.ValueOf(values[0])
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.IsValid() && val.Type() == first.ModelType && len(values) == 1 {
		// Extract field values from the struct in the order of ColumnNames
		extractedValues := extractFieldValues(val, first.ModelType, first.ColumnNames)
		s.Clauses = append(s.Clauses, SqlClause{Type: ClauseValues, Args: extractedValues})
		return s
	}

	// Otherwise, use values as-is
	s.Clauses = append(s.Clauses, SqlClause{Type: ClauseValues, Args: values})
	return s
}

func extractFieldValues(val reflect.Value, typ reflect.Type, columnNames []string) []any {
	columns := make(map[string]struct{}, len(columnNames))
	for _, c := range columnNames {
		columns[c] = struct{}{}
	}

	args := make([]any, 0, len(columns))
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
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
	return args
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
	var fieldFilter map[string]struct{}
	if opts != nil && len(opts.Fields) > 0 {
		fieldFilter = make(map[string]struct{}, len(opts.Fields))
		for _, f := range opts.Fields {
			fieldFilter[f] = struct{}{}
		}
	}

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
		if fieldFilter != nil {
			if _, ok := fieldFilter[tag]; !ok {
				continue
			}
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
		if c.Type == ClauseValues {
			if i == 0 {
				return "", 0, NewErrMisplacedClause(string(c.Type))
			}
			prevType := stmt.Clauses[i-1].Type
			if prevType == ClauseInsert {
				insertClause := parts[len(parts)-1]
				idx := strings.Index(insertClause, " VALUES ")
				if idx == -1 {
					return "", 0, fmt.Errorf("sqlcompose: malformed INSERT clause")
				}
				// The INSERT clause already consumed placeholders for its columns,
				// but we're replacing those with the VALUES clause placeholders.
				// We need to start from the position before the INSERT consumed them.
				insertColumns := len(stmt.Clauses[i-1].ColumnNames)
				valuesStartPos := argPosition - insertColumns
				placeholdersList := make([]string, len(c.Args))
				for j := range placeholdersList {
					placeholdersList[j] = renderer.Placeholder(valuesStartPos + j)
				}
				parts[len(parts)-1] = insertClause[:idx] + fmt.Sprintf(" VALUES (%s)", strings.Join(placeholdersList, ", "))
				// Adjust the position and total: we're replacing insertColumns placeholders with len(c.Args) placeholders
				argPosition = argPosition - insertColumns + len(placeholdersList)
				usedTotal = usedTotal - insertColumns + len(placeholdersList)
				continue
			} else if prevType == ClauseUpdate {
				// For UPDATE, VALUES provides the values for the SET clause
				// The UPDATE clause already has placeholders, we just need to ensure
				// the args are in the right order. The VALUES clause is transparent here.
				continue
			} else {
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
