package expr

import "github.com/guadalsistema/go-compose-sql/v2/table"

// ColumnExpr provides expression methods for columns
// This is added to Column[T] via methods

// Eq creates an equality expression (column = value OR column = column)
// Accepts either a raw value or another column (SQLValue)
func Eq[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue

	// Check if value already implements SQLValue (e.g., another column)
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		// Wrap raw value in Literal
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: "=",
		Right:    sqlValue,
	}
}

// Ne creates a not-equal expression (column != value OR column != column)
func Ne[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: "!=",
		Right:    sqlValue,
	}
}

// Lt creates a less-than expression (column < value OR column < column)
func Lt[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: "<",
		Right:    sqlValue,
	}
}

// Le creates a less-than-or-equal expression (column <= value OR column <= column)
func Le[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: "<=",
		Right:    sqlValue,
	}
}

// Gt creates a greater-than expression (column > value OR column > column)
func Gt[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: ">",
		Right:    sqlValue,
	}
}

// Ge creates a greater-than-or-equal expression (column >= value OR column >= column)
func Ge[T any](col *table.Column[T], value any) Expr {
	var sqlValue SQLValue
	if sv, ok := value.(SQLValue); ok {
		sqlValue = sv
	} else {
		sqlValue = V(value)
	}

	return &CompareExpr{
		Left:     col.FullName(),
		Operator: ">=",
		Right:    sqlValue,
	}
}

// IsNull creates an IS NULL expression
func IsNull[T any](col *table.Column[T]) Expr {
	return &UnaryExpr{
		Column:   col.FullName(),
		Operator: "IS NULL",
	}
}

// IsNotNull creates an IS NOT NULL expression
func IsNotNull[T any](col *table.Column[T]) Expr {
	return &UnaryExpr{
		Column:   col.FullName(),
		Operator: "IS NOT NULL",
	}
}

// In creates an IN expression (column IN (values...))
func In[T any](col *table.Column[T], values ...T) Expr {
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}
	return &InExpr{
		Column: col.FullName(),
		Values: vals,
		Not:    false,
	}
}

// NotIn creates a NOT IN expression
func NotIn[T any](col *table.Column[T], values ...T) Expr {
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}
	return &InExpr{
		Column: col.FullName(),
		Values: vals,
		Not:    true,
	}
}

// Like creates a LIKE expression
func Like(col *table.Column[string], pattern string) Expr {
	return &LikeExpr{
		Column:  col.FullName(),
		Pattern: pattern,
		Not:     false,
	}
}

// NotLike creates a NOT LIKE expression
func NotLike(col *table.Column[string], pattern string) Expr {
	return &LikeExpr{
		Column:  col.FullName(),
		Pattern: pattern,
		Not:     true,
	}
}

// ILike creates an ILIKE expression (case-insensitive)
func ILike(col *table.Column[string], pattern string) Expr {
	return &LikeExpr{
		Column:          col.FullName(),
		Pattern:         pattern,
		CaseInsensitive: true,
		Not:             false,
	}
}

// Between creates a BETWEEN expression
func Between[T any](col *table.Column[T], start, end T) Expr {
	return &BetweenExpr{
		Column: col.FullName(),
		Start:  start,
		End:    end,
		Not:    false,
	}
}

// NotBetween creates a NOT BETWEEN expression
func NotBetween[T any](col *table.Column[T], start, end T) Expr {
	return &BetweenExpr{
		Column: col.FullName(),
		Start:  start,
		End:    end,
		Not:    true,
	}
}
