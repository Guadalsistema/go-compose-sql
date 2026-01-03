package sqlcompose

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
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

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

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

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

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

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

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

func TestQueryInsertReturning(t *testing.T) {
	type User struct {
		ID   int    `sql:"id"`
		Name string `sql:"name"`
	}

	stmt := Insert[User](nil).Returning("id")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(42)

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlStr)).WillReturnRows(rows)

	iter, err := QueryContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	defer iter.Close()

	var user User
	if !iter.Next() {
		t.Fatalf("expected one row")
	}
	if err := iter.Scan(&user); err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if user.ID != 42 {
		t.Fatalf("unexpected user ID: %d", user.ID)
	}
	if iter.Next() {
		t.Fatalf("unexpected additional rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryDeleteReturning(t *testing.T) {
	type User struct {
		ID   int    `sql:"id"`
		Name string `sql:"name"`
	}

	stmt := Delete[User](nil).Where("id=?", 1).Returning("id", "name")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlStr)).WithArgs(1).WillReturnRows(rows)

	iter, err := QueryContext[User](context.Background(), db, stmt)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	defer iter.Close()

	var user User
	if !iter.Next() {
		t.Fatalf("expected one row")
	}
	if err := iter.Scan(&user); err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if user.ID != 1 || user.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if iter.Next() {
		t.Fatalf("unexpected additional rows")
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
	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

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
	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

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
	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.ExpectQuery(sqlStr).WillReturnRows(rows)

	if _, err := QueryOneContext[User](context.Background(), db, stmt); err == nil {
		t.Fatalf("expected error for multiple rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestQueryInvalidClause(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}
	stmt := Select[User](nil)
	stmt.Clauses = append(stmt.Clauses, SqlClause{Type: ClauseType("BAD")})

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := QueryContext[User](context.Background(), db, stmt); err == nil {
		t.Fatalf("expected error for invalid clause")
	} else {
		var clauseErr *ErrInvalidClause
		if !errors.As(err, &clauseErr) {
			t.Fatalf("expected ErrInvalidClause, got %v", err)
		}
	}
}

func TestQueryMisplacedClause(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}
	stmt := Select[User](nil).Desc()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := QueryContext[User](context.Background(), db, stmt); err == nil {
		t.Fatalf("expected error for misplaced clause")
	} else {
		var clauseErr *ErrMisplacedClause
		if !errors.As(err, &clauseErr) {
			t.Fatalf("expected ErrMisplacedClause, got %v", err)
		}
		if clauseErr.Clause != string(ClauseDesc) {
			t.Fatalf("unexpected clause: %s", clauseErr.Clause)
		}
	}
}
