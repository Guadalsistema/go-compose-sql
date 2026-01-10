package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/guadalsistema/go-compose-sql/v2/engine"
	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// User represents a user in our system
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

// UsersColumns defines the columns for the users table
type UsersColumns struct {
	ID        *table.Column[int64]
	Name      *table.Column[string]
	Email     *table.Column[string]
	CreatedAt *table.Column[time.Time]
	UpdatedAt *table.Column[sql.NullTime]
}

// Users is the table definition that works with both SQLite and PostgreSQL
var Users = table.NewTable("users", UsersColumns{
	ID:        table.Col[int64]("id").PrimaryKey().AutoIncrement(),
	Name:      table.Col[string]("name").NotNull(),
	Email:     table.Col[string]("email").Unique().NotNull(),
	CreatedAt: table.Col[time.Time]("created_at").NotNull(),
	UpdatedAt: table.Col[sql.NullTime]("updated_at"),
})

func main() {
	// Get database URL from environment, default to SQLite
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "sqlite://./example.db"
		fmt.Println("Using SQLite (default)")
	} else {
		fmt.Printf("Using database: %s\n", dbURL)
	}

	// Create engine
	eng, err := engine.NewEngine(dbURL, engine.EngineOpts{})
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// Connect
	conn, err := eng.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create table
	if err := createTable(conn); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Insert sample data
	fmt.Println("\n=== Inserting sample data ===")
	if err := insertSampleData(conn); err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	// Query all users
	fmt.Println("\n=== Querying all users ===")
	if err := queryAllUsers(conn); err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	// Query with filters
	fmt.Println("\n=== Querying with filters ===")
	if err := queryWithFilters(conn); err != nil {
		log.Fatalf("Failed to query filtered users: %v", err)
	}

	// Update a user
	fmt.Println("\n=== Updating a user ===")
	if err := updateUser(conn); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	// Transaction example
	fmt.Println("\n=== Transaction example ===")
	if err := transactionExample(conn); err != nil {
		log.Fatalf("Failed transaction: %v", err)
	}

	fmt.Println("\n=== Success! ===")
	fmt.Println("This same code worked with your database!")
	fmt.Println("Try switching between SQLite and PostgreSQL using DATABASE_URL environment variable.")
}

func createTable(conn *engine.Connection) error {
	// Note: In a real application, you'd use a migration tool
	// This is just for demonstration

	// Simplified: Try to create table with a generic SQL that works for most databases
	// For SQLite and PostgreSQL, the differences are minimal
	createSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP
		)`

	_, err := conn.Exec(createSQL)
	if err != nil {
		// If that fails, try PostgreSQL-specific syntax
		createSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP
		)`
		_, err = conn.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	fmt.Println("✓ Table created/verified")
	return nil
}

func insertSampleData(conn *engine.Connection) error {
	users := []struct {
		name  string
		email string
	}{
		{"Alice Johnson", "alice@example.com"},
		{"Bob Smith", "bob@example.com"},
		{"Carol White", "carol@example.com"},
	}

	for _, u := range users {
		_, err := conn.Insert(Users).
			Set("name", u.name).
			Set("email", u.email).
			Set("created_at", time.Now()).
			Exec()
		if err != nil {
			// Skip if user already exists (unique constraint)
			continue
		}
		fmt.Printf("✓ Inserted user: %s\n", u.name)
	}

	return nil
}

func queryAllUsers(conn *engine.Connection) error {
	var users []User
	err := conn.Query(Users).
		OrderBy("created_at").
		All(&users)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("  - %s <%s> (created: %s)\n",
			user.Name,
			user.Email,
			user.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func queryWithFilters(conn *engine.Connection) error {
	var users []User
	err := conn.Query(Users).
		Where(expr.Like(Users.C.Email, "%@example.com")).
		OrderBy("name").
		Limit(2).
		All(&users)
	if err != nil {
		return err
	}

	fmt.Printf("Users with @example.com email (limit 2):\n")
	for _, user := range users {
		fmt.Printf("  - %s <%s>\n", user.Name, user.Email)
	}

	return nil
}

func updateUser(conn *engine.Connection) error {
	// Update the first user's updated_at timestamp
	result, err := conn.Update(Users).
		Set("updated_at", time.Now()).
		Where(expr.Eq(Users.C.Email, "alice@example.com")).
		Exec()
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✓ Updated %d user(s)\n", rowsAffected)

	// Query to verify
	var user User
	err = conn.Query(Users).
		Where(expr.Eq(Users.C.Email, "alice@example.com")).
		One(&user)
	if err != nil {
		return err
	}

	if user.UpdatedAt.Valid {
		fmt.Printf("  UpdatedAt is now: %s\n", user.UpdatedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func transactionExample(conn *engine.Connection) error {
	// Start transaction
	if err := conn.Begin(); err != nil {
		return err
	}

	// Insert a new user
	_, err := conn.Insert(Users).
		Set("name", "Transaction User").
		Set("email", "tx@example.com").
		Set("created_at", time.Now()).
		Exec()
	if err != nil {
		conn.Rollback()
		return err
	}

	// Update existing user
	_, err = conn.Update(Users).
		Set("updated_at", time.Now()).
		Where(expr.Eq(Users.C.Email, "bob@example.com")).
		Exec()
	if err != nil {
		conn.Rollback()
		return err
	}

	// Commit transaction
	if err := conn.Commit(); err != nil {
		return err
	}

	fmt.Println("✓ Transaction committed successfully")
	return nil
}
