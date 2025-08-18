package sqlcompose

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kisielk/sqlstruct"
)

// ClauseType represents a SQL operation like INSERT or UPDATE.
type ClauseType string

const (
	ClauseInsert ClauseType = "INSERT"
	ClauseSelect ClauseType = "SELECT"
	ClauseDelete ClauseType = "DELETE"
)

// SqlOpts contains optional settings for building SQL clauses.
type SqlOpts struct {
	TableName string
}

// SqlClause represents a SQL statement before rendering.
//
// ModelType retains the generic type used when building the clause so that
// values can later be mapped to columns when executing the statement.
type SqlClause struct {
	Type        ClauseType
	TableName   string
	ColumnNames []string
	ModelType   reflect.Type
	WhereExpr   string
	WhereArgs   []any
}

func getTableName(def string, opts *SqlOpts) string {
	tableName := def
	if opts != nil && opts.TableName != "" {
		tableName = opts.TableName

	}
	return tableName
}

// Insert builds an INSERT clause for type T using the provided options.
//
// Fields are mapped to column names using the `db` struct tag; if absent, the
// field name is converted to snake_case. The table name defaults to the struct
// type name converted to snake_case when opts.TableName is empty. The reflected
// type is stored in the resulting SqlClause.
func Insert[T any](opts *SqlOpts) SqlClause {
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
		tag := f.Tag.Get("db")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = sqlstruct.ToSnakeCase(f.Name)
		}
		names = append(names, tag)
	}

	return SqlClause{
		Type:        ClauseInsert,
		TableName:   tableName,
		ColumnNames: names,
		ModelType:   typ,
	}
}

// Select builds a SELECT clause listing all exported fields of type T.
//
// Column names and table name follow the same rules as Insert. The reflected
// type is stored in the resulting SqlClause.
func Select[T any](opts *SqlOpts) SqlClause {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := getTableName(sqlstruct.ToSnakeCase(typ.Name()), opts)

	var names []string
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("db")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = sqlstruct.ToSnakeCase(f.Name)
		}
		names = append(names, tag)
	}

	return SqlClause{
		Type:        ClauseSelect,
		TableName:   tableName,
		ColumnNames: names,
		ModelType:   typ,
	}
}

// Delete builds a DELETE clause for type T.
//
// The table name defaults to the struct type name converted to snake_case when
// opts.TableName is empty. The reflected type is stored in the resulting
// SqlClause.
func Delete[T any](opts *SqlOpts) SqlClause {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := getTableName(sqlstruct.ToSnakeCase(typ.Name()), opts)

	return SqlClause{
		Type:      ClauseDelete,
		TableName: tableName,
		ModelType: typ,
	}
}

// Write renders the SQL clause to a string.
func (c SqlClause) Write() string {
	switch c.Type {
	case ClauseInsert:
		cols := strings.Join(c.ColumnNames, ", ")
		placeholders := strings.TrimRight(strings.Repeat("?, ", len(c.ColumnNames)), ", ")
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", c.TableName, cols, placeholders)
	case ClauseSelect:
		cols := strings.Join(c.ColumnNames, ", ")
		if c.WhereExpr != "" {
			return fmt.Sprintf("SELECT %s FROM %s WHERE %s;", cols, c.TableName, c.WhereExpr)
		}
		return fmt.Sprintf("SELECT %s FROM %s;", cols, c.TableName)
	case ClauseDelete:
		if c.WhereExpr != "" {
			return fmt.Sprintf("DELETE FROM %s WHERE %s;", c.TableName, c.WhereExpr)
		}
		return fmt.Sprintf("DELETE FROM %s;", c.TableName)
	default:
		return ""
	}
}

// Where returns a copy of the clause with a WHERE expression and args applied.
func (c SqlClause) Where(expr string, args ...any) SqlClause {
	c.WhereExpr = expr
	c.WhereArgs = args
	return c
}
