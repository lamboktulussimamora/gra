package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HybridMigrator provides EF Core-style migration functionality
type HybridMigrator struct {
	db               *sql.DB
	driver           DatabaseDriver
	registry         *ModelRegistry
	inspector        *DatabaseInspector
	changeDetector   *ChangeDetector
	sqlGenerator     *SQLGenerator
	migrationsDir    string
	migrationHistory *HybridMigrationHistory
	efManager        *EFMigrationManager // EF migration system for proper SQL execution
}

// HybridMigrationHistory tracks applied migrations for the hybrid system
type HybridMigrationHistory struct {
	db     *sql.DB
	driver DatabaseDriver
}

// MigrationRecord represents a migration in the history table
type MigrationRecord struct {
	ID            int64
	Name          string
	Checksum      string
	AppliedAt     time.Time
	IsDestructive bool
}

// NewHybridMigrator creates a new hybrid migrator
func NewHybridMigrator(db *sql.DB, driver DatabaseDriver, migrationsDir string) *HybridMigrator {
	registry := NewModelRegistry(driver)
	inspector := NewDatabaseInspector(db, driver)
	changeDetector := NewChangeDetector(registry, inspector)
	sqlGenerator := NewSQLGenerator(driver)
	migrationHistory := &HybridMigrationHistory{db: db, driver: driver}

	// Create EF migration manager for proper SQL execution with placeholder conversion
	efConfig := DefaultEFMigrationConfig()
	efManager := NewEFMigrationManager(db, efConfig)

	return &HybridMigrator{
		db:               db,
		driver:           driver,
		registry:         registry,
		inspector:        inspector,
		changeDetector:   changeDetector,
		sqlGenerator:     sqlGenerator,
		migrationsDir:    migrationsDir,
		migrationHistory: migrationHistory,
		efManager:        efManager,
	}
}

// DbSet registers a model with the migrator (EF Core-style)
func (hm *HybridMigrator) DbSet(model interface{}, tableName ...string) {
	// Note: RegisterModel now extracts table name from struct tags
	// The tableName parameter is ignored for now - could be enhanced later
	hm.registry.RegisterModel(model)
}

// AddMigration detects changes and creates a new migration file
func (hm *HybridMigrator) AddMigration(name string, mode MigrationMode) (*MigrationFile, error) {
	// Ensure migrations directory exists
	// #nosec G301 -- Directory must be user-accessible for migration files
	if err := os.MkdirAll(hm.migrationsDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Initialize migration history table if needed
	if err := hm.migrationHistory.ensureHistoryTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize migration history: %w", err)
	}

	// Detect changes
	plan, err := hm.changeDetector.DetectChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to detect changes: %w", err)
	}

	// Validate the plan
	if err := hm.changeDetector.ValidateMigrationPlan(plan); err != nil {
		return nil, fmt.Errorf("migration plan validation failed: %w", err)
	}

	// Check if there are any changes
	if len(plan.Changes) == 0 {
		return nil, fmt.Errorf("no changes detected")
	}

	// Check migration mode compatibility
	if err := hm.validateMigrationMode(plan, mode); err != nil {
		return nil, fmt.Errorf("migration mode validation failed: %w", err)
	}

	// Generate SQL
	migrationSQL, err := hm.sqlGenerator.GenerateMigrationSQL(plan)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Create migration file
	migrationFile := &MigrationFile{
		Name:      name,
		Timestamp: time.Now(),
		UpSQL:     []string{migrationSQL.UpScript},
		DownSQL:   []string{migrationSQL.DownScript},
		Checksum:  plan.PlanChecksum,
		Changes:   plan.Changes,
		Mode:      mode,
	}

	// Save migration file to disk
	filename := hm.generateMigrationFilename(name, migrationFile.Timestamp)
	migrationFile.FilePath = filepath.Join(hm.migrationsDir, filename)

	if err := hm.saveMigrationFile(migrationFile); err != nil {
		return nil, fmt.Errorf("failed to save migration file: %w", err)
	}

	return migrationFile, nil
}

// ApplyMigrations applies pending migrations
func (hm *HybridMigrator) ApplyMigrations(mode MigrationMode) error {
	// Initialize EF migration schema first
	if err := hm.efManager.EnsureSchema(); err != nil {
		return fmt.Errorf("failed to initialize EF migration schema: %w", err)
	}

	// Initialize migration history table
	if err := hm.migrationHistory.ensureHistoryTable(); err != nil {
		return fmt.Errorf("failed to initialize migration history: %w", err)
	}

	// Get pending migrations first
	pendingMigrations, err := hm.getPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	// Check for detected changes that don't have migration files yet
	plan, err := hm.changeDetector.DetectChanges()
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	// If there are no pending migrations but there are detected changes,
	// it means there are schema changes that need migration files created first
	if len(pendingMigrations) == 0 && len(plan.Changes) > 0 {
		return fmt.Errorf("detected %d schema changes that require migration files. Use CreateMigration() to create migration files first", len(plan.Changes))
	}

	// If there are pending migrations, validate them against migration mode
	if len(pendingMigrations) > 0 {
		// Create a plan from the pending migrations to validate mode compatibility
		migrationPlan := &MigrationPlan{
			Changes:        []MigrationChange{}, // We validate individual migrations later
			HasDestructive: false,               // Will be set per migration
			RequiresReview: false,               // Will be set per migration
		}

		// Check if any pending migration is destructive
		for _, migration := range pendingMigrations {
			if migration.HasDestructiveChanges() {
				migrationPlan.HasDestructive = true
				break
			}
		}

		// Validate migration mode for pending migrations
		if err := hm.validateMigrationMode(migrationPlan, mode); err != nil {
			return fmt.Errorf("pending migrations validation failed: %w", err)
		}
	}

	if len(pendingMigrations) == 0 {
		fmt.Println("No pending migrations")
		return nil
	}

	// Apply each migration
	for _, migration := range pendingMigrations {
		fmt.Printf("Applying migration: %s\n", migration.Name)

		// Validate migration mode
		if err := hm.validateMigrationMode(&MigrationPlan{
			HasDestructive: migration.HasDestructiveChanges(),
			RequiresReview: migration.RequiresReview(),
		}, mode); err != nil {
			return fmt.Errorf("migration %s failed mode validation: %w", migration.Name, err)
		}

		// Apply migration
		if err := hm.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		// Record in history
		if err := hm.migrationHistory.addRecord(migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
		}

		fmt.Printf("Applied migration: %s\n", migration.Name)
	}

	return nil
}

// RevertMigration reverts the last applied migration
func (hm *HybridMigrator) RevertMigration() error {
	// Get last applied migration
	lastMigration, err := hm.migrationHistory.getLastApplied()
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to revert")
	}

	// Load migration file
	migrationFile, err := hm.loadMigrationFile(lastMigration.Name)
	if err != nil {
		return fmt.Errorf("failed to load migration file: %w", err)
	}

	fmt.Printf("Reverting migration: %s\n", migrationFile.Name)

	// Execute down scripts directly with proper placeholder conversion
	for _, script := range migrationFile.DownSQL {
		// Convert placeholders for the database driver
		convertedScript := hm.efManager.ConvertQueryPlaceholders(script)
		if _, err := hm.db.Exec(convertedScript); err != nil {
			return fmt.Errorf("failed to execute down script: %w", err)
		}
	}

	// Remove from hybrid migration history
	if err := hm.migrationHistory.removeRecord(lastMigration.ID); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("Reverted migration: %s\n", migrationFile.Name)
	return nil
}

// GetMigrationStatus returns the current migration status
func (hm *HybridMigrator) GetMigrationStatus() (*MigrationStatus, error) {
	// Initialize EF migration schema first
	if err := hm.efManager.EnsureSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize EF migration schema: %w", err)
	}

	// Initialize migration history table
	if err := hm.migrationHistory.ensureHistoryTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize migration history: %w", err)
	}

	// Get all migration files
	allMigrations, err := hm.getAllMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := hm.migrationHistory.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create applied migrations map
	appliedMap := make(map[string]*MigrationRecord)
	for _, applied := range appliedMigrations {
		appliedMap[applied.Name] = applied
	}

	// Categorize migrations
	var pending, applied []*MigrationFile
	for _, migration := range allMigrations {
		if _, isApplied := appliedMap[migration.Name]; isApplied {
			applied = append(applied, migration)
		} else {
			pending = append(pending, migration)
		}
	}

	// Detect current changes
	plan, err := hm.changeDetector.DetectChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current changes: %w", err)
	}

	// HasPendingChanges should be true only if there are changes that can't be addressed
	// by applying existing migration files. If there are pending migration files that can
	// address the changes, then there are no "pending changes" in the sense of needing
	// new migration files to be created.
	hasPendingChanges := len(plan.Changes) > 0 && len(pending) == 0

	status := &MigrationStatus{
		PendingMigrations:     pending,
		AppliedMigrations:     applied,
		CurrentChanges:        plan.Changes,
		HasPendingChanges:     hasPendingChanges,
		HasDestructiveChanges: plan.HasDestructive,
		Summary:               hm.changeDetector.GetChangeSummary(plan),
	}

	return status, nil
}

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	PendingMigrations     []*MigrationFile
	AppliedMigrations     []*MigrationFile
	CurrentChanges        []MigrationChange
	HasPendingChanges     bool
	HasDestructiveChanges bool
	Summary               string
}

// validateMigrationMode validates if the migration can be applied in the given mode
func (hm *HybridMigrator) validateMigrationMode(plan *MigrationPlan, mode MigrationMode) error {
	switch mode {
	case Automatic:
		if plan.HasDestructive {
			return fmt.Errorf("automatic mode cannot apply destructive changes")
		}
		if plan.RequiresReview {
			return fmt.Errorf("automatic mode cannot apply changes that require review")
		}
	case Interactive:
		// Interactive mode can handle any changes with user confirmation
		return nil
	case GenerateOnly:
		// Generate only mode just creates files, no validation needed
		return nil
	case ForceDestructive:
		// Force mode can apply any changes
		return nil
	default:
		return fmt.Errorf("unknown migration mode: %s", mode)
	}
	return nil
}

// generateMigrationFilename generates a filename for a migration
func (hm *HybridMigrator) generateMigrationFilename(name string, timestamp time.Time) string {
	// Format: YYYYMMDDHHMMSS_migration_name.sql
	timestampStr := timestamp.Format("20060102150405")
	safeName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	return fmt.Sprintf("%s_%s.sql", timestampStr, safeName)
}

// saveMigrationFile saves a migration file to disk
func (hm *HybridMigrator) saveMigrationFile(migration *MigrationFile) error {
	content := hm.formatMigrationFileContent(migration)
	// #nosec G306 -- Migration files are not sensitive, but 0600 is stricter
	return os.WriteFile(migration.FilePath, []byte(content), 0600)
}

// formatMigrationFileContent formats the migration file content
func (hm *HybridMigrator) formatMigrationFileContent(migration *MigrationFile) string {
	var content strings.Builder

	// Header with metadata
	content.WriteString(fmt.Sprintf("-- Migration: %s\n", migration.Name))
	content.WriteString(fmt.Sprintf("-- Created: %s\n", migration.Timestamp.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("-- Checksum: %s\n", migration.Checksum))
	content.WriteString(fmt.Sprintf("-- Mode: %s\n", migration.Mode.String()))
	content.WriteString(fmt.Sprintf("-- Has Destructive: %t\n", migration.HasDestructiveChanges()))
	content.WriteString(fmt.Sprintf("-- Requires Review: %t\n", migration.RequiresReview()))
	content.WriteString("\n")

	// Warnings and errors
	warnings := migration.Warnings()
	if len(warnings) > 0 {
		content.WriteString("-- WARNINGS:\n")
		for _, warning := range warnings {
			content.WriteString(fmt.Sprintf("-- * %s\n", warning))
		}
		content.WriteString("\n")
	}

	errors := migration.Errors()
	if len(errors) > 0 {
		content.WriteString("-- ERRORS:\n")
		for _, error := range errors {
			content.WriteString(fmt.Sprintf("-- * %s\n", error))
		}
		content.WriteString("\n")
	}

	// Up script
	content.WriteString("-- +migrate Up\n")
	for _, script := range migration.UpSQL {
		content.WriteString(script)
		content.WriteString("\n")
	}

	// Down script
	content.WriteString("-- +migrate Down\n")
	for _, script := range migration.DownSQL {
		content.WriteString(script)
		content.WriteString("\n")
	}

	return content.String()
}

// getPendingMigrations returns migrations that haven't been applied
func (hm *HybridMigrator) getPendingMigrations() ([]*MigrationFile, error) {
	allMigrations, err := hm.getAllMigrationFiles()
	if err != nil {
		return nil, err
	}

	appliedMigrations, err := hm.migrationHistory.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create map of applied migrations
	appliedMap := make(map[string]bool)
	for _, applied := range appliedMigrations {
		appliedMap[applied.Name] = true
	}

	// Filter pending migrations
	var pending []*MigrationFile
	for _, migration := range allMigrations {
		if !appliedMap[migration.Name] {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// getAllMigrationFiles loads all migration files from the migrations directory
func (hm *HybridMigrator) getAllMigrationFiles() ([]*MigrationFile, error) {
	var migrations []*MigrationFile

	err := filepath.WalkDir(hm.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		migration, err := hm.parseMigrationFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse migration file %s: %w", path, err)
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by timestamp
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Timestamp.Before(migrations[j].Timestamp)
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file from disk
func (hm *HybridMigrator) parseMigrationFile(filePath string) (*MigrationFile, error) {
	// #nosec G304 -- File path is determined by migration manager logic, not user input
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	migration := &MigrationFile{
		FilePath: filePath,
	}

	var upScript, downScript strings.Builder
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse metadata from comments
		switch {
		case strings.HasPrefix(line, "-- Migration:"):
			migration.Name = strings.TrimSpace(strings.TrimPrefix(line, "-- Migration:"))
		case strings.HasPrefix(line, "-- Created:"):
			timestampStr := strings.TrimSpace(strings.TrimPrefix(line, "-- Created:"))
			if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				migration.Timestamp = timestamp
			}
		case strings.HasPrefix(line, "-- Checksum:"):
			migration.Checksum = strings.TrimSpace(strings.TrimPrefix(line, "-- Checksum:"))
		case strings.HasPrefix(line, "-- Mode:"):
			modeStr := strings.TrimSpace(strings.TrimPrefix(line, "-- Mode:"))
			migration.Mode = ParseMigrationMode(modeStr)
		case strings.HasPrefix(line, "-- Has Destructive:"):
			// Parse the destructive flag from file metadata
			destructiveStr := strings.TrimSpace(strings.TrimPrefix(line, "-- Has Destructive:"))
			hasDestructive := destructiveStr == "true"
			migration.ParsedHasDestructive = &hasDestructive
		case strings.HasPrefix(line, "-- Requires Review:"):
			// This is calculated dynamically from Changes and Mode, skip parsing
		}

		// Parse sections
		if line == "-- +migrate Up" {
			currentSection = "up"
			continue
		} else if line == "-- +migrate Down" {
			currentSection = "down"
			continue
		}

		// Add content to appropriate section
		if currentSection == "up" {
			upScript.WriteString(line + "\n")
		} else if currentSection == "down" {
			downScript.WriteString(line + "\n")
		}
	}

	// Convert the concatenated scripts back to slices
	if upScript.Len() > 0 {
		migration.UpSQL = []string{strings.TrimSpace(upScript.String())}
	}
	if downScript.Len() > 0 {
		migration.DownSQL = []string{strings.TrimSpace(downScript.String())}
	}

	return migration, nil
}

// loadMigrationFile loads a specific migration file by name
func (hm *HybridMigrator) loadMigrationFile(name string) (*MigrationFile, error) {
	allMigrations, err := hm.getAllMigrationFiles()
	if err != nil {
		return nil, err
	}

	for _, migration := range allMigrations {
		if migration.Name == name {
			return migration, nil
		}
	}

	return nil, fmt.Errorf("migration not found: %s", name)
}

// generateMigrationID generates a unique migration ID from name and timestamp
func (hm *HybridMigrator) generateMigrationID(name string, timestamp time.Time) string {
	version := timestamp.Unix()
	return fmt.Sprintf("%d_%s", version, strings.ReplaceAll(name, " ", "_"))
}

// applyMigration applies a single migration using the EF migration system
func (hm *HybridMigrator) applyMigration(migration *MigrationFile) error {
	// Ensure EF migration schema is initialized
	if err := hm.efManager.EnsureSchema(); err != nil {
		return fmt.Errorf("failed to ensure EF migration schema: %w", err)
	}

	// Convert MigrationFile to EF Migration format
	efMigration := Migration{
		ID:          hm.generateMigrationID(migration.Name, migration.Timestamp),
		Name:        migration.Name,
		Version:     migration.Timestamp.Unix(),
		Description: fmt.Sprintf("Hybrid migration: %s", migration.Name),
		UpSQL:       strings.Join(migration.UpSQL, ";\n"),
		DownSQL:     strings.Join(migration.DownSQL, ";\n"),
		State:       MigrationStatePending,
	}

	// Apply the migration using EF migration system (which handles placeholder conversion)
	if err := hm.efManager.applyMigration(efMigration); err != nil {
		return fmt.Errorf("failed to apply migration via EF system: %w", err)
	}

	return nil
}

// executeMigrationScript executes a migration script
func (hm *HybridMigrator) executeMigrationScript(script string) error {
	// Split script into individual statements
	statements := hm.splitSQL(script)

	// Execute each statement
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}

		if _, err := hm.db.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute statement '%s': %w", statement, err)
		}
	}

	return nil
}

// splitSQL splits a SQL script into individual statements
func (hm *HybridMigrator) splitSQL(script string) []string {
	// Simple SQL splitting - could be enhanced for more complex cases
	statements := strings.Split(script, ";")

	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// Migration History Management

// ensureHistoryTable creates the migration history table if it doesn't exist
func (mh *HybridMigrationHistory) ensureHistoryTable() error {
	var createTableSQL string

	switch mh.driver {
	case PostgreSQL:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS __migration_history (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				checksum VARCHAR(64) NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				is_destructive BOOLEAN NOT NULL DEFAULT FALSE
			);
		`
	case MySQL:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS __migration_history (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				checksum VARCHAR(64) NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				is_destructive BOOLEAN NOT NULL DEFAULT FALSE
			);
		`
	case SQLite:
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS __migration_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				checksum TEXT NOT NULL,
				applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				is_destructive INTEGER NOT NULL DEFAULT 0
			);
		`
	default:
		return fmt.Errorf("unsupported driver: %s", mh.driver)
	}

	_, err := mh.db.Exec(createTableSQL)
	return err
}

// addRecord adds a migration record to the history
func (mh *HybridMigrationHistory) addRecord(migration *MigrationFile) error {
	query := `
		INSERT INTO __migration_history (name, checksum, is_destructive)
		VALUES (?, ?, ?)
	`
	_, err := mh.db.Exec(query, migration.Name, migration.Checksum, migration.HasDestructive())
	return err
}

// removeRecord removes a migration record from the history
func (mh *HybridMigrationHistory) removeRecord(id int64) error {
	query := `DELETE FROM __migration_history WHERE id = ?`
	_, err := mh.db.Exec(query, id)
	return err
}

// getLastApplied returns the last applied migration
func (mh *HybridMigrationHistory) getLastApplied() (*MigrationRecord, error) {
	query := `
		SELECT id, name, checksum, applied_at, is_destructive
		FROM __migration_history
		ORDER BY applied_at DESC, id DESC
		LIMIT 1
	`

	var record MigrationRecord
	err := mh.db.QueryRow(query).Scan(
		&record.ID,
		&record.Name,
		&record.Checksum,
		&record.AppliedAt,
		&record.IsDestructive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// getAppliedMigrations returns all applied migrations
func (mh *HybridMigrationHistory) getAppliedMigrations() ([]*MigrationRecord, error) {
	query := `
		SELECT id, name, checksum, applied_at, is_destructive
		FROM __migration_history
		ORDER BY applied_at ASC, id ASC
	`

	rows, err := mh.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Warning: Failed to close rows: %v\n", closeErr)
		}
	}()

	var records []*MigrationRecord
	for rows.Next() {
		var record MigrationRecord
		err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.Checksum,
			&record.AppliedAt,
			&record.IsDestructive,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, rows.Err()
}
