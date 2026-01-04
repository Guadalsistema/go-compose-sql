package expr

import "github.com/guadalsistema/go-compose-sql/v2/table"

// ColumnExpr provides expression methods for columns
// This is added to Column[T] via methods

// Eq creates an equality expression (column = value)
func Eq[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: "=",
		Right:    value,
	}
}

// Ne creates a not-equal expression (column != value)
func Ne[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: "!=",
		Right:    value,
	}
}

// Lt creates a less-than expression (column < value)
func Lt[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: "<",
		Right:    value,
	}
}

// Le creates a less-than-or-equal expression (column <= value)
func Le[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: "<=",
		Right:    value,
	}
}

// Gt creates a greater-than expression (column > value)
func Gt[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: ">",
		Right:    value,
	}
}

// Ge creates a greater-than-or-equal expression (column >= value)
func Ge[T any](col *table.Column[T], value T) Expr {
	return &BinaryExpr{
		Left:     col.FullName(),
		Operator: ">=",
		Right:    value,
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
