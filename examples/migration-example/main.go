package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/lamboktulussimamora/gra/orm/models"
	"github.com/lamboktulussimamora/gra/orm/schema"
	_ "github.com/lib/pq"
)

// MigrationRunner handles automatic database migrations
type MigrationRunner struct {
	db     *sql.DB
	logger *log.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(connectionString string) (*MigrationRunner, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MigrationRunner{
		db:     db,
		logger: log.Default(),
	}, nil
}

// Close closes the database connection
func (mr *MigrationRunner) Close() error {
	return mr.db.Close()
}

// AutoMigrate automatically creates or updates database schema based on entity models
func (mr *MigrationRunner) AutoMigrate() error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all entity types to migrate in dependency order
	entities := []interface{}{
		&models.Role{},
		&models.Category{},
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
		&models.UserRole{},
	}

	// Migrate each entity
	for _, entity := range entities {
		if err := mr.migrateEntity(entity); err != nil {
			return fmt.Errorf("failed to migrate entity %T: %w", entity, err)
		}
	}

	mr.logger.Println("Auto migration completed successfully")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) createMigrationsTable() error {
	query := "CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL UNIQUE, executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	_, err := mr.db.Exec(query)
	return err
}

// migrateEntity migrates a single entity
func (mr *MigrationRunner) migrateEntity(entity interface{}) error {
	// Check if table exists
	tableName := getTableName(entity)
	exists, err := mr.tableExists(tableName)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if !exists {
		// Create table
		mr.logger.Printf("Creating table: %s", tableName)
		if err := mr.createTable(entity, tableName); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	} else {
		mr.logger.Printf("Table %s already exists, skipping", tableName)
	}

	return nil
}

// tableExists checks if a table exists in the database
func (mr *MigrationRunner) tableExists(tableName string) (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)"
	var exists bool
	err := mr.db.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

// createTable creates a new table from entity
func (mr *MigrationRunner) createTable(entity interface{}, tableName string) error {
	createSQL := schema.GenerateCreateTableSQL(entity, tableName)
	mr.logger.Printf("Executing SQL: %s", createSQL)
	_, err := mr.db.Exec(createSQL)
	return err
}

// getTableName gets the table name from an entity
func getTableName(entity interface{}) string {
	if tn, ok := entity.(interface{ TableName() string }); ok {
		return tn.TableName()
	}

	// Default naming convention
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	return strings.ToLower(name) + "s"
}

// ShowStatus shows the current migration status
func (mr *MigrationRunner) ShowStatus() error {
	query := "SELECT name, executed_at FROM migrations ORDER BY executed_at"

	rows, err := mr.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			mr.logger.Printf("Warning: Failed to close rows: %v", closeErr)
		}
	}()

	mr.logger.Println("Migration Status:")
	mr.logger.Println("================")

	for rows.Next() {
		var name string
		var executedAt string
		if err := rows.Scan(&name, &executedAt); err != nil {
			return fmt.Errorf("failed to scan migration row: %w", err)
		}
		mr.logger.Printf("✓ %s (executed: %s)", name, executedAt)
	}

	return rows.Err()
}

// Main function to demonstrate migration functionality
func main() {
	// Example usage of the migration runner
	connectionString := "host=localhost port=5432 user=postgres password=password dbname=ecommerce sslmode=disable"

	runner, err := NewMigrationRunner(connectionString)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer func() {
		if closeErr := runner.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close migration runner: %v", closeErr)
		}
	}()

	log.Println("Starting automatic migration...")
	if err := runner.AutoMigrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
