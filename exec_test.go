package sqlcompose

import (
	"context"
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

	mock.ExpectExec(regexp.QuoteMeta(stmt.Write())).
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

	mock.ExpectExec(regexp.QuoteMeta(stmt.Write())).
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

	mock.ExpectExec(regexp.QuoteMeta(stmt.Write())).
		WithArgs(u1.ID, u1.Name).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(stmt.Write())).
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

	mock.ExpectExec(regexp.QuoteMeta(stmt.Write())).
		WithArgs(u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := ExecContext(context.Background(), db, stmt, u); err != nil {
		t.Fatalf("ExecContext returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
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
