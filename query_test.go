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

func TestQueryOneInsertValuesReturningDifferentType(t *testing.T) {
	type User struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	user := User{Name: "Alice"}
	stmt := Insert[User](nil).Values(user).Returning("id")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect the query to return just the ID
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(42))
	mock.ExpectQuery(regexp.QuoteMeta(sqlStr)).
		WithArgs(int64(0), "Alice").
		WillReturnRows(rows)

	// QueryOne should return int64, not User
	id, err := QueryOne[int64](db, stmt)
	if err != nil {
		t.Fatalf("QueryOne returned error: %v", err)
	}

	if id != 42 {
		t.Fatalf("expected id=42, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInsertValuesReturningIdWithQueryOne(t *testing.T) {
	type OdooInstance struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		URL  string `db:"url"`
	}

	instance := OdooInstance{
		ID:   0, // Will be set by database
		Name: "Production",
		URL:  "https://example.odoo.com",
	}

	stmt := Insert[OdooInstance](nil).Values(instance).Returning("id")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock expects the INSERT to return the generated ID
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(42))
	mock.ExpectQuery(regexp.QuoteMeta(sqlStr)).
		WithArgs(int64(0), "Production", "https://example.odoo.com").
		WillReturnRows(rows)

	// QueryOne should successfully return int64
	id, err := QueryOne[int64](db, stmt)
	if err != nil {
		t.Fatalf("QueryOne returned error: %v", err)
	}

	if id != 42 {
		t.Fatalf("expected id=42, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInsertValuesReturningMultipleColumnsWithQueryOne(t *testing.T) {
	type User struct {
		ID        int64  `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	type InsertResult struct {
		ID        int64  `db:"id"`
		FirstName string `db:"first_name"`
	}

	user := User{
		ID:        0,
		FirstName: "John",
		LastName:  "Doe",
	}

	stmt := Insert[User](nil).Values(user).Returning("id", "first_name")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock expects the INSERT to return id and first_name
	rows := sqlmock.NewRows([]string{"id", "first_name"}).
		AddRow(int64(99), "John")
	mock.ExpectQuery(regexp.QuoteMeta(sqlStr)).
		WithArgs(int64(0), "John", "Doe").
		WillReturnRows(rows)

	// QueryOne should successfully return the InsertResult struct
	result, err := QueryOne[InsertResult](db, stmt)
	if err != nil {
		t.Fatalf("QueryOne returned error: %v", err)
	}

	if result.ID != 99 {
		t.Fatalf("expected id=99, got %d", result.ID)
	}

	if result.FirstName != "John" {
		t.Fatalf("expected first_name=John, got %s", result.FirstName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
