package sqlcompose

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestExec(t *testing.T) {
	type User struct {
		ID        int    `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	stmt := Insert[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := User{1, "Alice", "Smith"}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u.ID, u.FirstName, u.LastName).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt, u); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecPointer(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := Insert[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := &User{ID: 5, Name: "Bob"}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u.ID, u.Name).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt, u); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecMultiple(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := Insert[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u1 := User{ID: 1, Name: "Alice"}
	u2 := User{ID: 2, Name: "Bob"}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u1.ID, u1.Name).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u2.ID, u2.Name).
		WillReturnResult(sqlmock.NewResult(2, 1))

	if _, err := Exec(db, stmt, u1, u2); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecContext(t *testing.T) {
	type User struct {
		ID int `db:"id"`
	}

	stmt := Insert[User](nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := User{ID: 10}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := ExecContext(context.Background(), db, stmt, u); err != nil {
		t.Fatalf("ExecContext returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecReturning(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := Insert[User](nil).Returning("id")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := User{ID: 3, Name: "Alice"}

	if _, err := Exec(db, stmt, u); err == nil {
		t.Fatalf("expected error when using Exec with RETURNING clause")
	} else if !errors.Is(err, errors.New("sqlcompose: Exec cannot be used with RETURNING clause, use Query instead")) && err.Error() != "sqlcompose: Exec cannot be used with RETURNING clause, use Query instead" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecNonInsert(t *testing.T) {
	type User struct{ ID int }
	stmt := Select[User](nil)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := Exec(db, stmt, User{ID: 1}); err == nil {
		t.Fatalf("expected error for non-insert clause")
	}
}

func TestExecInvalidClause(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}
	stmt := Insert[User](nil)
	stmt.Clauses = append(stmt.Clauses, SqlClause{Type: ClauseType("BAD")})

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := Exec(db, stmt, User{ID: 1}); err == nil {
		t.Fatalf("expected error for invalid clause")
	} else {
		var clauseErr *ErrInvalidClause
		if !errors.As(err, &clauseErr) {
			t.Fatalf("expected ErrInvalidClause, got %v", err)
		}
	}
}

func TestExecMisplacedClause(t *testing.T) {
	type User struct {
		ID int `sql:"id"`
	}
	stmt := Insert[User](nil).Desc()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := Exec(db, stmt, User{ID: 1}); err == nil {
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

func TestExecUpdate(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
		City string `db:"city"`
	}

	stmt := Update[User](nil).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := User{ID: 1, Name: "Alice", City: "LA"}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u.ID, u.Name, u.City, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := Exec(db, stmt, u); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecUpdateWithFieldsOpt(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
		City string `db:"city"`
	}

	stmt := Update[User](&SqlOpts{Fields: []string{"name"}}).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	u := User{ID: 1, Name: "Alice", City: "LA"}

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(u.Name, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := Exec(db, stmt, u); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecDelete(t *testing.T) {
	type User struct {
		ID int `db:"id"`
	}

	stmt := Delete[User](nil).Where("id=?", 10)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecJoinQueryReturnsError(t *testing.T) {
	type User struct{ ID int }
	stmt := SQLStatement{
		Clauses: []SqlClause{
			{Type: ClauseSelect, TableName: "user", ColumnNames: []string{"id"}, ModelType: reflect.TypeOf(User{})},
			{Type: ClauseJoin, JoinStatement: Select[User](nil), Identifier: "u", Expr: "u.id = user.id"},
		},
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	if _, err := Exec(db, stmt, User{ID: 1}); err == nil {
		t.Fatalf("expected error for Exec with SELECT clause")
	}
}

func TestExecWithValues(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := Insert[User](nil).Values(42, "Alice")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(42, "Alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecWithValuesMultiple(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	stmt := Insert[User](nil).Values(1, "Alice")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(1, "Alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecWithValuesStruct(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	user := User{ID: 42, Name: "Alice"}
	stmt := Insert[User](nil).Values(user)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(42, "Alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecUpdateWithValuesStruct(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	user := User{ID: 1, Name: "Updated Name"}
	stmt := Update[User](nil).Values(user).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(1, "Updated Name", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateWithoutValuesOrModelFails(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	// This should fail because UPDATE needs either Values() or a model passed to Exec
	stmt := Update[User](nil).Where("id=?", 1)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Should fail with clear error message
	_, err = Exec(db, stmt)
	if err == nil {
		t.Fatalf("expected error for UPDATE without Values or model")
	}

	expectedErr := "sqlcompose: Exec requires at least one model"
	if err.Error() != expectedErr {
		t.Fatalf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestUpdateWithValuesWorks(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	user := User{ID: 1, Name: "Updated Name"}
	stmt := Update[User](nil).Values(user).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect UPDATE with values from struct + WHERE clause arg
	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(1, "Updated Name", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if _, err := Exec(db, stmt); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateWithModelWorks(t *testing.T) {
	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	user := User{ID: 1, Name: "Updated Name"}
	stmt := Update[User](nil).Where("id=?", 1)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlStr, err := stmt.Write()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect UPDATE with values from model passed to Exec + WHERE clause arg
	mock.ExpectExec(regexp.QuoteMeta(sqlStr)).
		WithArgs(1, "Updated Name", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Pass model to Exec (traditional approach)
	if _, err := Exec(db, stmt, user); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
