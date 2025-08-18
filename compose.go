package sqlcompose

import (
	"fmt"
	"reflect"
	"strings"
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
type SqlClause struct {
	Type        ClauseType
	TableName   string
	ColumnNames []string
}

// Insert builds an INSERT clause based on the supplied struct and options.
//
// Fields are mapped to column names using the `db` struct tag; if absent, the
// field name is converted to snake_case. The table name defaults to the struct
// type name converted to snake_case when opts.TableName is empty.
func Insert[T any](columns T, opts SqlOpts) SqlClause {
	typ := reflect.TypeOf(columns)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := opts.TableName
	if tableName == "" {
		tableName = toSnakeCase(typ.Name())
	}

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
			tag = toSnakeCase(f.Name)
		}
		names = append(names, tag)
	}

	return SqlClause{
		Type:        ClauseInsert,
		TableName:   tableName,
		ColumnNames: names,
	}
}

// Select builds a SELECT clause listing all exported fields from the struct.
//
// Column names and table name follow the same rules as Insert.
func Select[T any](model T, opts SqlOpts) SqlClause {
	typ := reflect.TypeOf(model)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := opts.TableName
	if tableName == "" {
		tableName = toSnakeCase(typ.Name())
	}

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
			tag = toSnakeCase(f.Name)
		}
		names = append(names, tag)
	}

	return SqlClause{
		Type:        ClauseSelect,
		TableName:   tableName,
		ColumnNames: names,
	}
}

// Delete builds a DELETE clause for the given struct type.
//
// The table name defaults to the struct type name converted to snake_case when
// opts.TableName is empty.
func Delete[T any](model T, opts SqlOpts) SqlClause {
	typ := reflect.TypeOf(model)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	tableName := opts.TableName
	if tableName == "" {
		tableName = toSnakeCase(typ.Name())
	}

	return SqlClause{
		Type:      ClauseDelete,
		TableName: tableName,
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
		return fmt.Sprintf("SELECT %s FROM %s;", cols, c.TableName)
	case ClauseDelete:
		return fmt.Sprintf("DELETE FROM %s;", c.TableName)
	default:
		return ""
	}
}

// toSnakeCase converts CamelCase strings to snake_case.
func toSnakeCase(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, rune(strings.ToLower(string(r))[0]))
	}
	return string(out)
}
