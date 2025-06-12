package migrations

import (
	"database/sql"
	"testing"
	"time"
)

// TestPostgreSQLIntegration tests PostgreSQL-specific functionality
func TestPostgreSQLIntegration(t *testing.T) {
	PostgreSQLIntegrationTest(t, func(t *testing.T, db *sql.DB) {
		migrator := SetupAutoMigrationTestWithDB(t, db)

		// Test PostgreSQL-specific features
		t.Run("PostgreSQL_AutoMigration", func(t *testing.T) {
			err := migrator.MigrateModels(&AutoTestUser{}, &AutoTestProduct{})
			if err != nil {
				t.Fatalf("Failed to migrate models in PostgreSQL: %v", err)
			}

			// Verify tables were created
			var userTableExists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_name = 'auto_test_user'
				)
			`).Scan(&userTableExists)
			if err != nil {
				t.Fatalf("Failed to check table existence: %v", err)
			}
			if !userTableExists {
				t.Error("Expected auto_test_user table to exist in PostgreSQL")
			}
		})

		t.Run("PostgreSQL_DataTypes", func(t *testing.T) {
			// Test PostgreSQL-specific data types
			err := migrator.createMigrationsTable()
			if err != nil {
				t.Fatalf("Failed to create migrations table: %v", err)
			}

			// Test with PostgreSQL-specific model
			type PostgreSQLTestModel struct {
				ID       int64     `db:"id" json:"id"`
				UUID     string    `db:"uuid" json:"uuid" sql:"UUID"`
				JSONData string    `db:"json_data" json:"json_data" sql:"JSONB"`
				Created  time.Time `db:"created_at" json:"created_at" sql:"TIMESTAMP WITH TIME ZONE"`
			}

			err = migrator.migrateModel(&PostgreSQLTestModel{})
			if err != nil {
				t.Fatalf("Failed to migrate PostgreSQL-specific model: %v", err)
			}

			// Verify table structure
			var columnCount int
			err = db.QueryRow(`
				SELECT COUNT(*) FROM information_schema.columns 
				WHERE table_name = 'postgre_s_q_l_test_model'
			`).Scan(&columnCount)
			if err != nil {
				t.Fatalf("Failed to count columns: %v", err)
			}
			if columnCount == 0 {
				t.Error("Expected columns to be created for PostgreSQL model")
			}
		})

		t.Run("PostgreSQL_Transactions", func(t *testing.T) {
			// Test transaction handling in PostgreSQL
			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}
			defer tx.Rollback()

			// Create table in transaction
			_, err = tx.Exec(`
				CREATE TABLE transaction_test (
					id SERIAL PRIMARY KEY,
					name VARCHAR(100) NOT NULL
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create table in transaction: %v", err)
			}

			// Test rollback
			err = tx.Rollback()
			if err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}

			// Verify table doesn't exist after rollback
			var tableExists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_name = 'transaction_test'
				)
			`).Scan(&tableExists)
			if err != nil {
				t.Fatalf("Failed to check table existence after rollback: %v", err)
			}
			if tableExists {
				t.Error("Table should not exist after transaction rollback")
			}
		})
	})
}

// TestDatabaseDriverCompatibility tests compatibility across different database drivers
func TestDatabaseDriverCompatibility(t *testing.T) {
	DatabaseDriverSpecificTest(t, func(t *testing.T, driver string, db *sql.DB) {
		migrator := SetupAutoMigrationTestWithDB(t, db)

		t.Logf("Testing with driver: %s", driver)

		// Test basic migration functionality
		err := migrator.MigrateModels(&AutoTestUser{})
		if err != nil {
			t.Fatalf("Failed to migrate models with %s: %v", driver, err)
		}

		// Test table naming consistency
		tableName := migrator.getTableName(&AutoTestUser{})
		if tableName != "auto_test_user" {
			t.Errorf("Table name mismatch for %s: expected 'auto_test_user', got '%s'", driver, tableName)
		}

		// Test column detection
		columns, err := migrator.getCurrentTableColumns("auto_test_user")
		if err != nil {
			t.Fatalf("Failed to get table columns with %s: %v", driver, err)
		}

		expectedColumns := []string{"id", "email", "name", "is_active", "created_at"}
		for _, col := range expectedColumns {
			if _, exists := columns[col]; !exists {
				t.Errorf("Column '%s' missing in %s", col, driver)
			}
		}

		// Test schema update
		err = migrator.migrateModel(&AutoTestUser{})
		if err != nil {
			t.Fatalf("Failed to re-migrate model with %s: %v", driver, err)
		}
	})
}

// TestPostgreSQLSpecificFeatures tests PostgreSQL-only features
func TestPostgreSQLSpecificFeatures(t *testing.T) {
	PostgreSQLIntegrationTest(t, func(t *testing.T, db *sql.DB) {
		t.Run("Serial_Primary_Keys", func(t *testing.T) {
			// Test SERIAL primary key generation
			_, err := db.Exec(`
				CREATE TABLE serial_test (
					id SERIAL PRIMARY KEY,
					name VARCHAR(100) NOT NULL
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create table with SERIAL: %v", err)
			}

			// Insert test data
			var id int
			err = db.QueryRow("INSERT INTO serial_test (name) VALUES ('test') RETURNING id").Scan(&id)
			if err != nil {
				t.Fatalf("Failed to insert with SERIAL: %v", err)
			}

			if id == 0 {
				t.Error("Expected SERIAL to generate non-zero ID")
			}
		})

		t.Run("JSON_Support", func(t *testing.T) {
			// Test JSON/JSONB support
			_, err := db.Exec(`
				CREATE TABLE json_test (
					id SERIAL PRIMARY KEY,
					data JSONB NOT NULL
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create table with JSONB: %v", err)
			}

			// Insert JSON data
			_, err = db.Exec("INSERT INTO json_test (data) VALUES ($1)", `{"key": "value", "number": 42}`)
			if err != nil {
				t.Fatalf("Failed to insert JSON data: %v", err)
			}

			// Query JSON data
			var jsonData string
			err = db.QueryRow("SELECT data->>'key' FROM json_test WHERE data->>'number' = '42'").Scan(&jsonData)
			if err != nil {
				t.Fatalf("Failed to query JSON data: %v", err)
			}

			if jsonData != "value" {
				t.Errorf("Expected JSON query to return 'value', got '%s'", jsonData)
			}
		})

		t.Run("Arrays_Support", func(t *testing.T) {
			// Test array support
			_, err := db.Exec(`
				CREATE TABLE array_test (
					id SERIAL PRIMARY KEY,
					tags TEXT[] NOT NULL
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create table with arrays: %v", err)
			}

			// Insert array data
			_, err = db.Exec("INSERT INTO array_test (tags) VALUES ($1)", `{"tag1", "tag2", "tag3"}`)
			if err != nil {
				t.Fatalf("Failed to insert array data: %v", err)
			}

			// Query array data
			var arrayLength int
			err = db.QueryRow("SELECT array_length(tags, 1) FROM array_test").Scan(&arrayLength)
			if err != nil {
				t.Fatalf("Failed to query array length: %v", err)
			}

			if arrayLength != 3 {
				t.Errorf("Expected array length 3, got %d", arrayLength)
			}
		})

		t.Run("Indexes_And_Constraints", func(t *testing.T) {
			// Test advanced indexing
			_, err := db.Exec(`
				CREATE TABLE index_test (
					id SERIAL PRIMARY KEY,
					email VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(100) NOT NULL,
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create table with constraints: %v", err)
			}
			// Create partial index with proper immutable condition
			_, err = db.Exec("CREATE INDEX idx_active_users ON index_test (name) WHERE name IS NOT NULL")
			if err != nil {
				t.Fatalf("Failed to create partial index: %v", err)
			}

			// Verify index exists
			var indexExists bool
			err = db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM pg_indexes 
					WHERE indexname = 'idx_active_users'
				)
			`).Scan(&indexExists)
			if err != nil {
				t.Fatalf("Failed to check index existence: %v", err)
			}

			if !indexExists {
				t.Error("Expected partial index to be created")
			}
		})
	})
}

// TestHighVolumeDataMigration tests migration performance with larger datasets
func TestHighVolumeDataMigration(t *testing.T) {
	PostgreSQLIntegrationTest(t, func(t *testing.T, db *sql.DB) {
		migrator := SetupAutoMigrationTestWithDB(t, db)

		// Create a model with more fields
		type LargeTestModel struct {
			ID        int64     `db:"id" json:"id"`
			Field1    string    `db:"field1" json:"field1"`
			Field2    string    `db:"field2" json:"field2"`
			Field3    string    `db:"field3" json:"field3"`
			Field4    string    `db:"field4" json:"field4"`
			Field5    string    `db:"field5" json:"field5"`
			Field6    int       `db:"field6" json:"field6"`
			Field7    int       `db:"field7" json:"field7"`
			Field8    bool      `db:"field8" json:"field8"`
			Field9    bool      `db:"field9" json:"field9"`
			Field10   float64   `db:"field10" json:"field10"`
			CreatedAt time.Time `db:"created_at" json:"created_at"`
			UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
		}

		start := time.Now()

		// Migrate the large model
		err := migrator.MigrateModels(&LargeTestModel{})
		if err != nil {
			t.Fatalf("Failed to migrate large model: %v", err)
		}

		migrationTime := time.Since(start)
		t.Logf("Large model migration took: %v", migrationTime)

		// Verify all columns were created
		columns, err := migrator.getCurrentTableColumns("large_test_model")
		if err != nil {
			t.Fatalf("Failed to get columns for large model: %v", err)
		}

		expectedColumnCount := 12 // Number of fields in LargeTestModel
		if len(columns) < expectedColumnCount {
			t.Errorf("Expected at least %d columns, got %d", expectedColumnCount, len(columns))
		}

		// Test inserting data to verify table works
		_, err = db.Exec(`
			INSERT INTO large_test_model (
				field1, field2, field3, field4, field5,
				field6, field7, field8, field9, field10,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, "test1", "test2", "test3", "test4", "test5",
			1, 2, true, false, 3.14,
			time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to insert data into large model: %v", err)
		}

		// Verify data was inserted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM large_test_model").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, got %d", count)
		}
	})
}
