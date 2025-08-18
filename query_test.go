package sqlcompose

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQuery(t *testing.T) {
	type User struct {
		ID        int    `sql:"id"`
		FirstName string `sql:"first_name"`
		LastName  string `sql:"last_name"`
	}

	stmt := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name"}).
		AddRow(1, "Alice", "Smith").
		AddRow(2, "Bob", "Jones")

	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	got, err := QueryContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	defer got.Close()

	var user User
	got.Next()
	err = got.Scan(&user)
	if err != nil || user.FirstName != "Alice" {
		t.Fatalf("unexpected result: %+v", got)
	}

	want := User{1, "Alice", "Smith"}
	if !reflect.DeepEqual(user, want) {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryWhereArgs(t *testing.T) {
	type User struct {
		ID        int    `sql:"id"`
		FirstName string `sql:"first_name"`
	}

	stmt := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	got, err := QueryContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	defer got.Close()

	var user User
	got.Next()
	err = got.Scan(&user)
	if err != nil || user.FirstName != "Alice" {
		t.Fatalf("unexpected result: %+v", got)
	}
	got.Next()
	err = got.Scan(&user)
	if err != nil || user.FirstName != "Bob" {
		t.Fatalf("unexpected result: %+v", got)
	}
	follow := got.Next()
	if follow {
		t.Fatalf("unmet expectations, it should ahve no more data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryPointer(t *testing.T) {
	type User struct {
		ID        int    `sql:"id"`
		FirstName string `sql:"first_name"`
	}

	stmt := Select[*User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	got, err := QueryContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	defer got.Close()

	var user User
	got.Next()
	err = got.Scan(&user)
	if err != nil || user.FirstName != "Alice" {
		t.Fatalf("unexpected result: %+v", got)
	}
	got.Next()
	err = got.Scan(&user)
	if err != nil || user.FirstName != "Bob" {
		t.Fatalf("unexpected result: %+v", got)
	}
	follow := got.Next()
	if follow {
		t.Fatalf("unmet expectations, it should ahve no more data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryNonSelect(t *testing.T) {
	type User struct{ ID int }
	stmt := Insert[User](nil)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := QueryContext[User](context.Background(), db, stmt); err == nil {
		t.Fatalf("expected error for non-select clause")
	}
}

func TestQueryOne(t *testing.T) {
	type User struct {
		ID        int    `sql:"id"`
		FirstName string `sql:"first_name"`
	}

	stmt := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, "Alice")
	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	got, err := QueryOneContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("QueryOne returned error: %v", err)
	}

	want := User{1, "Alice"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryOneNoRows(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}

	stmt := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	if _, err := QueryOneContext[User](context.Background(), db, stmt); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryOneMultipleRows(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}

	stmt := Select[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mock.ExpectQuery(stmt.Write()).WillReturnRows(rows)

	if _, err := QueryOneContext[User](context.Background(), db, stmt); err == nil {
		t.Fatalf("expected error for multiple rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
