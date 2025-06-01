package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// MigrationState represents the state of a migration
type MigrationState int

const (
	MigrationStatePending MigrationState = iota
	MigrationStateApplied
	MigrationStateFailed
)

func (s MigrationState) String() string {
	switch s {
	case MigrationStatePending:
		return "Pending"
	case MigrationStateApplied:
		return "Applied"
	case MigrationStateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Migration represents a database migration with EF Core-like structure
type Migration struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Version     int64          `json:"version"`
	Description string         `json:"description"`
	UpSQL       string         `json:"up_sql"`
	DownSQL     string         `json:"down_sql"`
	AppliedAt   time.Time      `json:"applied_at,omitempty"`
	State       MigrationState `json:"state"`
}

// MigrationHistory represents the complete migration history
type MigrationHistory struct {
	Applied []Migration `json:"applied"`
	Pending []Migration `json:"pending"`
	Failed  []Migration `json:"failed"`
}

// EFMigrationManager provides Entity Framework Core-like migration lifecycle
type EFMigrationManager struct {
	db                *sql.DB
	logger            *log.Logger
	migrationTable    string
	historyTable      string
	snapshotTable     string
	autoMigrate       bool
	pendingMigrations []Migration
	loadedMigrations  map[string]Migration // Store all loaded migrations with their SQL
}

// EFMigrationConfig configures the migration manager
type EFMigrationConfig struct {
	AutoMigrate    bool
	MigrationTable string
	HistoryTable   string
	SnapshotTable  string
	Logger         *log.Logger
}

// DefaultEFMigrationConfig returns default configuration
func DefaultEFMigrationConfig() *EFMigrationConfig {
	return &EFMigrationConfig{
		AutoMigrate:    false,
		MigrationTable: "__ef_migrations_history",
		HistoryTable:   "__migration_history",
		SnapshotTable:  "__model_snapshot",
		Logger:         log.Default(),
	}
}

// NewEFMigrationManager creates a new EF Core-like migration manager
func NewEFMigrationManager(db *sql.DB, config *EFMigrationConfig) *EFMigrationManager {
	if config == nil {
		config = DefaultEFMigrationConfig()
	}

	return &EFMigrationManager{
		db:                db,
		logger:            config.Logger,
		migrationTable:    config.MigrationTable,
		historyTable:      config.HistoryTable,
		snapshotTable:     config.SnapshotTable,
		autoMigrate:       config.AutoMigrate,
		pendingMigrations: make([]Migration, 0),
		loadedMigrations:  make(map[string]Migration),
	}
}

// detectDatabaseType detects the database type from the driver
func (em *EFMigrationManager) detectDatabaseType() string {
	// Query to detect database type
	var dbType string
	row := em.db.QueryRow("SELECT 1")
	if err := row.Scan(&dbType); err == nil {
		// Try PostgreSQL specific query
		row = em.db.QueryRow("SELECT version()")
		var version string
		if err := row.Scan(&version); err == nil {
			if strings.Contains(strings.ToLower(version), "postgresql") {
				return "postgres"
			}
		}

		// Try SQLite specific query
		row = em.db.QueryRow("SELECT sqlite_version()")
		if err := row.Scan(&version); err == nil {
			return "sqlite"
		}
	}

	// Default to postgres
	return "postgres"
}

// getAutoIncrementSQL returns the appropriate auto-increment SQL for the database type
func (em *EFMigrationManager) getAutoIncrementSQL() string {
	dbType := em.detectDatabaseType()
	switch dbType {
	case "sqlite":
		return "INTEGER PRIMARY KEY AUTOINCREMENT"
	default: // postgres
		return "SERIAL PRIMARY KEY"
	}
}

// EnsureSchema creates necessary migration tracking tables
func (em *EFMigrationManager) EnsureSchema() error {
	autoIncrement := em.getAutoIncrementSQL()

	queries := []string{
		// EF Migrations History table (equivalent to EF Core's __EFMigrationsHistory)
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				migration_id VARCHAR(150) PRIMARY KEY,
				product_version VARCHAR(32) NOT NULL,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, em.migrationTable),

		// Migration history with full details
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id %s,
				migration_id VARCHAR(150) NOT NULL,
				name VARCHAR(255) NOT NULL,
				version BIGINT NOT NULL,
				description TEXT,
				up_sql TEXT NOT NULL,
				down_sql TEXT,
				applied_at TIMESTAMP,
				rolled_back_at TIMESTAMP,
				state VARCHAR(20) DEFAULT 'pending',
				execution_time_ms INTEGER,
				error_message TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, em.historyTable, autoIncrement),

		// Model snapshot table (equivalent to EF Core's model snapshot)
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id %s,
				model_hash VARCHAR(64) NOT NULL,
				model_definition TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, em.snapshotTable, autoIncrement),

		// Indexes for performance
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_version ON %s(version)`,
			strings.ReplaceAll(em.historyTable, "__", ""), em.historyTable),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_state ON %s(state)`,
			strings.ReplaceAll(em.historyTable, "__", ""), em.historyTable),
	}

	for _, query := range queries {
		if _, err := em.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create migration schema: %w", err)
		}
	}

	em.logger.Println("✓ Migration schema initialized")
	return nil
}

// AddMigration adds a new migration (equivalent to Add-Migration in EF Core)
func (em *EFMigrationManager) AddMigration(name, description string, upSQL, downSQL string) *Migration {
	version := time.Now().Unix()
	migrationID := fmt.Sprintf("%d_%s", version, strings.ReplaceAll(name, " ", "_"))

	migration := Migration{
		ID:          migrationID,
		Name:        name,
		Version:     version,
		Description: description,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
		State:       MigrationStatePending,
	}

	em.pendingMigrations = append(em.pendingMigrations, migration)
	em.logger.Printf("✓ Added migration: %s", migrationID)

	return &migration
}

// AddLoadedMigration adds a migration loaded from filesystem
func (em *EFMigrationManager) AddLoadedMigration(migration Migration) {
	// Store the loaded migration with its SQL content
	em.loadedMigrations[migration.ID] = migration

	// Check if migration is already applied by querying the database
	var query string
	dbType := em.detectDatabaseType()
	if dbType == "postgres" {
		query = fmt.Sprintf(`
			SELECT COUNT(*) FROM %s WHERE migration_id = $1
		`, em.historyTable)
	} else {
		query = fmt.Sprintf(`
			SELECT COUNT(*) FROM %s WHERE migration_id = ?
		`, em.historyTable)
	}

	var count int
	err := em.db.QueryRow(query, migration.ID).Scan(&count)
	if err != nil {
		// If error querying, assume it's pending
		em.pendingMigrations = append(em.pendingMigrations, migration)
		return
	}

	// Only add to pending if not already applied
	if count == 0 {
		em.pendingMigrations = append(em.pendingMigrations, migration)
		em.logger.Printf("✓ Loaded migration from file: %s", migration.ID)
	}
}

// GetMigrationHistory retrieves complete migration history (like Get-Migration)
func (em *EFMigrationManager) GetMigrationHistory() (*MigrationHistory, error) {
	history := &MigrationHistory{
		Applied: make([]Migration, 0),
		Pending: make([]Migration, 0),
		Failed:  make([]Migration, 0),
	}

	// Get all migrations from history table
	query := fmt.Sprintf(`
		SELECT migration_id, name, version, description, up_sql, down_sql, 
		       applied_at, state
		FROM %s
		ORDER BY version ASC
	`, em.historyTable)

	rows, err := em.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var migration Migration
		var appliedAt sql.NullTime
		var state string

		err := rows.Scan(
			&migration.ID,
			&migration.Name,
			&migration.Version,
			&migration.Description,
			&migration.UpSQL,
			&migration.DownSQL,
			&appliedAt,
			&state,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}

		if appliedAt.Valid {
			migration.AppliedAt = appliedAt.Time
		}

		switch state {
		case "applied":
			migration.State = MigrationStateApplied
			history.Applied = append(history.Applied, migration)
		case "failed":
			migration.State = MigrationStateFailed
			history.Failed = append(history.Failed, migration)
		default:
			migration.State = MigrationStatePending
			history.Pending = append(history.Pending, migration)
		}
	}

	// Add pending migrations from memory
	history.Pending = append(history.Pending, em.pendingMigrations...)

	return history, nil
}

// UpdateDatabase applies pending migrations (equivalent to Update-Database)
func (em *EFMigrationManager) UpdateDatabase(targetMigration ...string) error {
	if err := em.EnsureSchema(); err != nil {
		return err
	}

	// Get pending migrations
	history, err := em.GetMigrationHistory()
	if err != nil {
		return err
	}

	migrations := history.Pending
	if len(migrations) == 0 {
		em.logger.Println("✓ No pending migrations")
		return nil
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Apply up to target migration if specified
	if len(targetMigration) > 0 {
		target := targetMigration[0]
		for i, migration := range migrations {
			if migration.ID == target || migration.Name == target {
				migrations = migrations[:i+1]
				break
			}
		}
	}

	em.logger.Printf("Applying %d migration(s)...", len(migrations))

	for _, migration := range migrations {
		if err := em.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
		}
	}

	em.logger.Println("✓ All migrations applied successfully")
	return nil
}

// applyMigration applies a single migration
func (em *EFMigrationManager) applyMigration(migration Migration) error {
	startTime := time.Now()

	// Begin transaction
	tx, err := em.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	em.logger.Printf("Applying migration: %s", migration.ID)

	// Execute UP SQL
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		// Record failed migration
		em.recordMigrationResult(migration, MigrationStateFailed, 0, err.Error())
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// Record in EF migrations history table
	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s (migration_id, product_version) VALUES ($1, $2)", em.migrationTable),
		migration.ID, "GRA-1.1.0",
	)
	if err != nil {
		return fmt.Errorf("failed to record in EF history: %w", err)
	}

	// Record in detailed history table
	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO %s (migration_id, name, version, description, up_sql, down_sql, applied_at, state, execution_time_ms)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, em.historyTable),
		migration.ID, migration.Name, migration.Version, migration.Description,
		migration.UpSQL, migration.DownSQL, time.Now(), "applied", executionTime,
	)
	if err != nil {
		return fmt.Errorf("failed to record in history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	em.logger.Printf("✓ Applied migration: %s (%dms)", migration.ID, executionTime)
	return nil
}

// RollbackMigration rolls back to a specific migration (equivalent to Update-Database with target)
func (em *EFMigrationManager) RollbackMigration(targetMigration string) error {
	// Get applied migrations
	history, err := em.GetMigrationHistory()
	if err != nil {
		return err
	}

	// Find target migration
	targetIndex := -1
	for i, migration := range history.Applied {
		if migration.ID == targetMigration || migration.Name == targetMigration {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("migration not found: %s", targetMigration)
	}

	// Get migrations to rollback (all after target)
	toRollback := history.Applied[targetIndex+1:]

	// Sort in reverse order for rollback
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].Version > toRollback[j].Version
	})

	em.logger.Printf("Rolling back %d migration(s)...", len(toRollback))

	for _, migration := range toRollback {
		// Get the loaded migration with DOWN SQL from filesystem
		if loadedMigration, exists := em.loadedMigrations[migration.ID]; exists {
			// Use the loaded migration which has the DOWN SQL
			if err := em.rollbackMigration(loadedMigration); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
			}
		} else {
			// Fallback to the migration from history (might not have DOWN SQL)
			if err := em.rollbackMigration(migration); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
			}
		}
	}

	em.logger.Println("✓ Rollback completed successfully")
	return nil
}

// rollbackMigration rolls back a single migration
func (em *EFMigrationManager) rollbackMigration(migration Migration) error {
	if migration.DownSQL == "" {
		return fmt.Errorf("no down migration available for: %s", migration.ID)
	}

	startTime := time.Now()

	// Begin transaction
	tx, err := em.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	em.logger.Printf("Rolling back migration: %s", migration.ID)

	// Execute DOWN SQL
	if _, err := tx.Exec(migration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove from EF migrations history
	_, err = tx.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE migration_id = $1", em.migrationTable),
		migration.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove from EF history: %w", err)
	}

	// Update history table
	executionTime := int(time.Since(startTime).Milliseconds())
	_, err = tx.Exec(
		fmt.Sprintf(`
			UPDATE %s 
			SET rolled_back_at = $1, state = 'rolled_back'
			WHERE migration_id = $2
		`, em.historyTable),
		time.Now(), migration.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	em.logger.Printf("✓ Rolled back migration: %s (%dms)", migration.ID, executionTime)
	return nil
}

// GetAppliedMigrations returns list of applied migrations
func (em *EFMigrationManager) GetAppliedMigrations() ([]string, error) {
	query := fmt.Sprintf("SELECT migration_id FROM %s ORDER BY applied_at", em.migrationTable)

	rows, err := em.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var migrationID string
		if err := rows.Scan(&migrationID); err != nil {
			return nil, err
		}
		migrations = append(migrations, migrationID)
	}

	return migrations, nil
}

// GetPendingMigrations returns list of pending migrations
func (em *EFMigrationManager) GetPendingMigrations() ([]Migration, error) {
	history, err := em.GetMigrationHistory()
	if err != nil {
		return nil, err
	}
	return history.Pending, nil
}

// HasPendingMigrations checks if there are pending migrations
func (em *EFMigrationManager) HasPendingMigrations() (bool, error) {
	pending, err := em.GetPendingMigrations()
	if err != nil {
		return false, err
	}
	return len(pending) > 0, nil
}

// recordMigrationResult records the result of a migration attempt
func (em *EFMigrationManager) recordMigrationResult(migration Migration, state MigrationState, executionTime int, errorMessage string) {
	stateStr := "pending"
	switch state {
	case MigrationStateApplied:
		stateStr = "applied"
	case MigrationStateFailed:
		stateStr = "failed"
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (migration_id, name, version, description, up_sql, down_sql, state, execution_time_ms, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (migration_id) DO UPDATE SET 
			state = EXCLUDED.state,
			execution_time_ms = EXCLUDED.execution_time_ms,
			error_message = EXCLUDED.error_message
	`, em.historyTable)

	_, err := em.db.Exec(query,
		migration.ID, migration.Name, migration.Version, migration.Description,
		migration.UpSQL, migration.DownSQL, stateStr, executionTime, errorMessage,
	)

	if err != nil {
		em.logger.Printf("Warning: Failed to record migration result: %v", err)
	}
}

// CreateAutoMigrations creates migrations automatically based on model changes
func (em *EFMigrationManager) CreateAutoMigrations(entities []interface{}, migrationName string) error {
	// This would compare current model with snapshot and generate migrations
	// For now, we'll create a basic implementation

	upSQL := em.generateCreateTablesSQL(entities)
	downSQL := em.generateDropTablesSQL(entities)

	migration := em.AddMigration(
		migrationName,
		fmt.Sprintf("Auto-generated migration for %d entities", len(entities)),
		upSQL,
		downSQL,
	)

	em.logger.Printf("✓ Created auto-migration: %s", migration.ID)
	return nil
}

// generateCreateTablesSQL generates SQL to create tables for entities
func (em *EFMigrationManager) generateCreateTablesSQL(entities []interface{}) string {
	var sql strings.Builder

	for _, entity := range entities {
		tableName := em.getTableName(entity)
		sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
		sql.WriteString("    id SERIAL PRIMARY KEY,\n")
		sql.WriteString("    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\n")
		sql.WriteString("    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n")
		sql.WriteString(");\n\n")
	}

	return sql.String()
}

// generateDropTablesSQL generates SQL to drop tables for entities
func (em *EFMigrationManager) generateDropTablesSQL(entities []interface{}) string {
	var sql strings.Builder

	for _, entity := range entities {
		tableName := em.getTableName(entity)
		sql.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", tableName))
	}

	return sql.String()
}

// getTableName gets table name from entity
func (em *EFMigrationManager) getTableName(entity interface{}) string {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// Convert CamelCase to snake_case
	name := entityType.Name()
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r + 32) // Convert to lowercase
	}

	return result.String()
}
