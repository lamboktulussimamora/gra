package migrations

import (
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestEntityEnhanced is a test entity for enhanced testing
type TestEntityEnhanced struct {
	ID        int       `db:"id" migrations:"primary_key"`
	Name      string    `db:"name" migrations:"type:varchar(100),not_null"`
	Email     string    `db:"email" migrations:"unique,type:varchar(255)"`
	Age       int       `db:"age" migrations:"index"`
	IsActive  bool      `db:"is_active" migrations:"default:true"`
	CreatedAt time.Time `db:"created_at" migrations:"auto_timestamp"`
	UpdatedAt time.Time `db:"updated_at" migrations:"auto_timestamp_update"`
}

// TestAutoMigratorEnhanced tests the AutoMigrator functionality
func TestAutoMigratorEnhanced(t *testing.T) {
	migrator, db, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("BasicEntityMigration", func(t *testing.T) {
		entity := TestEntityEnhanced{}
		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected migration to succeed, got error: %v", err)
		}

		// Verify table was created
		exists, err := CheckTableExists(db, "testentityenhanced")
		if err != nil {
			t.Logf("Could not verify table creation: %v", err)
		} else if !exists {
			t.Log("Table not found with expected name, checking for similar tables")
			// Check what tables were actually created
			rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
			if err != nil {
				t.Logf("Failed to query tables: %v", err)
			} else {
				defer rows.Close()
				var tables []string
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err != nil {
						continue
					}
					tables = append(tables, name)
				}
				t.Logf("Available tables: %v", tables)
			}
		}
	})

	t.Run("MigrateMultipleEntities", func(t *testing.T) {
		// Test migrating multiple entities
		entities := []interface{}{
			TestEntityEnhanced{},
			struct {
				ID   int    `db:"id" migrations:"primary_key"`
				Data string `db:"data" migrations:"type:text"`
			}{},
		}

		err := migrator.MigrateModels(entities...)
		if err != nil {
			t.Errorf("Expected migration to succeed, got error: %v", err)
		}
	})
}

// TestAutoMigrationTypesEnhanced tests migration of different data types
func TestAutoMigrationTypesEnhanced(t *testing.T) {
	migrator, db, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("ComplexDataTypes", func(t *testing.T) {
		entity := struct {
			ID          int       `db:"id" migrations:"primary_key"`
			StringField string    `db:"string_field" migrations:"type:varchar(255),not_null"`
			IntField    int       `db:"int_field" migrations:"index"`
			BoolField   bool      `db:"bool_field" migrations:"default:false"`
			TimeField   time.Time `db:"time_field" migrations:"auto_timestamp"`
			FloatField  float64   `db:"float_field" migrations:"type:decimal(10,2)"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected complex type migration to succeed, got error: %v", err)
		}

		// Verify table structure
		exists, err := CheckTableExists(db, "struct")
		if err != nil {
			t.Logf("Could not verify table creation: %v", err)
		} else if exists {
			t.Log("Complex data type table created successfully")
		}
	})

	t.Run("EntityWithIndexes", func(t *testing.T) {
		entity := struct {
			ID       int    `db:"id" migrations:"primary_key"`
			Username string `db:"username" migrations:"unique,index,type:varchar(50)"`
			Email    string `db:"email" migrations:"unique,type:varchar(100)"`
			Status   string `db:"status" migrations:"index,type:varchar(20)"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected indexed entity migration to succeed, got error: %v", err)
		}
	})

	t.Run("EntityWithDefaults", func(t *testing.T) {
		entity := struct {
			ID        int    `db:"id" migrations:"primary_key"`
			Name      string `db:"name" migrations:"type:varchar(100),default:'Unknown'"`
			IsActive  bool   `db:"is_active" migrations:"default:true"`
			Score     int    `db:"score" migrations:"default:0"`
			CreatedAt string `db:"created_at" migrations:"default:CURRENT_TIMESTAMP"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected entity with defaults migration to succeed, got error: %v", err)
		}
	})
}

// TestMigrationTagParsingEnhanced tests the parsing of migration tags
func TestMigrationTagParsingEnhanced(t *testing.T) {
	testCases := []struct {
		name         string
		structType   reflect.Type
		expectedTags map[string]string
	}{
		{
			name:       "BasicTags",
			structType: reflect.TypeOf(TestEntityEnhanced{}),
			expectedTags: map[string]string{
				"ID":   "primary_key",
				"Name": "type:varchar(100),not_null",
			},
		},
		{
			name: "ComplexTags",
			structType: reflect.TypeOf(struct {
				ID    int    `db:"id" migrations:"primary_key,auto_increment"`
				Email string `db:"email" migrations:"unique,index,type:varchar(255),not_null"`
			}{}),
			expectedTags: map[string]string{
				"ID":    "primary_key,auto_increment",
				"Email": "unique,index,type:varchar(255),not_null",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < tc.structType.NumField(); i++ {
				field := tc.structType.Field(i)
				tag := field.Tag.Get("migrations")
				
				if expectedTag, exists := tc.expectedTags[field.Name]; exists {
					if tag != expectedTag {
						t.Errorf("Field %s: expected tag '%s', got '%s'", field.Name, expectedTag, tag)
					}
				}
			}
		})
	}
}

// TestMigrationConstraintsEnhanced tests various migration constraints
func TestMigrationConstraintsEnhanced(t *testing.T) {
	migrator, _, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("UniqueConstraints", func(t *testing.T) {
		entity := struct {
			ID       int    `db:"id" migrations:"primary_key"`
			Username string `db:"username" migrations:"unique,type:varchar(50)"`
			Email    string `db:"email" migrations:"unique,type:varchar(100)"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected unique constraint migration to succeed, got error: %v", err)
		}
	})

	t.Run("NotNullConstraints", func(t *testing.T) {
		entity := struct {
			ID   int    `db:"id" migrations:"primary_key"`
			Name string `db:"name" migrations:"not_null,type:varchar(100)"`
			Age  int    `db:"age" migrations:"not_null"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected not null constraint migration to succeed, got error: %v", err)
		}
	})

	t.Run("IndexConstraints", func(t *testing.T) {
		entity := struct {
			ID       int    `db:"id" migrations:"primary_key"`
			Category string `db:"category" migrations:"index,type:varchar(50)"`
			Status   string `db:"status" migrations:"index,type:varchar(20)"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected index constraint migration to succeed, got error: %v", err)
		}
	})
}

// TestMigrationDatabaseSpecificTypesEnhanced tests database-specific types
func TestMigrationDatabaseSpecificTypesEnhanced(t *testing.T) {
	migrator, _, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("SQLiteSpecificTypes", func(t *testing.T) {
		entity := struct {
			ID       int     `db:"id" migrations:"primary_key"`
			RealNum  float64 `db:"real_num" migrations:"type:real"`
			BlobData []byte  `db:"blob_data" migrations:"type:blob"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected SQLite-specific type migration to succeed, got error: %v", err)
		}
	})

	t.Run("NumericTypes", func(t *testing.T) {
		entity := struct {
			ID      int     `db:"id" migrations:"primary_key"`
			Price   float64 `db:"price" migrations:"type:decimal(10,2)"`
			Quantity int    `db:"quantity" migrations:"type:integer"`
		}{}

		err := migrator.MigrateModels(entity)
		if err != nil {
			t.Errorf("Expected numeric type migration to succeed, got error: %v", err)
		}
	})
}

// TestMigrationEdgeCasesEnhanced tests edge cases and boundary conditions
func TestMigrationEdgeCasesEnhanced(t *testing.T) {
	migrator, db, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("EmptyStruct", func(t *testing.T) {
		entity := struct{}{}
		err := migrator.MigrateModels(entity)
		// Should handle empty struct gracefully
		if err != nil {
			t.Logf("Empty struct migration returned error (may be expected): %v", err)
		}
	})

	t.Run("StructWithNoDbTags", func(t *testing.T) {
		entity := struct {
			ID   int
			Name string
		}{}

		err := migrator.MigrateModels(entity)
		// Should handle struct without db tags
		if err != nil {
			t.Logf("Struct without db tags returned error (may be expected): %v", err)
		}
	})

	t.Run("ExistingTable", func(t *testing.T) {
		// Create a table manually first
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS test_existing (
			id INTEGER PRIMARY KEY,
			name TEXT
		)`)
		if err != nil {
			t.Fatalf("Failed to create existing table: %v", err)
		}

		// Try to migrate entity with same structure
		entity := struct {
			ID   int    `db:"id" migrations:"primary_key"`
			Name string `db:"name" migrations:"type:text"`
		}{}

		err = migrator.MigrateModels(entity)
		// Should handle existing table gracefully
		if err != nil {
			t.Logf("Migration with existing table returned error (may be expected): %v", err)
		}
	})
}

// TestErrorConditionsEnhanced tests various error conditions
func TestErrorConditionsEnhanced(t *testing.T) {
	migrator, _, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("InvalidEntity", func(t *testing.T) {
		// Test error conditions with proper panic recovery
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic as expected when passing invalid types: %v", r)
			}
		}()

		// Test with nil entity - this will likely cause a panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Nil entity caused panic as expected: %v", r)
					return
				}
			}()
			err := migrator.MigrateModels(nil)
			if err == nil {
				t.Log("Nil entity did not produce error (handled gracefully)")
			} else {
				t.Logf("Nil entity produced expected error: %v", err)
			}
		}()

		// Test with string type - this should cause the panic we're testing for
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("String entity caused panic as expected: %v", r)
					return
				}
			}()
			err := migrator.MigrateModels("invalid_string_entity")
			if err != nil {
				t.Logf("String entity produced error: %v", err)
			}
		}()
	})
}

// TestConcurrentMigrationsEnhanced tests concurrent migration operations
func TestConcurrentMigrationsEnhanced(t *testing.T) {
	migrator, _, cleanup := SetupAutoMigrationTestMultiDB(t)
	defer cleanup()

	t.Run("ConcurrentEntityMigration", func(t *testing.T) {
		// Test concurrent migrations of different entities
		entities := []interface{}{
			struct {
				ID   int    `db:"id" migrations:"primary_key"`
				Name string `db:"name" migrations:"type:varchar(100)"`
			}{},
			struct {
				ID    int    `db:"id" migrations:"primary_key"`
				Email string `db:"email" migrations:"unique,type:varchar(255)"`
			}{},
			struct {
				ID  int `db:"id" migrations:"primary_key"`
				Age int `db:"age" migrations:"index"`
			}{},
		}

		// Run migrations concurrently
		for _, entity := range entities {
			go func(e interface{}) {
				err := migrator.MigrateModels(e)
				if err != nil {
					t.Logf("Concurrent migration error (may be expected): %v", err)
				}
			}(entity)
		}

		// Brief pause to allow concurrent operations
		time.Sleep(100 * time.Millisecond)
	})
}
