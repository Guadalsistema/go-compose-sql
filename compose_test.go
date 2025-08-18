package sqlcompose

import (
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	type User struct {
		ID        int `db:"id"`
		FirstName string
		LastName  string `db:"last_name"`
	}

	clause := Insert[User](SqlOpts{})
	expected := "INSERT INTO user (id, first_name, last_name) VALUES (?, ?, ?);"
	if got := clause.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
	if clause.ModelType != reflect.TypeOf(User{}) {
		t.Fatalf("unexpected model type: %v", clause.ModelType)
	}
}

func TestInsertWithTableOpt(t *testing.T) {
	type Widget struct {
		Name string
	}

	clause := Insert[Widget](SqlOpts{TableName: "widgets"})
	expected := "INSERT INTO widgets (name) VALUES (?);"
	if got := clause.Write(); got != expected {
		t.Fatalf("unexpected SQL with table opt: %s", got)
	}
}

func TestSelect(t *testing.T) {
	type User struct {
		ID        int `db:"id"`
		FirstName string
	}

	clause := Select[User](SqlOpts{})
	expected := "SELECT id, first_name FROM user;"
	if got := clause.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestDelete(t *testing.T) {
	type User struct{}

	clause := Delete[User](SqlOpts{})
	expected := "DELETE FROM user;"
	if got := clause.Write(); got != expected {
		t.Fatalf("unexpected SQL: %s", got)
	}
}

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"User":       "user",
		"UserRole":   "user_role",
		"HTTPServer": "h_t_t_p_server",
	}
	for in, want := range cases {
		if got := toSnakeCase(in); got != want {
			t.Fatalf("toSnakeCase(%q)=%q; want %q", in, got, want)
		}
	}
}
