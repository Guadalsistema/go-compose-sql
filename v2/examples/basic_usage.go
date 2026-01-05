package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/guadalsistema/go-compose-sql/v2/engine"
	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/session"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// User represents a user record
type User struct {
	ID        int64
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
}

// Define table columns
type UsersColumns struct {
	ID        *table.Column[int64]
	Name      *table.Column[string]
	Email     *table.Column[string]
	Age       *table.Column[int]
	CreatedAt *table.Column[time.Time]
}

// Users table definition
var Users = table.NewTable("users", UsersColumns{
	ID:        table.Col[int64]("id").PrimaryKey().AutoIncrement(),
	Name:      table.Col[string]("name").NotNull(),
	Email:     table.Col[string]("email").Unique().NotNull(),
	Age:       table.Col[int]("age"),
	CreatedAt: table.Col[time.Time]("created_at").NotNull(),
})

func main() {
	// Create engine from connection URL (SQLAlchemy style)
	eng, err := engine.NewEngine("sqlite+pysqlite:///:memory:", engine.EngineConfig{
		Logger: slog.Default(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer eng.Close()

	// Create a session
	sess := session.NewSession(context.Background(), eng)
	defer sess.Close()

	// Example 1: Simple SELECT with WHERE
	fmt.Println("=== Example 1: Simple SELECT ===")
	query := sess.Query(Users).
		Where(expr.Eq(Users.C.ID, int64(1)))

	sql, args, _ := query.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE users.id = ?
	// Args: [1]

	// Example 2: SELECT with multiple WHERE conditions
	fmt.Println("=== Example 2: Multiple WHERE ===")
	query2 := sess.Query(Users).
		Where(expr.Gt(Users.C.Age, 18)).
		Where(expr.Like(Users.C.Email, "%@example.com"))

	sql, args, _ = query2.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE users.age > ? AND users.email LIKE ?
	// Args: [18 %@example.com]

	// Example 3: SELECT with OR conditions
	fmt.Println("=== Example 3: OR conditions ===")
	query3 := sess.Query(Users).
		Where(expr.Or(
			expr.Eq(Users.C.Name, "John"),
			expr.Eq(Users.C.Name, "Jane"),
		))

	sql, args, _ = query3.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE ((users.name = ?) OR (users.name = ?))
	// Args: [John Jane]

	// Example 4: SELECT with ORDER BY and LIMIT
	fmt.Println("=== Example 4: ORDER BY and LIMIT ===")
	query4 := sess.Query(Users).
		Where(expr.Gt(Users.C.Age, 21)).
		OrderByDesc("created_at").
		Limit(10)

	sql, args, _ = query4.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE users.age > ? ORDER BY created_at DESC LIMIT 10
	// Args: [21]

	// Example 5: SELECT specific columns
	fmt.Println("=== Example 5: Specific columns ===")
	query5 := sess.Query(Users).
		Select("id", "name", "email").
		Where(expr.IsNotNull(Users.C.Email))

	sql, args, _ = query5.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT id, name, email FROM users WHERE users.email IS NOT NULL

	// Example 6: SELECT with GROUP BY and HAVING
	fmt.Println("=== Example 6: GROUP BY and HAVING ===")
	query6 := sess.Query(Users).
		Select("age", "COUNT(*) as count").
		GroupBy("age").
		Having(expr.Raw("COUNT(*) > ?", 5))

	sql, args, _ = query6.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT age, COUNT(*) as count FROM users GROUP BY age HAVING COUNT(*) > ?
	// Args: [5]

	// Example 7: INSERT
	fmt.Println("=== Example 7: INSERT ===")
	insert := sess.Insert(Users).
		Set("name", "John Doe").
		Set("email", "john@example.com").
		Set("age", 30)

	sql, args, _ = insert.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: INSERT INTO users (name, email, age) VALUES (?, ?, ?)
	// Args: [John Doe john@example.com 30]

	// Example 8: UPDATE
	fmt.Println("=== Example 8: UPDATE ===")
	update := sess.Update(Users).
		Set("age", 31).
		Set("email", "john.doe@example.com").
		Where(expr.Eq(Users.C.ID, int64(1)))

	sql, args, _ = update.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: UPDATE users SET age = ?, email = ? WHERE users.id = ?
	// Args: [31 john.doe@example.com 1]

	// Example 9: DELETE
	fmt.Println("=== Example 9: DELETE ===")
	delete := sess.Delete(Users).
		Where(expr.Lt(Users.C.Age, 18))

	sql, args, _ = delete.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: DELETE FROM users WHERE users.age < ?
	// Args: [18]

	// Example 10: INSERT with RETURNING
	fmt.Println("=== Example 10: INSERT with RETURNING ===")
	insert2 := sess.Insert(Users).
		Set("name", "Jane Doe").
		Set("email", "jane@example.com").
		Set("age", 25).
		Returning("id", "created_at")

	sql, args, _ = insert2.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: INSERT INTO users (name, email, age) VALUES (?, ?, ?) RETURNING id, created_at
	// Args: [Jane Doe jane@example.com 25]

	// Example 11: IN clause
	fmt.Println("=== Example 11: IN clause ===")
	query7 := sess.Query(Users).
		Where(expr.In(Users.C.ID, int64(1), int64(2), int64(3)))

	sql, args, _ = query7.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE users.id IN (?, ?, ?)
	// Args: [1 2 3]

	// Example 12: BETWEEN clause
	fmt.Println("=== Example 12: BETWEEN clause ===")
	query8 := sess.Query(Users).
		Where(expr.Between(Users.C.Age, 18, 65))

	sql, args, _ = query8.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE users.age BETWEEN ? AND ?
	// Args: [18 65]

	// Example 13: Complex AND/OR combinations
	fmt.Println("=== Example 13: Complex conditions ===")
	query9 := sess.Query(Users).
		Where(expr.And(
			expr.Gt(Users.C.Age, 18),
			expr.Or(
				expr.Like(Users.C.Email, "%@gmail.com"),
				expr.Like(Users.C.Email, "%@yahoo.com"),
			),
		))

	sql, args, _ = query9.ToSQL()
	fmt.Printf("SQL: %s\nArgs: %v\n\n", sql, args)
	// Output: SELECT * FROM users WHERE ((users.age > ?) AND ((users.email LIKE ?) OR (users.email LIKE ?)))
	// Args: [18 %@gmail.com %@yahoo.com]

	// Example 14: Transaction
	fmt.Println("=== Example 14: Transaction ===")
	txSession := session.NewSession(context.Background(), eng)
	if err := txSession.Begin(); err != nil {
		log.Fatal(err)
	}

	// Perform operations in transaction
	_, err = txSession.Insert(Users).
		Set("name", "Transaction User").
		Set("email", "tx@example.com").
		Set("age", 28).
		Exec()

	if err != nil {
		txSession.Rollback()
		log.Fatal(err)
	}

	// Commit transaction
	err = txSession.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction committed successfully")
}
