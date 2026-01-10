//go:build sqlite

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

func TestIntegrationInsertValuesRespectsFieldsOption(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instance := OdooInstance{
		ID:        999, // This should NOT be inserted
		Name:      "Partial Instance",
		URL:       "https://partial.odoo.com",
		Database:  "partial_db",
		Username:  "admin",
		Password:  "secret",
		ClientID:  1,
		CreatedAt: time.Now(), // This should NOT be inserted
	}

	// Only insert specific fields - excluding id and created_at
	opts := &SqlOpts{
		Fields: []string{"name", "url", "database"},
	}
	// Only return the fields we actually inserted (plus id)
	query := Insert[OdooInstance](opts).Values(instance).Returning("id", "name", "url", "database")

	type PartialResult struct {
		ID       int64  `db:"id"`
		Name     string `db:"name"`
		URL      string `db:"url"`
		Database string `db:"database"`
	}

	result, err := QueryOne[PartialResult](db, query)
	if err != nil {
		t.Fatalf("INSERT with partial fields failed: %v", err)
	}

	// Verify the specified fields were inserted
	if result.Name != "Partial Instance" {
		t.Fatalf("expected name 'Partial Instance', got %s", result.Name)
	}
	if result.URL != "https://partial.odoo.com" {
		t.Fatalf("expected url 'https://partial.odoo.com', got %s", result.URL)
	}
	if result.Database != "partial_db" {
		t.Fatalf("expected database 'partial_db', got %s", result.Database)
	}

	// Verify the excluded fields were NOT inserted (should be default/empty)
	if result.ID == 999 {
		t.Fatalf("id should be auto-generated, not 999")
	}

	// Query the database directly to verify excluded fields are NULL/empty
	var username, password sql.NullString
	err = db.QueryRow("SELECT username, password FROM odoo_instance WHERE id = ?", result.ID).Scan(&username, &password)
	if err != nil {
		t.Fatalf("failed to verify excluded fields: %v", err)
	}

	if username.Valid && username.String != "" {
		t.Fatalf("username should be NULL (not inserted), got %s", username.String)
	}
	if password.Valid && password.String != "" {
		t.Fatalf("password should be NULL (not inserted), got %s", password.String)
	}

	t.Logf("Successfully inserted only specified fields: name=%s, url=%s, database=%s", result.Name, result.URL, result.Database)
	t.Logf("Excluded fields correctly defaulted: id=%d, username=NULL, password=NULL", result.ID)
}

func TestIntegrationUpdateValuesRespectsFieldsOption(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// First, insert a complete record
	instance := OdooInstance{
		Name:     "Original Instance",
		URL:      "https://original.odoo.com",
		Database: "original_db",
		Username: "originaluser",
		Password: "originalpass",
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

	// Now update only specific fields using Values
	updatedInstance := OdooInstance{
		ID:       id,
		Name:     "Updated Name",         // Should be updated
		URL:      "https://new.odoo.com", // Should be updated
		Database: "should_not_change",    // Should NOT be updated (not in Fields)
		Username: "should_not_change",    // Should NOT be updated (not in Fields)
		Password: "newpass",              // Should be updated
		ClientID: 1,
	}

	// Only update name, url, and password - exclude database and username
	updateOpts := &SqlOpts{
		Fields: []string{"name", "url", "password"},
	}
	updateQuery := Update[OdooInstance](updateOpts).Values(updatedInstance).Where("id=?", id)
	res, err := Exec(db, updateQuery)
	if err != nil {
		t.Fatalf("UPDATE with partial fields failed: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}

	// Verify only the specified fields were updated
	var name, url, database, username, password string
	err = db.QueryRow("SELECT name, url, database, username, password FROM odoo_instance WHERE id = ?", id).
		Scan(&name, &url, &database, &username, &password)
	if err != nil {
		t.Fatalf("failed to verify update: %v", err)
	}

	// Check updated fields
	if name != "Updated Name" {
		t.Fatalf("expected name 'Updated Name', got %s", name)
	}
	if url != "https://new.odoo.com" {
		t.Fatalf("expected url 'https://new.odoo.com', got %s", url)
	}
	if password != "newpass" {
		t.Fatalf("expected password 'newpass', got %s", password)
	}

	// Check that excluded fields were NOT updated
	if database != "original_db" {
		t.Fatalf("database should remain 'original_db', got %s", database)
	}
	if username != "originaluser" {
		t.Fatalf("username should remain 'originaluser', got %s", username)
	}

	t.Logf("Successfully updated only specified fields:")
	t.Logf("  Updated: name=%s, url=%s, password=%s", name, url, password)
	t.Logf("  Unchanged: database=%s, username=%s", database, username)
}

func TestIntegrationValuesWithFieldsDebug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	instance := OdooInstance{
		ID:       999,
		Name:     "Test",
		URL:      "https://test.com",
		Database: "test_db",
		Username: "admin",
		Password: "secret",
		ClientID: 1,
	}

	// Test INSERT with partial fields
	insertOpts := &SqlOpts{
		Fields: []string{"name", "url"},
	}
	insertQuery := Insert[OdooInstance](insertOpts).Values(instance)

	sql, err := insertQuery.Write()
	if err != nil {
		t.Fatalf("failed to write SQL: %v", err)
	}

	t.Logf("INSERT SQL: %s", sql)
	t.Logf("INSERT Args: %v", insertQuery.Args())

	// Verify only 2 fields are in the SQL and args
	expectedSQL := "INSERT INTO odoo_instance (name, url) VALUES (?, ?);"
	if sql != expectedSQL {
		t.Fatalf("expected SQL %q, got %q", expectedSQL, sql)
	}

	args := insertQuery.Args()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "Test" || args[1] != "https://test.com" {
		t.Fatalf("expected args [Test, https://test.com], got %v", args)
	}

	// Test UPDATE with partial fields
	updateOpts := &SqlOpts{
		Fields: []string{"name", "database"},
	}
	updateQuery := Update[OdooInstance](updateOpts).Values(instance).Where("id=?", 1)

	sql, err = updateQuery.Write()
	if err != nil {
		t.Fatalf("failed to write SQL: %v", err)
	}

	t.Logf("UPDATE SQL: %s", sql)
	t.Logf("UPDATE Args: %v", updateQuery.Args())

	// Verify only 2 fields are in the UPDATE SET clause
	expectedSQL = "UPDATE odoo_instance SET name=?, database=? WHERE id=?;"
	if sql != expectedSQL {
		t.Fatalf("expected SQL %q, got %q", expectedSQL, sql)
	}

	args = updateQuery.Args()
	if len(args) != 3 { // 2 from Values + 1 from WHERE
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != "Test" || args[1] != "test_db" || args[2] != 1 {
		t.Fatalf("expected args [Test, test_db, 1], got %v", args)
	}

	t.Log("Fields option correctly filters which struct fields are used in Values()")
}

func TestIntegrationInsertPartialFieldsReturningID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a similar table to tiktok_order_sync_log
	_, err := db.Exec(`
		CREATE TABLE tiktok_order_sync_log (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			client_id INTEGER NOT NULL,
			last_success_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	type TiktokOrderSyncLog struct {
		ID            int64     `db:"id"`
		ClientID      int64     `db:"client_id"`
		LastSuccessAt time.Time `db:"last_success_at"`
		CreatedAt     time.Time `db:"created_at"`
	}

	rec := TiktokOrderSyncLog{
		ClientID:      1,
		LastSuccessAt: time.Now(),
	}

	// Only insert specific fields - same as your code
	opts := &SqlOpts{
		Fields: []string{"client_id", "last_success_at"},
	}
	stmt := Insert[TiktokOrderSyncLog](opts).Values(rec).Returning("id")

	// Debug: Show the SQL
	sql, _ := stmt.Write()
	t.Logf("Generated SQL: %s", sql)
	t.Logf("Args: %v", stmt.Args())

	// Try to execute
	id, err := QueryOne[int64](db, stmt)
	if err != nil {
		t.Fatalf("INSERT with RETURNING id failed: %v", err)
	}

	t.Logf("Returned ID: %d", id)

	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

func TestIntegrationInsertWithNullableIDConditionalFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create tiktok_order_sync_log table with nullable columns
	_, err := db.Exec(`
		CREATE TABLE tiktok_order_sync_log (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			client_id INTEGER NOT NULL,
			last_success_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	type TiktokOrderSyncLog struct {
		ID            sql.NullInt64 `db:"id"`
		ClientID      int64         `db:"client_id"`
		LastSuccessAt sql.NullTime  `db:"last_success_at"`
		CreatedAt     sql.NullTime  `db:"created_at"`
		UpdatedAt     sql.NullTime  `db:"updated_at"`
	}

	// Test case 1: Insert without ID (auto-increment)
	t.Run("AutoIncrementID", func(t *testing.T) {
		rec := TiktokOrderSyncLog{
			ID:            sql.NullInt64{Valid: false}, // No ID provided
			ClientID:      1,
			LastSuccessAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		opts := &SqlOpts{
			Fields: []string{"client_id", "last_success_at"},
		}
		stmt := Insert[TiktokOrderSyncLog](opts).Values(rec).Returning("id")

		// RETURNING always returns a non-NULL value for auto-increment, so use int64
		id, err := QueryOne[int64](db, stmt)
		if err != nil {
			t.Fatalf("INSERT with RETURNING id failed: %v", err)
		}

		if id <= 0 {
			t.Fatalf("expected positive id, got %d", id)
		}

		t.Logf("Successfully inserted record with auto-generated ID: %d", id)

		// Verify the record exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM tiktok_order_sync_log WHERE id = ?", id).Scan(&count)
		if err != nil {
			t.Fatalf("failed to verify insert: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 record, found %d", count)
		}
	})

	// Test case 2: Insert WITH specific ID (conditional field inclusion)
	t.Run("SpecificID", func(t *testing.T) {
		specificID := int64(100)
		rec := TiktokOrderSyncLog{
			ID:            sql.NullInt64{Int64: specificID, Valid: true},
			ClientID:      2,
			LastSuccessAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		// Conditionally add 'id' to fields if it's valid
		opts := &SqlOpts{
			Fields: []string{"client_id", "last_success_at"},
		}
		if rec.ID.Valid && rec.ID.Int64 > 0 {
			opts.Fields = append(opts.Fields, "id")
		}

		// When providing a specific ID, we don't need RETURNING - we already know the ID
		stmt := Insert[TiktokOrderSyncLog](opts).Values(rec)

		_, err := Exec(db, stmt)
		if err != nil {
			t.Fatalf("INSERT with specific ID failed: %v", err)
		}

		t.Logf("Successfully inserted record with specific ID: %d", specificID)

		// Verify the record exists with the specific ID
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM tiktok_order_sync_log WHERE id = ?", specificID).Scan(&count)
		if err != nil {
			t.Fatalf("failed to verify insert: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 record, found %d", count)
		}
	})

	// Test case 3: Verify the Create repository pattern works
	t.Run("RepositoryPattern", func(t *testing.T) {
		// Simulate the Create function from your repository
		createFunc := func(rec TiktokOrderSyncLog) (int64, error) {
			opts := &SqlOpts{
				Fields: []string{"client_id", "last_success_at"},
			}
			if rec.ID.Valid && rec.ID.Int64 > 0 {
				// If ID is provided, include it in the insert and return it directly
				opts.Fields = append(opts.Fields, "id")
				stmt := Insert[TiktokOrderSyncLog](opts).Values(rec)
				_, err := Exec(db, stmt)
				if err != nil {
					return 0, err
				}
				return rec.ID.Int64, nil
			}
			// If no ID, use RETURNING to get auto-generated ID
			stmt := Insert[TiktokOrderSyncLog](opts).Values(rec).Returning("id")
			id, err := QueryOne[int64](db, stmt)
			if err != nil {
				return 0, err
			}
			return id, nil
		}

		// Test with auto-increment
		rec1 := TiktokOrderSyncLog{
			ClientID:      3,
			LastSuccessAt: sql.NullTime{Time: time.Now(), Valid: true},
		}
		id1, err := createFunc(rec1)
		if err != nil {
			t.Fatalf("createFunc with auto-increment failed: %v", err)
		}
		if id1 <= 0 {
			t.Fatalf("expected positive ID, got %d", id1)
		}
		t.Logf("Repository pattern (auto-increment): ID=%d", id1)

		// Test with specific ID
		rec2 := TiktokOrderSyncLog{
			ID:            sql.NullInt64{Int64: 200, Valid: true},
			ClientID:      4,
			LastSuccessAt: sql.NullTime{Time: time.Now(), Valid: true},
		}
		id2, err := createFunc(rec2)
		if err != nil {
			t.Fatalf("createFunc with specific ID failed: %v", err)
		}
		if id2 != 200 {
			t.Fatalf("expected ID 200, got %d", id2)
		}
		t.Logf("Repository pattern (specific ID): ID=%d", id2)
	})
}
