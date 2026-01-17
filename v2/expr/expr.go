package expr

// Expr represents a SQL expression (WHERE, HAVING, etc.)
type Expr interface {
	// ToSQL converts the expression to SQL with placeholders
	ToSQL() (string, []interface{})
}

// SQLValue represents a value that can be used in SQL comparisons
// It can be either a column reference or a literal value
type SQLValue interface {
	// SQLString returns the SQL representation
	// Returns (sqlString, isLiteral) where:
	// - For columns: ("table.column", false)
	// - For values: ("?", true) with the actual value stored separately
	SQLString() (string, bool)
	// Value returns the actual value if this is a literal, nil otherwise
	Value() interface{}
}

// BinaryExpr represents a binary operation (=, !=, <, >, etc.)
type BinaryExpr struct {
	Left     string
	Operator string
	Right    interface{}
}

func (b *BinaryExpr) ToSQL() (string, []interface{}) {
	return b.Left + " " + b.Operator + " ?", []interface{}{b.Right}
}

// CompareExpr represents a comparison operation that supports both column and value comparisons
type CompareExpr struct {
	Left     string
	Operator string
	Right    SQLValue
}

func (c *CompareExpr) ToSQL() (string, []interface{}) {
	rightSQL, isLiteral := c.Right.SQLString()
	if isLiteral {
		// Value comparison: column = ?
		return c.Left + " " + c.Operator + " " + rightSQL, []interface{}{c.Right.Value()}
	}
	// Column comparison: column1 = column2
	return c.Left + " " + c.Operator + " " + rightSQL, nil
}

// Literal wraps a value to implement SQLValue interface
type Literal struct {
	Val interface{}
}

func (l Literal) SQLString() (string, bool) {
	return "?", true
}

func (l Literal) Value() interface{} {
	return l.Val
}

// V creates a Literal SQLValue from any value
func V(value interface{}) SQLValue {
	return Literal{Val: value}
}

// LogicalExpr represents AND/OR combinations
type LogicalExpr struct {
	Operator string // "AND" or "OR"
	Exprs    []Expr
}

func (l *LogicalExpr) ToSQL() (string, []interface{}) {
	if len(l.Exprs) == 0 {
		return "", nil
	}

	var sqlParts []string
	var args []interface{}

	for _, expr := range l.Exprs {
		sql, exprArgs := expr.ToSQL()
		if sql != "" {
			sqlParts = append(sqlParts, "("+sql+")")
			args = append(args, exprArgs...)
		}
	}

	if len(sqlParts) == 0 {
		return "", nil
	}

	if len(sqlParts) == 1 {
		return sqlParts[0], args
	}

	sql := "(" + sqlParts[0]
	for i := 1; i < len(sqlParts); i++ {
		sql += " " + l.Operator + " " + sqlParts[i]
	}
	sql += ")"

	return sql, args
}

// UnaryExpr represents unary operations (IS NULL, IS NOT NULL, NOT)
type UnaryExpr struct {
	Column   string
	Operator string
}

func (u *UnaryExpr) ToSQL() (string, []interface{}) {
	return u.Column + " " + u.Operator, nil
}

// InExpr represents IN/NOT IN operations
type InExpr struct {
	Column string
	Values []interface{}
	Not    bool
}

func (i *InExpr) ToSQL() (string, []interface{}) {
	if len(i.Values) == 0 {
		return "", nil
	}

	op := "IN"
	if i.Not {
		op = "NOT IN"
	}

	placeholders := ""
	for idx := range i.Values {
		if idx > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	sql := i.Column + " " + op + " (" + placeholders + ")"
	return sql, i.Values
}

// LikeExpr represents LIKE/ILIKE operations
type LikeExpr struct {
	Column          string
	Pattern         string
	CaseInsensitive bool
	Not             bool
}

func (l *LikeExpr) ToSQL() (string, []interface{}) {
	op := "LIKE"
	if l.CaseInsensitive {
		op = "ILIKE"
	}
	if l.Not {
		op = "NOT " + op
	}

	sql := l.Column + " " + op + " ?"
	return sql, []interface{}{l.Pattern}
}

// BetweenExpr represents BETWEEN operations
type BetweenExpr struct {
	Column string
	Start  interface{}
	End    interface{}
	Not    bool
}

func (b *BetweenExpr) ToSQL() (string, []interface{}) {
	op := "BETWEEN"
	if b.Not {
		op = "NOT BETWEEN"
	}

	sql := b.Column + " " + op + " ? AND ?"
	return sql, []interface{}{b.Start, b.End}
}

// RawExpr represents a raw SQL expression
type RawExpr struct {
	SQL  string
	Args []interface{}
}

func (r *RawExpr) ToSQL() (string, []interface{}) {
	return r.SQL, r.Args
}

// Helper functions for building expressions

// And combines multiple expressions with AND
func And(exprs ...Expr) Expr {
	return &LogicalExpr{
		Operator: "AND",
		Exprs:    exprs,
	}
}

// Or combines multiple expressions with OR
func Or(exprs ...Expr) Expr {
	return &LogicalExpr{
		Operator: "OR",
		Exprs:    exprs,
	}
}

// Raw creates a raw SQL expression
func Raw(sql string, args ...interface{}) Expr {
	return &RawExpr{
		SQL:  sql,
		Args: args,
	}
}
