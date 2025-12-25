package sqlcompose

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kisielk/sqlstruct"
)

func TestInsert(t *testing.T) {
	type User struct {
		ID        int `db:"id"`
		FirstName string
		LastName  string `db:"last_name"`
	}

	stmt := Insert[User](nil)
	expected := "INSERT INTO user (id, first_name, last_name) VALUES (?, ?, ?);"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
	if stmt.Clauses[0].ModelType != reflect.TypeOf(User{}) {
		t.Fatalf("unexpected model type: %v", stmt.Clauses[0].ModelType)
	}
}

func TestInsertWithTableOpt(t *testing.T) {
	type Widget struct {
		Name string
	}

	stmt := Insert[Widget](&SqlOpts{TableName: "widgets"})
	expected := "INSERT INTO widgets (name) VALUES (?);"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL with table opt: %s", got)
	}
}

func TestSelect(t *testing.T) {
	type User struct {
		ID        int `db:"id"`
		FirstName string
	}

	stmt := Select[User](nil)
	expected := "SELECT id, first_name FROM user;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestSelectWithFieldsOpt(t *testing.T) {
	type User struct {
		ID        int `db:"id"`
		FirstName string
		LastName  string `db:"last_name"`
	}

	stmt := Select[User](&SqlOpts{Fields: []string{"id", "last_name"}})
	expected := "SELECT id, last_name FROM user;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL with fields opt: %s", got)
	}
}

func TestSelectWhere(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Where("id=?", 1)
	expected := "SELECT id, first_name FROM user WHERE id=?;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestSelectOrderByDesc(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).OrderBy("id").Desc()
	expected := "SELECT id, first_name FROM user ORDER BY id DESC;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestSelectOrderByAsc(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).OrderBy("id").Asc()
	expected := "SELECT id, first_name FROM user ORDER BY id ASC;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestDescRequiresOrderBy(t *testing.T) {
	type User struct{ ID int }
	stmt := Select[User](nil).Desc()
	_, err := stmt.Write()
	var clauseErr *ErrMisplacedClause
	if !errors.As(err, &clauseErr) {
		t.Fatalf("expected ErrMisplacedClause, got %v", err)
	}
	if clauseErr.Clause != string(ClauseDesc) {
		t.Fatalf("unexpected clause: %s", clauseErr.Clause)
	}
}

func TestAscRequiresOrderBy(t *testing.T) {
	type User struct{ ID int }
	stmt := Select[User](nil).Asc()
	_, err := stmt.Write()
	var clauseErr *ErrMisplacedClause
	if !errors.As(err, &clauseErr) {
		t.Fatalf("expected ErrMisplacedClause, got %v", err)
	}
	if clauseErr.Clause != string(ClauseAsc) {
		t.Fatalf("unexpected clause: %s", clauseErr.Clause)
	}
}

func TestSelectLimit(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Limit(5)
	expected := "SELECT id, first_name FROM user LIMIT ?;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
	args := stmt.Args()
	if len(args) != 1 || args[0] != 5 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelectOffset(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Limit(5).Offset(10)
	expected := "SELECT id, first_name FROM user LIMIT ? OFFSET ?;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
	args := stmt.Args()
	if len(args) != 2 || args[0] != 5 || args[1] != 10 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestSelectCoalesce(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Coalesce("first_name", "'unknown'")
	expected := "SELECT id, first_name, COALESCE(first_name, 'unknown') FROM user;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestCoalesceRequiresSelect(t *testing.T) {
	stmt := Insert[struct{}](nil).Coalesce("a", "b")
	_, err := stmt.Write()
	var clauseErr *ErrMisplacedClause
	if !errors.As(err, &clauseErr) {
		t.Fatalf("expected ErrMisplacedClause, got %v", err)
	}
	if clauseErr.Clause != string(ClauseCoalesce) {
		t.Fatalf("unexpected clause: %s", clauseErr.Clause)
	}
}

func TestCoalesceRequiresTwoValues(t *testing.T) {
	stmt := Select[struct{}](nil).Coalesce("only_one")
	_, err := stmt.Write()
	var coErr *ErrInvalidCoalesceArgs
	if !errors.As(err, &coErr) {
		t.Fatalf("expected ErrInvalidCoalesceArgs, got %v", err)
	}
}

func TestCoalesceFormatsAnyValues(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Coalesce("first_name", nil, 0)
	expected := "SELECT id, first_name, COALESCE(first_name, NULL, 0) FROM user;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestDelete(t *testing.T) {
	type User struct{}

	stmt := Delete[User](nil)
	expected := "DELETE FROM user;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestDeleteReturning(t *testing.T) {
	type User struct{}

	stmt := Delete[User](nil).Where("id=?", 1).Returning("id")
	expected := "DELETE FROM user WHERE id=? RETURNING id;"
	got, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestReturningRequiresDML(t *testing.T) {
	stmt := Select[struct{}](nil).Returning("id")
	if _, err := stmt.Write(); err == nil {
		t.Fatalf("expected error for misplaced RETURNING clause")
	} else {
		var clauseErr *ErrMisplacedClause
		if !errors.As(err, &clauseErr) {
			t.Fatalf("expected ErrMisplacedClause, got %v", err)
		}
		if clauseErr.Clause != string(ClauseReturning) {
			t.Fatalf("unexpected clause: %s", clauseErr.Clause)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"User":       "user",
		"UserRole":   "user_role",
		"HTTPServer": "httpserver",
	}
	for in, want := range cases {
		if got := sqlstruct.ToSnakeCase(in); got != want {
			t.Fatalf("ToSnakeCase(%q)=%q; want %q", in, got, want)
		}
	}
}

func TestInvalidClause(t *testing.T) {
	stmt := SQLStatement{Clauses: []SqlClause{{Type: ClauseType("BAD")}}}
	_, err := stmt.Write()
	var clauseErr *ErrInvalidClause
	if !errors.As(err, &clauseErr) {
		t.Fatalf("expected ErrInvalidClause, got %v", err)
	}
	if clauseErr.Clause != "BAD" {
		t.Fatalf("unexpected clause name: %s", clauseErr.Clause)
	}
}
