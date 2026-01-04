package sqlcompose

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

type OdooInstance struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	URL       string    `db:"url" json:"url"`
	Database  string    `db:"database" json:"database"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password" json:"password"`
	ClientID  int64     `db:"client_id" json:"client_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create client table first (referenced by foreign key)
	_, err = db.Exec(`
		CREATE TABLE client (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create client table: %v", err)
	}

	// Insert a test client
	_, err = db.Exec(`INSERT INTO client (id, name) VALUES (1, 'Test Client')`)
	if err != nil {
		t.Fatalf("failed to insert test client: %v", err)
	}

	// Create odoo_instance table
	_, err = db.Exec(`
		CREATE TABLE odoo_instance (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			url VARCHAR(255) NOT NULL,
			database VARCHAR(100) NOT NULL,
			username VARCHAR(100),
			password VARCHAR(255),
			client_id INTEGER REFERENCES client(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create odoo_instance table: %v", err)
	}

	return db
}

func TestIntegrationInsertValuesReturningID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instance := OdooInstance{
		Name:     "Production Instance",
		URL:      "https://prod.odoo.com",
		Database: "production_db",
		Username: "admin",
		Password: "secret123",
		ClientID: 1,
	}

	// Test INSERT with Values and RETURNING id
	// Exclude 'id' and 'created_at' as they are auto-generated
	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	query := Insert[OdooInstance](opts).Values(instance).Returning("id")
	
	idInstance, err := QueryOne[int64](db, query)
	if err != nil {
		t.Fatalf("INSERT with RETURNING failed: %v", err)
	}

	if idInstance <= 0 {
		t.Fatalf("expected positive id, got %d", idInstance)
	}

	t.Logf("Successfully inserted instance with ID: %d", idInstance)

	// Verify the record was actually inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM odoo_instance WHERE id = ?", idInstance).Scan(&count)
	if err != nil {
		t.Fatalf("failed to verify insert: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 record, found %d", count)
	}
}

func TestIntegrationInsertValuesReturningMultipleColumns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instance := OdooInstance{
		Name:     "Staging Instance",
		URL:      "https://staging.odoo.com",
		Database: "staging_db",
		Username: "admin",
		Password: "secret456",
		ClientID: 1,
	}

	type InsertResult struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		URL  string `db:"url"`
	}

	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	query := Insert[OdooInstance](opts).Values(instance).Returning("id", "name", "url")
	
	result, err := QueryOne[InsertResult](db, query)
	if err != nil {
		t.Fatalf("INSERT with RETURNING multiple columns failed: %v", err)
	}

	if result.ID <= 0 {
		t.Fatalf("expected positive id, got %d", result.ID)
	}
	if result.Name != "Staging Instance" {
		t.Fatalf("expected name 'Staging Instance', got %s", result.Name)
	}
	if result.URL != "https://staging.odoo.com" {
		t.Fatalf("expected url 'https://staging.odoo.com', got %s", result.URL)
	}

	t.Logf("Successfully inserted instance: ID=%d, Name=%s, URL=%s", result.ID, result.Name, result.URL)
}

func TestIntegrationUpdateWithValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// First, insert a record
	instance := OdooInstance{
		Name:     "Test Instance",
		URL:      "https://test.odoo.com",
		Database: "test_db",
		Username: "admin",
		Password: "secret",
		ClientID: 1,
	}

	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	insertQuery := Insert[OdooInstance](opts).Values(instance).Returning("id")
	id, err := QueryOne[int64](db, insertQuery)
	if err != nil {
		t.Fatalf("failed to insert initial record: %v", err)
	}

	t.Logf("Inserted record with ID: %d", id)

	// Now update it using Values
	updatedInstance := OdooInstance{
		ID:       id,
		Name:     "Updated Instance",
		URL:      "https://updated.odoo.com",
		Database: "updated_db",
		Username: "newadmin",
		Password: "newsecret",
		ClientID: 1,
	}

	updateQuery := Update[OdooInstance](nil).Values(updatedInstance).Where("id=?", id)
	res, err := Exec(db, updateQuery)
	if err != nil {
		t.Fatalf("UPDATE with Values failed: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}

	// Verify the update
	var name, url, database string
	err = db.QueryRow("SELECT name, url, database FROM odoo_instance WHERE id = ?", id).Scan(&name, &url, &database)
	if err != nil {
		t.Fatalf("failed to verify update: %v", err)
	}

	if name != "Updated Instance" {
		t.Fatalf("expected name 'Updated Instance', got %s", name)
	}
	if url != "https://updated.odoo.com" {
		t.Fatalf("expected url 'https://updated.odoo.com', got %s", url)
	}
	if database != "updated_db" {
		t.Fatalf("expected database 'updated_db', got %s", database)
	}

	t.Logf("Successfully updated instance ID %d", id)
}

func TestIntegrationUpdateWithModel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// First, insert a record
	instance := OdooInstance{
		Name:     "Original Instance",
		URL:      "https://original.odoo.com",
		Database: "original_db",
		Username: "admin",
		Password: "secret",
		ClientID: 1,
	}

	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	insertQuery := Insert[OdooInstance](opts).Values(instance).Returning("id")
	id, err := QueryOne[int64](db, insertQuery)
	if err != nil {
		t.Fatalf("failed to insert initial record: %v", err)
	}

	// Update using traditional approach (passing model to Exec)
	updatedInstance := OdooInstance{
		ID:       id,
		Name:     "Modified Instance",
		URL:      "https://modified.odoo.com",
		Database: "modified_db",
		Username: "modadmin",
		Password: "modsecret",
		ClientID: 1,
	}

	updateQuery := Update[OdooInstance](nil).Where("id=?", id)
	res, err := Exec(db, updateQuery, updatedInstance)
	if err != nil {
		t.Fatalf("UPDATE with model failed: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}

	// Verify the update
	var name string
	err = db.QueryRow("SELECT name FROM odoo_instance WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("failed to verify update: %v", err)
	}

	if name != "Modified Instance" {
		t.Fatalf("expected name 'Modified Instance', got %s", name)
	}

	t.Logf("Successfully updated instance ID %d using model", id)
}

func TestIntegrationSelectWithQueryOne(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test record
	instance := OdooInstance{
		Name:     "Query Test Instance",
		URL:      "https://querytest.odoo.com",
		Database: "querytest_db",
		Username: "admin",
		Password: "secret",
		ClientID: 1,
	}

	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	insertQuery := Insert[OdooInstance](opts).Values(instance).Returning("id")
	id, err := QueryOne[int64](db, insertQuery)
	if err != nil {
		t.Fatalf("failed to insert record: %v", err)
	}

	// Query it back
	selectQuery := Select[OdooInstance](nil).Where("id=?", id)
	result, err := QueryOne[OdooInstance](db, selectQuery)
	if err != nil {
		t.Fatalf("SELECT with QueryOne failed: %v", err)
	}

	if result.ID != id {
		t.Fatalf("expected id %d, got %d", id, result.ID)
	}
	if result.Name != "Query Test Instance" {
		t.Fatalf("expected name 'Query Test Instance', got %s", result.Name)
	}
	if result.URL != "https://querytest.odoo.com" {
		t.Fatalf("expected url 'https://querytest.odoo.com', got %s", result.URL)
	}

	t.Logf("Successfully queried instance: %+v", result)
}

func TestIntegrationDebugSQL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instance := OdooInstance{
		Name:     "Debug Instance",
		URL:      "https://debug.odoo.com",
		Database: "debug_db",
		Username: "admin",
		Password: "secret",
		ClientID: 1,
	}

	opts := &SqlOpts{
		Fields: []string{"name", "url", "database", "username", "password", "client_id"},
	}
	query := Insert[OdooInstance](opts).Values(instance).Returning("id")

	sql, err := query.Write()
	if err != nil {
		t.Fatalf("failed to write SQL: %v", err)
	}
	
	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", query.Args())
	
	// Try executing manually to see what happens
	rows, err := db.Query(sql, query.Args()...)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()
	
	if rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		t.Logf("Returned ID: %d", id)
	} else {
		t.Fatal("no rows returned")
	}
}
