package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SimpleMigrator provides a simplified hybrid migration system
type SimpleMigrator struct {
	db            *sql.DB
	driver        DatabaseDriver
	registry      *ModelRegistry
	migrationsDir string
}

// NewSimpleMigrator creates a new simplified migrator
func NewSimpleMigrator(db *sql.DB, driver DatabaseDriver, migrationsDir string) *SimpleMigrator {
	return &SimpleMigrator{
		db:            db,
		driver:        driver,
		registry:      NewModelRegistry(driver),
		migrationsDir: migrationsDir,
	}
}

// DbSet registers a model (EF Core-style)
func (sm *SimpleMigrator) DbSet(model interface{}) {
	sm.registry.RegisterModel(model)
}

// GetRegisteredModels returns all registered models
func (sm *SimpleMigrator) GetRegisteredModels() map[string]*ModelSnapshot {
	return sm.registry.GetModels()
}

// TableExists checks if a table exists in the database
func (sm *SimpleMigrator) TableExists(tableName string) (bool, error) {
	var query string
	switch sm.driver {
	case PostgreSQL:
		query = `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1
		)`
	case MySQL:
		query = `SELECT COUNT(*) > 0 FROM information_schema.tables 
				WHERE table_schema = DATABASE() AND table_name = ?`
	case SQLite:
		query = `SELECT COUNT(*) > 0 FROM sqlite_master 
				WHERE type='table' AND name = ?`
	default:
		return false, fmt.Errorf("unsupported database driver: %s", sm.driver)
	}

	var exists bool
	err := sm.db.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

// GenerateCreateTableSQL generates SQL for creating a table
func (sm *SimpleMigrator) GenerateCreateTableSQL(snapshot *ModelSnapshot) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", snapshot.TableName))

	var columns []string
	var primaryKeys []string

	for _, col := range snapshot.Columns {
		colDef := fmt.Sprintf("  %s %s", col.Name, col.SQLType)

		if !col.Nullable {
			colDef += " NOT NULL"
		}

		if col.Default != nil {
			colDef += fmt.Sprintf(" DEFAULT %s", *col.Default)
		}

		if col.IsPrimaryKey {
			primaryKeys = append(primaryKeys, col.Name)
		}

		columns = append(columns, colDef)
	}

	sql.WriteString(strings.Join(columns, ",\n"))

	if len(primaryKeys) > 0 {
		sql.WriteString(",\n")
		sql.WriteString(fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	sql.WriteString("\n);")

	return sql.String()
}

// CreateInitialMigration creates a migration for all registered models
func (sm *SimpleMigrator) CreateInitialMigration(name string) (*MigrationFile, error) {
	models := sm.registry.GetModels()

	if len(models) == 0 {
		return nil, fmt.Errorf("no models registered")
	}

	// Create migrations directory
	if err := os.MkdirAll(sm.migrationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s.sql", timestamp, name)
	filepath := filepath.Join(sm.migrationsDir, filename)

	var upSQL strings.Builder
	var downSQL strings.Builder
	var changes []MigrationChange

	// Generate CREATE statements for all models
	for tableName, snapshot := range models {
		// Check if table already exists
		exists, err := sm.TableExists(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to check if table exists: %w", err)
		}

		if !exists {
			createSQL := sm.GenerateCreateTableSQL(snapshot)
			upSQL.WriteString(createSQL)
			upSQL.WriteString("\n\n")

			dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
			downSQL.WriteString(dropSQL)
			downSQL.WriteString("\n")

			changes = append(changes, MigrationChange{
				Type:        CreateTable,
				TableName:   tableName,
				Description: fmt.Sprintf("Create table %s", tableName),
			})
		}
	}

	if len(changes) == 0 {
		return nil, nil // No changes needed
	}

	// Create migration file
	migrationFile := &MigrationFile{
		Name:        name,
		Description: fmt.Sprintf("Initial migration: %s", name),
		Filename:    filename,
		Timestamp:   time.Now(),
		Changes:     changes,
		UpSQL:       []string{upSQL.String()},
		DownSQL:     []string{downSQL.String()},
	}

	// Write SQL file
	err := os.WriteFile(filepath, []byte(upSQL.String()), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write migration file: %w", err)
	}

	return migrationFile, nil
}

// ApplyMigration applies a single migration file
func (sm *SimpleMigrator) ApplyMigration(migrationFile *MigrationFile) error {
	for _, sql := range migrationFile.UpSQL {
		if strings.TrimSpace(sql) == "" {
			continue
		}

		_, err := sm.db.Exec(sql)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migrationFile.Name, err)
		}
	}

	return nil
}

// SimpleMigrationStatus represents the current migration status
type SimpleMigrationStatus struct {
	AppliedMigrations []string
	PendingMigrations []string
	HasPendingChanges bool
	Summary           string
}

// GetMigrationStatus returns the current migration status
func (sm *SimpleMigrator) GetMigrationStatus() (*SimpleMigrationStatus, error) {
	models := sm.registry.GetModels()

	var pendingTables []string
	for tableName := range models {
		exists, err := sm.TableExists(tableName)
		if err != nil {
			return nil, err
		}
		if !exists {
			pendingTables = append(pendingTables, tableName)
		}
	}

	status := &SimpleMigrationStatus{
		AppliedMigrations: []string{}, // Simplified - not tracking history yet
		PendingMigrations: pendingTables,
		HasPendingChanges: len(pendingTables) > 0,
	}

	if len(pendingTables) > 0 {
		status.Summary = fmt.Sprintf("Need to create %d tables: %s",
			len(pendingTables), strings.Join(pendingTables, ", "))
	} else {
		status.Summary = "Database is up to date"
	}

	return status, nil
}
