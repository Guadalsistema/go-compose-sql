package sqlcompose

import (
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQuery(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	clause := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name"}).
		AddRow(1, "Alice", "Smith").
		AddRow(2, "Bob", "Jones")

	mock.ExpectQuery(clause.Write()).WillReturnRows(rows)

	got, err := QueryContext[User](context.Background(), db, clause)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	want := []User{{1, "Alice", "Smith"}, {2, "Bob", "Jones"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryWhereArgs(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	clause := Select[User](nil).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, "Alice")
	mock.ExpectQuery(regexp.QuoteMeta(clause.Write())).WithArgs(1).WillReturnRows(rows)

	got, err := QueryContext[User](context.Background(), db, clause)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	if len(got) != 1 || got[0].ID != 1 || got[0].FirstName != "Alice" {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryPointer(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
	}

	clause := Select[*User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery(clause.Write()).WillReturnRows(rows)

	got, err := QueryContext[*User](context.Background(), db, clause)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}

	if len(got) != 2 || got[0].FirstName != "Alice" || got[1].FirstName != "Bob" {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryNonSelect(t *testing.T) {
	type User struct{ ID int }
	clause := Insert[User](nil)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := QueryContext[User](context.Background(), db, clause); err == nil {
		t.Fatalf("expected error for non-select clause")
	}
}
