package sqlcompose

import (
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
	if got := stmt.Write(); got != expected {
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
	if got := stmt.Write(); got != expected {
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
	if got := stmt.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestSelectWhere(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Where("id=?", 1)
	expected := "SELECT id, first_name FROM user WHERE id=?;"
	if got := stmt.Write(); got != expected {
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
	if got := stmt.Write(); got != expected {
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
	if got := stmt.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestDescRequiresOrderBy(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when DESC used without ORDER BY")
		}
	}()
	type User struct{ ID int }
	Select[User](nil).Desc()
}

func TestAscRequiresOrderBy(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when ASC used without ORDER BY")
		}
	}()
	type User struct{ ID int }
	Select[User](nil).Asc()
}

func TestSelectLimit(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	stmt := Select[User](nil).Limit(5)
	expected := "SELECT id, first_name FROM user LIMIT ?;"
	if got := stmt.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
	args := stmt.Args()
	if len(args) != 1 || args[0] != 5 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestDelete(t *testing.T) {
	type User struct{}

	stmt := Delete[User](nil)
	expected := "DELETE FROM user;"
	if got := stmt.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
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
