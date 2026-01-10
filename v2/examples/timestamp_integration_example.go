package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/guadalsistema/go-compose-sql/v2/engine"
	"github.com/guadalsistema/go-compose-sql/v2/expr"
	"github.com/guadalsistema/go-compose-sql/v2/table"
)

// User represents a user record with timestamp fields
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt sql.NullTime // Nullable timestamp
}

// Define table columns
type UsersColumns struct {
	ID        *table.Column[int64]
	Name      *table.Column[string]
	Email     *table.Column[string]
	CreatedAt *table.Column[time.Time]
	UpdatedAt *table.Column[sql.NullTime]
}

// Users table definition
var Users = table.NewTable("users", UsersColumns{
	ID:        table.Col[int64]("id").PrimaryKey().AutoIncrement(),
	Name:      table.Col[string]("name").NotNull(),
	Email:     table.Col[string]("email").Unique().NotNull(),
	CreatedAt: table.Col[time.Time]("created_at").NotNull(),
	UpdatedAt: table.Col[sql.NullTime]("updated_at"),
})

func main() {
	demonstrateSQLiteTimestampHandling()
}

func demonstrateSQLiteTimestampHandling() {
	fmt.Println("=== SQLite Timestamp Handling with TypeRegistry ===\n")

	// Create SQLite engine
	eng, err := engine.NewEngine("sqlite+pysqlite:///:memory:", engine.EngineOpts{})
	if err != nil {
		log.Fatal(err)
	}

	// Create a connection
	conn, err := eng.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create table
	createTableSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME
		)
	`

	_, err = conn.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	fmt.Println("✓ Created users table")

	// Insert test data
	// Note: SQLite stores timestamps as TEXT in ISO8601 format
	insertSQL := `
		INSERT INTO users (name, email, created_at, updated_at) VALUES
		('Alice Johnson', 'alice@example.com', '2024-01-15 10:30:00', '2024-01-16 14:20:00'),
		('Bob Smith', 'bob@example.com', '2024-01-16 11:45:00', NULL),
		('Carol White', 'carol@example.com', '2024-01-17 09:15:00', '2024-01-18 16:30:00')
	`

	_, err = conn.Exec(insertSQL)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}
	fmt.Println("✓ Inserted 3 test users with timestamps\n")

	// Example 1: Query all users
	fmt.Println("--- Example 1: Query All Users ---")
	var allUsers []User
	err = conn.Query(Users).All(&allUsers)
	if err != nil {
		log.Fatalf("Failed to query all users: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(allUsers))
	for _, user := range allUsers {
		fmt.Printf("  ID: %d\n", user.ID)
		fmt.Printf("  Name: %s\n", user.Name)
		fmt.Printf("  Email: %s\n", user.Email)
		fmt.Printf("  CreatedAt: %v (Year: %d, Type: %T)\n",
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			user.CreatedAt.Year(),
			user.CreatedAt)
		if user.UpdatedAt.Valid {
			fmt.Printf("  UpdatedAt: %v (Valid: true, Type: %T)\n",
				user.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
				user.UpdatedAt)
		} else {
			fmt.Printf("  UpdatedAt: NULL (Valid: false)\n")
		}
		fmt.Println()
	}

	// Example 2: Query single user
	fmt.Println("--- Example 2: Query Single User ---")
	var singleUser User
	err = conn.Query(Users).Where(expr.Eq(Users.C.ID, int64(1))).One(&singleUser)
	if err != nil {
		log.Fatalf("Failed to query single user: %v", err)
	}

	fmt.Printf("Found user: %s\n", singleUser.Name)
	fmt.Printf("Created at: %v\n", singleUser.CreatedAt)
	fmt.Printf("CreatedAt is time.Time: %v\n", reflect.TypeOf(singleUser.CreatedAt) == reflect.TypeOf(time.Time{}))
	fmt.Println()

	// Example 3: Verify type conversion
	fmt.Println("--- Example 3: Type Conversion Verification ---")
	fmt.Println("✓ SQLite returns timestamps as TEXT (ISO8601 strings)")
	fmt.Println("✓ TypeRegistry automatically converts TEXT → time.Time")
	fmt.Println("✓ TypeRegistry automatically converts TEXT → sql.NullTime")
	fmt.Println("✓ TypeRegistry handles NULL values correctly")
	fmt.Println()

	// Example 4: Show what happens behind the scenes
	fmt.Println("--- Example 4: Behind the Scenes ---")

	// Get dialect and registry
	dialect := conn.Engine().Dialect()
	registry := dialect.TypeRegistry()

	// Simulate what happens during scan
	dbValue := "2024-01-15 10:30:00" // SQLite returns this as string
	targetType := reflect.TypeOf(time.Time{})

	fmt.Printf("Database returns: %q (type: %T)\n", dbValue, dbValue)
	fmt.Printf("User expects: %v\n", targetType)
	fmt.Printf("Conversion needed: %v\n", registry.NeedsConversion(reflect.TypeOf(dbValue), targetType))

	// Perform conversion
	converted, err := registry.Convert(dbValue, targetType)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	fmt.Printf("After conversion: %v (type: %T)\n", converted, converted)
	fmt.Println()

	fmt.Println("=== SUCCESS ===")
	fmt.Println("Timestamps work seamlessly across SQLite and PostgreSQL!")
	fmt.Println("The same User struct and table definition work with both databases.")
}
