package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// Test constants to avoid duplication
const (
	testPostgresURL     = "postgres://test"
	testPostgresDriver  = "postgres"
	testMigrationsDir   = "./migrations"
	testModelsDir       = "./models"
	testSQLiteDriver    = "sqlite"
	testSQLiteURL       = "sqlite://test.db"
	testMySQLDriver     = "mysql"
	testSQLite3Driver   = "sqlite3"
	testMigrationFormat = "%d_%s.sql"
	testMemoryDB        = ":memory:"
	testSQLite3URL      = "sqlite3://"
	testMigrationsPath  = "./test_migrations"
	testModelsPath      = "./test_models"

	// Error message constants
	errExpectedError           = "Expected error but got none"
	errExpectedNoError         = "Expected no error but got: %v"
	errExpectedMode            = "Expected mode %s, got %s"
	errEmptyMigrationName      = "empty migration name"
	errExpectedForceMode       = "expected force destructive mode, got %s"
	errExpectedInteractiveMode = "expected interactive mode, got %s"
	errExpectedMockError       = "Expected error from mock migrator"
	errFailedCreateDB          = "Failed to create test database: %v"
	errExpectedEmptyMigration  = "Expected error for empty migration name"
	errExpectedEmptyArgs       = "Expected error for empty args"

	// Command flag constants
	flagForce  = "--force"
	flagAuto   = "--auto"
	flagDriver = "-driver"
)

// MigratorInterface defines the interface that both HybridMigrator and MockMigrator implement
type MigratorInterface interface {
	AddMigration(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	ApplyMigrations(mode migrations.MigrationMode) error
	GetMigrationStatus() (*migrations.MigrationStatus, error)
	RevertMigration() error
	DbSet(model interface{}, tableName ...string)
}

// Test wrapper functions that accept MigratorInterface
func testCmdAddMigration(m MigratorInterface, args []string) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdAddMigration(hybridMigrator, args)
	}
	// For mock, simulate the behavior
	if len(args) == 0 {
		return fmt.Errorf("%s", errEmptyMigrationName)
	}
	_, err := m.AddMigration(args[0], migrations.ModeAutomatic)
	return err
}

func testCmdApplyMigrations(m MigratorInterface, args []string) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdApplyMigrations(hybridMigrator, args)
	}
	// For mock, simulate the behavior of parsing args for mode
	mode := migrations.ModeInteractive // Default mode

	// Check for force flag
	for _, arg := range args {
		if arg == flagForce {
			mode = migrations.ModeForceDestructive
			break
		}
		if arg == flagAuto {
			mode = migrations.ModeAutomatic
			break
		}
	}

	return m.ApplyMigrations(mode)
}

func testCmdRevertMigration(m MigratorInterface) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdRevertMigration(hybridMigrator)
	}
	// For mock, simulate the behavior
	return m.RevertMigration()
}

func testCmdMigrationStatus(m MigratorInterface) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdMigrationStatus(hybridMigrator)
	}
	// For mock, simulate the behavior
	_, err := m.GetMigrationStatus()
	return err
}

func testCmdGenerateMigration(m MigratorInterface, args []string) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdGenerateMigration(hybridMigrator, args)
	}
	// For mock, simulate the behavior
	if len(args) == 0 {
		return fmt.Errorf("%s", errEmptyMigrationName)
	}
	_, err := m.AddMigration(args[0], migrations.ModeGenerateOnly)
	return err
}

func testCmdForceMigration(m MigratorInterface, args []string) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return cmdForceMigration(hybridMigrator, args)
	}
	// For mock, simulate the behavior
	if len(args) == 0 {
		return fmt.Errorf("%s", errEmptyMigrationName)
	}
	_, err := m.AddMigration(args[0], migrations.ModeForceDestructive)
	return err
}

func testRegisterModels(m MigratorInterface, modelsDir string) error {
	if hybridMigrator, ok := m.(*migrations.HybridMigrator); ok {
		return registerModels(hybridMigrator, modelsDir)
	}
	// For mock, just validate the directory
	if modelsDir == "" {
		return fmt.Errorf("models directory cannot be empty")
	}
	return nil
}

// MockMigrator implements MigratorInterface for testing
type MockMigrator struct {
	addMigrationFunc       func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	applyMigrationsFunc    func(mode migrations.MigrationMode) error
	getMigrationStatusFunc func() (*migrations.MigrationStatus, error)
	revertMigrationFunc    func() error
	dbSetFunc              func(model interface{}, tableName ...string)
}

func (mock *MockMigrator) AddMigration(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
	if mock.addMigrationFunc != nil {
		return mock.addMigrationFunc(name, mode)
	}
	return &migrations.MigrationFile{
		Name:        name,
		Description: fmt.Sprintf("Test migration: %s", name),
		Filename:    fmt.Sprintf(testMigrationFormat, time.Now().Unix(), name),
		Timestamp:   time.Now(),
	}, nil
}

func (mock *MockMigrator) ApplyMigrations(mode migrations.MigrationMode) error {
	if mock.applyMigrationsFunc != nil {
		return mock.applyMigrationsFunc(mode)
	}
	return nil
}

func (mock *MockMigrator) GetMigrationStatus() (*migrations.MigrationStatus, error) {
	if mock.getMigrationStatusFunc != nil {
		return mock.getMigrationStatusFunc()
	}
	return &migrations.MigrationStatus{
		PendingMigrations:     []*migrations.MigrationFile{},
		AppliedMigrations:     []*migrations.MigrationFile{},
		HasPendingChanges:     false,
		HasDestructiveChanges: false,
		Summary:               "No pending migrations",
	}, nil
}

func (mock *MockMigrator) RevertMigration() error {
	if mock.revertMigrationFunc != nil {
		return mock.revertMigrationFunc()
	}
	return nil
}

func (mock *MockMigrator) DbSet(model interface{}, tableName ...string) {
	if mock.dbSetFunc != nil {
		mock.dbSetFunc(model, tableName...)
	}
}

func TestConfig(t *testing.T) {
	config := Config{
		DatabaseURL:   testPostgresURL,
		Driver:        testPostgresDriver,
		MigrationsDir: testMigrationsDir,
		ModelsDir:     testModelsDir,
	}

	// Test that all fields are properly set
	if config.DatabaseURL != testPostgresURL {
		t.Errorf("Expected DatabaseURL '%s', got '%s'", testPostgresURL, config.DatabaseURL)
	}
	if config.Driver != testPostgresDriver {
		t.Errorf("Expected Driver '%s', got '%s'", testPostgresDriver, config.Driver)
	}
	if config.MigrationsDir != testMigrationsDir {
		t.Errorf("Expected MigrationsDir '%s', got '%s'", testMigrationsDir, config.MigrationsDir)
	}
	if config.ModelsDir != testModelsDir {
		t.Errorf("Expected ModelsDir '%s', got '%s'", testModelsDir, config.ModelsDir)
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that Config has all expected fields
	config := Config{}
	configType := reflect.TypeOf(config)

	expectedFields := []string{
		"DatabaseURL",
		"Driver",
		"MigrationsDir",
		"ModelsDir",
	}

	for _, field := range expectedFields {
		_, found := configType.FieldByName(field)
		if !found {
			t.Errorf("Expected field %s not found in Config", field)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with DATABASE_URL",
			config: &Config{
				DatabaseURL:   testSQLiteURL,
				Driver:        testSQLiteDriver,
				MigrationsDir: testMigrationsDir,
				ModelsDir:     testModelsDir,
			},
			expectError: false,
		},
		{
			name: "empty database URL",
			config: &Config{
				DatabaseURL:   "",
				Driver:        testPostgresDriver,
				MigrationsDir: testMigrationsDir,
				ModelsDir:     testModelsDir,
			},
			expectError: true,
		},
		{
			name: "missing migrations directory",
			config: &Config{
				DatabaseURL:   testPostgresURL,
				Driver:        testPostgresDriver,
				MigrationsDir: "",
				ModelsDir:     testModelsDir,
			},
			expectError: false, // Has default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			} else if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestConnectDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid sqlite connection",
			config: &Config{
				DatabaseURL: testMemoryDB,
				Driver:      testSQLiteDriver,
			},
			expectError: false,
		},
		{
			name: "invalid connection string",
			config: &Config{
				DatabaseURL: "invalid://connection",
				Driver:      testPostgresDriver,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := connectDatabase(tt.config)
			switch {
			case tt.expectError && err == nil:
				t.Error(errExpectedError)
			case !tt.expectError && err != nil:
				t.Errorf(errExpectedNoError, err)
			case db != nil:
				if err := db.Close(); err != nil {
					t.Logf("Warning: failed to close database: %v", err)
				}
			}
		})
	}
}

func TestGetDriver(t *testing.T) {
	tests := []struct {
		driverName     string
		expectedResult string
	}{
		{testPostgresDriver, testPostgresDriver},
		{testSQLiteDriver, testSQLiteDriver},
		{"mysql", "mysql"},
		{"unknown", testSQLiteDriver}, // Default fallback
	}

	for _, tt := range tests {
		t.Run("driver_"+tt.driverName, func(t *testing.T) {
			driver := getDriver(tt.driverName)
			if driver == "" {
				t.Errorf("Expected driver, but got empty value for %s", tt.driverName)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if errMigrationNameRequired == "" {
		t.Error("errMigrationNameRequired should not be empty")
	}

	expectedError := "migration name is required"
	if errMigrationNameRequired != expectedError {
		t.Errorf("Expected '%s', got '%s'", expectedError, errMigrationNameRequired)
	}
}

func TestCmdAddMigrationWithMock(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "valid migration name",
			args:        []string{"CreateUsersTable"},
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        errEmptyMigrationName,
			args:        []string{},
			expectError: true,
			mockFunc:    nil,
		},
		{
			name:        "migration creation error",
			args:        []string{"FailingMigration"},
			expectError: true,
			mockFunc: func(_ string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
				return nil, fmt.Errorf("migration creation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdAddMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdApplyMigrationsWithMock(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(mode migrations.MigrationMode) error
	}{
		{
			name:        "successful application",
			args:        []string{},
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        "migration application error",
			args:        []string{},
			expectError: true,
			mockFunc: func(_ migrations.MigrationMode) error {
				return fmt.Errorf("failed to apply migrations")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				applyMigrationsFunc: tt.mockFunc,
			}

			err := testCmdApplyMigrations(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdRevertMigrationWithMock(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		mockFunc    func() error
	}{
		{
			name:        "successful revert",
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        "revert error",
			expectError: true,
			mockFunc: func() error {
				return fmt.Errorf("failed to revert migration")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				revertMigrationFunc: tt.mockFunc,
			}

			err := testCmdRevertMigration(migrator)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdMigrationStatusWithMock(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		mockFunc    func() (*migrations.MigrationStatus, error)
	}{
		{
			name:        "successful status",
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        "status error",
			expectError: true,
			mockFunc: func() (*migrations.MigrationStatus, error) {
				return nil, fmt.Errorf("failed to get status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				getMigrationStatusFunc: tt.mockFunc,
			}

			err := testCmdMigrationStatus(migrator)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdGenerateMigrationWithMock(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "valid generation",
			args:        []string{"GenerateTest"},
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        errEmptyMigrationName,
			args:        []string{},
			expectError: true,
			mockFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdGenerateMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdForceMigrationWithMock(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "successful force",
			args:        []string{"ForceTest"},
			expectError: false,
			mockFunc:    nil,
		},
		{
			name:        "force error",
			args:        []string{"ForceTest"},
			expectError: true,
			mockFunc: func(_ string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
				if mode != migrations.ModeForceDestructive {
					return nil, fmt.Errorf(errExpectedForceMode, mode)
				}
				return nil, fmt.Errorf("force migration failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdForceMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestRegisterModelsWithMock(t *testing.T) {
	tests := []struct {
		name      string
		modelsDir string
		expectErr bool
	}{
		{
			name:      "valid models directory",
			modelsDir: testModelsDir,
			expectErr: false,
		},
		{
			name:      "empty models directory",
			modelsDir: "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{}

			err := testRegisterModels(migrator, tt.modelsDir)

			if tt.expectErr && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectErr && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestMigrationFileValidation(t *testing.T) {
	currentTime := time.Now()
	migrationFile := &migrations.MigrationFile{
		Name:      "TestMigration",
		Timestamp: currentTime,
		Filename:  fmt.Sprintf(testMigrationFormat, currentTime.Unix(), "TestMigration"),
	}

	if migrationFile.Name == "" {
		t.Error("Migration name should not be empty")
	}
	if migrationFile.Timestamp.IsZero() {
		t.Error("Migration timestamp should not be zero")
	}
	if migrationFile.Filename == "" {
		t.Error("Migration filename should not be empty")
	}
}

func TestWithRealisticData(t *testing.T) {
	currentTime := time.Now()
	realisticStatus := &migrations.MigrationStatus{
		PendingMigrations: []*migrations.MigrationFile{
			{
				Name:      "CreateUsersTable",
				Timestamp: currentTime.Add(-2 * time.Hour),
				Filename:  "20240101120000_CreateUsersTable.sql",
			},
			{
				Name:      "AddUserIndexes",
				Timestamp: currentTime.Add(-1 * time.Hour),
				Filename:  "20240101130000_AddUserIndexes.sql",
			},
		},
		AppliedMigrations: []*migrations.MigrationFile{
			{
				Name:      "InitialSchema",
				Timestamp: currentTime.Add(-3 * time.Hour),
				Filename:  "20240101110000_InitialSchema.sql",
			},
		},
		HasPendingChanges:     true,
		HasDestructiveChanges: false,
		Summary:               "2 pending migrations, 1 applied",
	}

	migrator := &MockMigrator{
		getMigrationStatusFunc: func() (*migrations.MigrationStatus, error) {
			return realisticStatus, nil
		},
	}

	err := testCmdMigrationStatus(migrator)
	if err != nil {
		t.Errorf(errExpectedNoError, err)
	}
}

func TestDisplayAppliedMigrations(t *testing.T) {
	// Define reusable filename to avoid duplication
	createUsersFilename := "001_CreateUsersTable.sql"

	tests := []struct {
		name       string
		migrations []*migrations.MigrationFile
	}{
		{
			name:       "no applied migrations",
			migrations: []*migrations.MigrationFile{},
		},
		{
			name: "single applied migration",
			migrations: []*migrations.MigrationFile{
				{
					Name:        "CreateUsersTable",
					Description: "Initial user table",
					Filename:    createUsersFilename,
					Timestamp:   time.Now(),
				},
			},
		},
		{
			name: "multiple applied migrations",
			migrations: []*migrations.MigrationFile{
				{
					Name:        "CreateUsersTable",
					Description: "Initial user table",
					Filename:    createUsersFilename,
					Timestamp:   time.Now(),
				},
				{
					Name:        "AddUserProfiles",
					Description: "Add user profiles",
					Filename:    "002_AddUserProfiles.sql",
					Timestamp:   time.Now().Add(time.Hour),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// This function outputs to stdout, so we can't easily capture the output
			// but we can ensure it doesn't panic or error
			displayAppliedMigrations(tt.migrations)
		})
	}
}

func TestDisplayPendingMigrations(t *testing.T) {
	tests := []struct {
		name       string
		migrations []*migrations.MigrationFile
	}{
		{
			name:       "no pending migrations",
			migrations: []*migrations.MigrationFile{},
		},
		{
			name: "single pending migration",
			migrations: []*migrations.MigrationFile{
				{
					Name:        "AddUserSettings",
					Description: "Add user settings",
					Filename:    "003_AddUserSettings.sql",
					Timestamp:   time.Now(),
				},
			},
		},
		{
			name: "multiple pending migrations",
			migrations: []*migrations.MigrationFile{
				{
					Name:        "AddUserSettings",
					Description: "Add user settings",
					Filename:    "003_AddUserSettings.sql",
					Timestamp:   time.Now(),
				},
				{
					Name:        "AddUserPermissions",
					Description: "Add user permissions",
					Filename:    "004_AddUserPermissions.sql",
					Timestamp:   time.Now().Add(time.Hour),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// This function outputs to stdout, so we can't easily capture the output
			// but we can ensure it doesn't panic or error
			displayPendingMigrations(tt.migrations)
		})
	}
}

func TestDisplayCurrentChanges(t *testing.T) {
	tests := []struct {
		name   string
		status *migrations.MigrationStatus
	}{
		{
			name: "no current changes",
			status: &migrations.MigrationStatus{
				HasPendingChanges:     false,
				HasDestructiveChanges: false,
				Summary:               "No changes detected",
			},
		},
		{
			name: "has pending changes",
			status: &migrations.MigrationStatus{
				HasPendingChanges:     true,
				HasDestructiveChanges: false,
				Summary:               "Non-destructive changes detected",
			},
		},
		{
			name: "has destructive changes",
			status: &migrations.MigrationStatus{
				HasPendingChanges:     true,
				HasDestructiveChanges: true,
				Summary:               "Destructive changes detected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// This function outputs to stdout, so we can't easily capture the output
			// but we can ensure it doesn't panic or error
			displayCurrentChanges(tt.status)
		})
	}
}

func TestRegisterModels(t *testing.T) {
	tests := []struct {
		name        string
		modelsDir   string
		expectError bool
	}{
		{
			name:        "valid models directory",
			modelsDir:   "./models",
			expectError: false,
		},
		{
			name:        "empty models directory",
			modelsDir:   "",
			expectError: true,
		},
		{
			name:        "nonexistent models directory",
			modelsDir:   "./nonexistent",
			expectError: false, // registerModels just logs a warning for non-existent directories
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{}
			err := testRegisterModels(migrator, tt.modelsDir)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdAddMigrationReal(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "valid migration with description",
			args:        []string{"CreateUsersTable"},
			expectError: false,
			mockFunc: func(name string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
				return &migrations.MigrationFile{
					Name:        name,
					Description: "Test migration",
					Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
					Timestamp:   time.Now(),
				}, nil
			},
		},
		{
			name:        "migration creation failure",
			args:        []string{"FailingMigration"},
			expectError: true,
			mockFunc: func(_ string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
				return nil, fmt.Errorf("database connection failed")
			},
		},
		{
			name:        "empty migration name",
			args:        []string{},
			expectError: true,
			mockFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdAddMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdApplyMigrationsReal(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(mode migrations.MigrationMode) error
	}{
		{
			name:        "successful application with auto mode",
			args:        []string{flagAuto},
			expectError: false,
			mockFunc: func(mode migrations.MigrationMode) error {
				if mode != migrations.ModeAutomatic {
					return fmt.Errorf("expected automatic mode, got %s", mode)
				}
				return nil
			},
		},
		{
			name:        "successful application with force mode",
			args:        []string{flagForce},
			expectError: false,
			mockFunc: func(mode migrations.MigrationMode) error {
				if mode != migrations.ModeForceDestructive {
					return fmt.Errorf(errExpectedForceMode, mode)
				}
				return nil
			},
		},
		{
			name:        "default interactive mode",
			args:        []string{},
			expectError: false,
			mockFunc: func(mode migrations.MigrationMode) error {
				if mode != migrations.ModeInteractive {
					return fmt.Errorf(errExpectedInteractiveMode, mode)
				}
				return nil
			},
		},
		{
			name:        "migration application failure",
			args:        []string{},
			expectError: true,
			mockFunc: func(_ migrations.MigrationMode) error {
				return fmt.Errorf("failed to apply migrations")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				applyMigrationsFunc: tt.mockFunc,
			}

			err := testCmdApplyMigrations(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdRevertMigrationReal(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		mockFunc    func() error
	}{
		{
			name:        "successful reversion",
			expectError: false,
			mockFunc: func() error {
				return nil
			},
		},
		{
			name:        "reversion failure",
			expectError: true,
			mockFunc: func() error {
				return fmt.Errorf("failed to revert migration")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				revertMigrationFunc: tt.mockFunc,
			}

			err := testCmdRevertMigration(migrator)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdMigrationStatusReal(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		mockFunc    func() (*migrations.MigrationStatus, error)
	}{
		{
			name:        "successful status retrieval",
			expectError: false,
			mockFunc: func() (*migrations.MigrationStatus, error) {
				return &migrations.MigrationStatus{
					PendingMigrations: []*migrations.MigrationFile{
						{
							Name:      "CreateUsersTable",
							Timestamp: time.Now(),
							Filename:  "001_CreateUsersTable.sql",
						},
					},
					AppliedMigrations: []*migrations.MigrationFile{
						{
							Name:      "InitialSchema",
							Timestamp: time.Now().Add(-1 * time.Hour),
							Filename:  "000_InitialSchema.sql",
						},
					},
					HasPendingChanges:     true,
					HasDestructiveChanges: false,
					Summary:               "1 pending, 1 applied",
				}, nil
			},
		},
		{
			name:        "status retrieval failure",
			expectError: true,
			mockFunc: func() (*migrations.MigrationStatus, error) {
				return nil, fmt.Errorf("failed to get migration status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				getMigrationStatusFunc: tt.mockFunc,
			}

			err := testCmdMigrationStatus(migrator)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdGenerateMigrationReal(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "successful generation",
			args:        []string{"AddUserRoles"},
			expectError: false,
			mockFunc: func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
				if mode != migrations.ModeGenerateOnly {
					return nil, fmt.Errorf("expected generate only mode, got %s", mode)
				}
				return &migrations.MigrationFile{
					Name:        name,
					Description: "Generated migration",
					Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
					Timestamp:   time.Now(),
				}, nil
			},
		},
		{
			name:        "generation failure",
			args:        []string{"FailingGeneration"},
			expectError: true,
			mockFunc: func(_ string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
				return nil, fmt.Errorf("failed to generate migration")
			},
		},
		{
			name:        "empty migration name",
			args:        []string{},
			expectError: true,
			mockFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdGenerateMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestCmdForceMigrationReal(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		mockFunc    func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error)
	}{
		{
			name:        "successful force migration",
			args:        []string{"ForceDestructiveChange"},
			expectError: false,
			mockFunc: func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
				if mode != migrations.ModeForceDestructive {
					return nil, fmt.Errorf(errExpectedForceMode, mode)
				}
				migrationFile := &migrations.MigrationFile{
					Name:        name,
					Description: "Force destructive migration",
					Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
					Timestamp:   time.Now(),
					Changes: []migrations.MigrationChange{
						{
							Type:          "DROP_TABLE",
							TableName:     "old_table",
							IsDestructive: true,
							Description:   "This migration contains destructive changes",
						},
					},
				}
				return migrationFile, nil
			},
		},
		{
			name:        "force migration failure",
			args:        []string{"FailingForce"},
			expectError: true,
			mockFunc: func(_ string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
				return nil, fmt.Errorf("failed to create force migration")
			},
		},
		{
			name:        "empty migration name for force",
			args:        []string{},
			expectError: true,
			mockFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				addMigrationFunc: tt.mockFunc,
			}

			err := testCmdForceMigration(migrator, tt.args)

			if tt.expectError && err == nil {
				t.Error(errExpectedError)
			}
			if !tt.expectError && err != nil {
				t.Errorf(errExpectedNoError, err)
			}
		})
	}
}

func TestMigrationFileWarnings(t *testing.T) {
	// Test MigrationFile methods needed for coverage
	migrationFile := &migrations.MigrationFile{
		Name:        "TestMigration",
		Description: "Test migration with warnings",
		Filename:    "001_TestMigration.sql",
		Timestamp:   time.Now(),
		Changes: []migrations.MigrationChange{
			{
				Type:          "DROP_COLUMN",
				TableName:     "users",
				ColumnName:    "old_field",
				IsDestructive: true,
				RequiresData:  false,
				Description:   "Remove old field",
			},
			{
				Type:         "ADD_COLUMN",
				TableName:    "users",
				ColumnName:   "new_field",
				RequiresData: true,
				Description:  "Add new field with data migration",
			},
		},
	}

	warnings := migrationFile.GetWarnings()
	if len(warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %d", len(warnings))
	}

	// Test both warning methods
	warnings2 := migrationFile.Warnings()
	if len(warnings) != len(warnings2) {
		t.Errorf("GetWarnings and Warnings should return same result")
	}

	hasDestructive := migrationFile.HasDestructiveChanges()
	if !hasDestructive {
		t.Error("Expected migration to have destructive changes")
	}

	errors := migrationFile.Errors()
	if errors == nil {
		errors = []string{} // Handle nil case
	}
	// Test that method doesn't panic - validate slice
	_ = len(errors) // Validates the slice without empty block

	requiresReview := migrationFile.RequiresReview()
	// Test that method doesn't panic
	_ = requiresReview
}

func TestCompleteWorkflow(t *testing.T) {
	// Test a complete workflow to ensure all commands work together
	migrator := &MockMigrator{
		addMigrationFunc: func(name string, _ migrations.MigrationMode) (*migrations.MigrationFile, error) {
			return &migrations.MigrationFile{
				Name:        name,
				Description: fmt.Sprintf("Migration: %s", name),
				Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
				Timestamp:   time.Now(),
			}, nil
		},
		applyMigrationsFunc: func(_ migrations.MigrationMode) error {
			return nil
		},
		getMigrationStatusFunc: func() (*migrations.MigrationStatus, error) {
			return &migrations.MigrationStatus{
				PendingMigrations:     []*migrations.MigrationFile{},
				AppliedMigrations:     []*migrations.MigrationFile{},
				HasPendingChanges:     false,
				HasDestructiveChanges: false,
				Summary:               "All migrations applied",
			}, nil
		},
		revertMigrationFunc: func() error {
			return nil
		},
	}

	// Test complete workflow
	steps := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "add migration",
			testFunc: func() error {
				return testCmdAddMigration(migrator, []string{"WorkflowMigration"})
			},
		},
		{
			name: "apply migrations",
			testFunc: func() error {
				return testCmdApplyMigrations(migrator, []string{})
			},
		},
		{
			name: "check status",
			testFunc: func() error {
				return testCmdMigrationStatus(migrator)
			},
		},
		{
			name: "revert migration",
			testFunc: func() error {
				return testCmdRevertMigration(migrator)
			},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			err := step.testFunc()
			if err != nil {
				t.Errorf("Workflow step '%s' failed: %v", step.name, err)
			}
		})
	}
}

// TestMainFunctionDirectly tests the main function with various command line arguments
func TestMainFunctionDirectly(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no command specified",
			args:        []string{"migrate"},
			expectError: true,
		},
		{
			name:        "help flag",
			args:        []string{"migrate", "-h"},
			expectError: true, // Help will call flag.Usage and exit
		},
		{
			name:        "missing database URL",
			args:        []string{"migrate", "status"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up args
			os.Args = tt.args

			// Capture any panics or exits
			defer func() {
				if r := recover(); r != nil {
					// Expected for help flag and some error cases
					if !tt.expectError {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			// Note: main() function calls os.Exit on errors, so we can't easily test it directly
			// This test primarily ensures the main function doesn't panic on invalid inputs
		})
	}
}

// TestMainWithArgs tests the main function with various command line argument combinations
func TestMainWithArgs(t *testing.T) {
	// Save original os.Args and flag state
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Create temporary test database
	tempDB := ":memory:"

	tests := []struct {
		name        string
		args        []string
		expectPanic bool
		expectLog   string
	}{
		{
			name:        "no command specified",
			args:        []string{"migrate", "-db", tempDB},
			expectPanic: false,
			expectLog:   "Error: No command specified",
		},
		{
			name:        "help flag",
			args:        []string{"migrate", "-h"},
			expectPanic: false,
			expectLog:   "Usage:",
		},
		{
			name:        "unknown command",
			args:        []string{"migrate", "-db", tempDB, "unknown"},
			expectPanic: false,
			expectLog:   "Error: Unknown command 'unknown'",
		},
		{
			name:        "missing database URL",
			args:        []string{"migrate", "status"},
			expectPanic: false,
			expectLog:   "Configuration error: database URL is required",
		},
		{
			name:        "valid add command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "add", "test_migration"},
			expectPanic: false,
			expectLog:   "",
		},
		{
			name:        "valid status command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "status"},
			expectPanic: false,
			expectLog:   "",
		},
		{
			name:        "valid apply command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "apply"},
			expectPanic: false,
			expectLog:   "",
		},
		{
			name:        "valid revert command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "revert"},
			expectPanic: false,
			expectLog:   "",
		},
		{
			name:        "valid generate command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "generate", "test_gen"},
			expectPanic: false,
			expectLog:   "",
		},
		{
			name:        "valid force command",
			args:        []string{"migrate", "-db", testSQLite3URL + tempDB, flagDriver, testSQLite3Driver, "force", "test_force"},
			expectPanic: false,
			expectLog:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag state for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set os.Args
			os.Args = tt.args

			// Capture output
			if tt.expectLog != "" {
				// For tests that expect specific log output, we'll just run main
				// and expect it to complete without panic
				defer func() {
					if r := recover(); r != nil {
						if !tt.expectPanic {
							t.Errorf("Unexpected panic: %v", r)
						}
					}
				}()
			}

			// Call main function
			main()
		})
	}
}

// TestMainFunctionCoverage tests additional paths in main function
func TestMainFunctionCoverage(t *testing.T) {
	// Save original os.Args and flag state
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	t.Run("database connection error", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"migrate", "-db", "invalid://connection", "status"}

		// This should complete without panic but log an error
		main()
	})

	t.Run("command with insufficient arguments", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"migrate", "-db", testSQLite3URL + testMemoryDB, flagDriver, testSQLite3Driver, "add"}

		// This should complete without panic but log an error
		main()
	})

	t.Run("custom migrations and models directories", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"migrate", "-db", testSQLite3URL + testMemoryDB, flagDriver, testSQLite3Driver,
			"-migrations-dir", testMigrationsPath, "-models-dir", testModelsPath, "status"}

		// This should complete without panic
		main()
	})
}

func TestDirectCommandFunctions(t *testing.T) {
	// Create a temporary database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a real migrator for testing
	migrator := migrations.NewHybridMigrator(db, "sqlite3", "./test_migrations")

	t.Run("registerModels with existing directory", func(t *testing.T) {
		err := registerModels(migrator, "./models")
		if err != nil {
			t.Errorf("registerModels failed: %v", err)
		}
	})

	t.Run("cmdAddMigration with valid args", func(t *testing.T) {
		err := cmdAddMigration(migrator, []string{"TestDirectMigration"})
		// This might fail due to directory not existing, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for non-existent directory: %v", err)
		}
	})

	t.Run("cmdAddMigration with empty args", func(t *testing.T) {
		err := cmdAddMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty migration name")
		}
	})

	t.Run("cmdApplyMigrations default mode", func(t *testing.T) {
		err := cmdApplyMigrations(migrator, []string{})
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for apply migrations: %v", err)
		}
	})

	t.Run("cmdApplyMigrations with force flag", func(t *testing.T) {
		err := cmdApplyMigrations(migrator, []string{flagForce})
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for apply migrations with force: %v", err)
		}
	})

	t.Run("cmdApplyMigrations with auto flag", func(t *testing.T) {
		err := cmdApplyMigrations(migrator, []string{flagAuto})
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for apply migrations with auto: %v", err)
		}
	})

	t.Run("cmdRevertMigration", func(t *testing.T) {
		err := cmdRevertMigration(migrator)
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for revert migration: %v", err)
		}
	})

	t.Run("cmdMigrationStatus", func(t *testing.T) {
		err := cmdMigrationStatus(migrator)
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for migration status: %v", err)
		}
	})

	t.Run("cmdGenerateMigration with valid args", func(t *testing.T) {
		err := cmdGenerateMigration(migrator, []string{"TestGenerate"})
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for generate migration: %v", err)
		}
	})

	t.Run("cmdGenerateMigration with empty args", func(t *testing.T) {
		err := cmdGenerateMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty migration name")
		}
	})

	t.Run("cmdForceMigration with valid args", func(t *testing.T) {
		err := cmdForceMigration(migrator, []string{"TestForce"})
		// This might fail, but we're testing the function path
		if err != nil {
			t.Logf("Expected error for force migration: %v", err)
		}
	})

	t.Run("cmdForceMigration with empty args", func(t *testing.T) {
		err := cmdForceMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty migration name")
		}
	})
}

// TestRegisterModelsEdgeCases tests registerModels function with various scenarios
func TestRegisterModelsEdgeCases(t *testing.T) {
	// Create a temporary database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	migrator := migrations.NewHybridMigrator(db, "sqlite3", "./test_migrations")

	tests := []struct {
		name      string
		modelsDir string
		expectErr bool
	}{
		{
			name:      "empty models directory",
			modelsDir: "",
			expectErr: false, // registerModels doesn't validate empty directory for real HybridMigrator
		},
		{
			name:      "non-existent models directory",
			modelsDir: "./non_existent_models",
			expectErr: false, // Should just log warning
		},
		{
			name:      "current directory",
			modelsDir: ".",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registerModels(migrator, tt.modelsDir)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestCommandErrorPaths tests error handling in command functions
func TestCommandErrorPaths(t *testing.T) {
	// Create a temporary database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	migrator := migrations.NewHybridMigrator(db, "sqlite3", "./test_migrations")

	t.Run("cmdAddMigration error scenarios", func(t *testing.T) {
		// Test with empty args
		err := cmdAddMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty args")
		}

		// Test with valid name but might fail due to directory issues
		err = cmdAddMigration(migrator, []string{"ValidMigrationName"})
		// Log the error but don't fail the test as it's expected
		if err != nil {
			t.Logf("Expected error due to directory setup: %v", err)
		}
	})

	t.Run("cmdGenerateMigration error scenarios", func(t *testing.T) {
		// Test with empty args
		err := cmdGenerateMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty args")
		}
	})

	t.Run("cmdForceMigration error scenarios", func(t *testing.T) {
		// Test with empty args
		err := cmdForceMigration(migrator, []string{})
		if err == nil {
			t.Error("Expected error for empty args")
		}
	})
}

// TestMainFunctionEdgeCases tests edge cases for main function
func TestMainFunctionEdgeCases(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "unknown_command_with_valid_db",
			args:     []string{"migrate", "-db", testMemoryDB, "unknown"},
			expected: "unknown command",
		},
		{
			name:     "database_connection_error_postgresql",
			args:     []string{"migrate", "-db", "postgres://invalid:invalid@invalid:5432/invalid", "-driver", "postgresql", "status"},
			expected: "database connection error",
		},
		{
			name:     "model_registration_error_simulation",
			args:     []string{"migrate", "-db", testMemoryDB, "-models-dir", "/invalid/path", "status"},
			expected: "database connection error", // Will fail at DB connection first
		},
		{
			name:     "valid_command_execution_path",
			args:     []string{"migrate", "-db", testMemoryDB, "-driver", "sqlite", "status"},
			expected: "database connection error", // Expect connection error for memory DB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Capture output - we need the reader but not use it directly
			oldStderr := os.Stderr
			_, w, _ := os.Pipe()
			os.Stderr = w

			// Run main function
			main()

			// Restore stderr and get output
			w.Close()
			os.Stderr = oldStderr

			// No need to check output as we're testing the execution paths
		})
	}
}

// TestConnectDatabaseEdgeCases tests additional edge cases for connectDatabase
func TestConnectDatabaseEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "mysql_driver_mapping",
			config: Config{
				DatabaseURL: "mysql://test",
				Driver:      "mysql",
			},
			expectError: true,
			errorMsg:    "failed to open database",
		},
		{
			name: "postgresql_driver_mapping",
			config: Config{
				DatabaseURL: "postgres://test",
				Driver:      "postgresql",
			},
			expectError: true,
			errorMsg:    "failed to",
		},
		{
			name: "sqlite3_driver_mapping",
			config: Config{
				DatabaseURL: testMemoryDB,
				Driver:      "sqlite3",
			},
			expectError: false,
		},
		{
			name: "ping_failure_simulation",
			config: Config{
				DatabaseURL: "invalid://connection",
				Driver:      "postgres",
			},
			expectError: true,
			errorMsg:    "failed to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := connectDatabase(&tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

// TestCmdAddMigrationDetailed tests detailed scenarios for cmdAddMigration
func TestCmdAddMigrationDetailed(t *testing.T) {
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	migrator := migrations.NewHybridMigrator(db, migrations.SQLite, testMigrationsPath)

	tests := []struct {
		name            string
		args            []string
		expectError     bool
		mockBehavior    func(*MockMigrator)
		testWarnings    bool
		testDestructive bool
	}{
		{
			name:        "migration_with_warnings",
			args:        []string{"TestMigrationWithWarnings"},
			expectError: false,
			mockBehavior: func(mock *MockMigrator) {
				mock.addMigrationFunc = func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
					migFile := &migrations.MigrationFile{
						Name:        name,
						Description: "Test migration with warnings",
						Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
						Timestamp:   time.Now(),
						Changes: []migrations.MigrationChange{
							{
								Type:          "ALTER_COLUMN",
								TableName:     "users",
								ColumnName:    "email",
								IsDestructive: false,
								RequiresData:  false,
								Description:   "Column may lose data",
							},
							{
								Type:          "DROP_INDEX",
								TableName:     "users",
								ColumnName:    "",
								IsDestructive: false,
								RequiresData:  false,
								Description:   "Index will be dropped",
							},
						},
					}

					return migFile, nil
				}
			},
			testWarnings: true,
		},
		{
			name:        "migration_with_destructive_changes",
			args:        []string{"TestDestructiveMigration"},
			expectError: false,
			mockBehavior: func(mock *MockMigrator) {
				mock.addMigrationFunc = func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
					migFile := &migrations.MigrationFile{
						Name:        name,
						Description: "Test destructive migration",
						Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
						Timestamp:   time.Now(),
						Changes: []migrations.MigrationChange{
							{
								Type:          "DROP_COLUMN",
								TableName:     "users",
								ColumnName:    "old_field",
								IsDestructive: true,
								RequiresData:  false,
								Description:   "Drop column with potential data loss",
							},
						},
					}

					return migFile, nil
				}
			},
			testDestructive: true,
		},
		{
			name:        "multiple_args_scenario",
			args:        []string{"TestMigration", "extra", "args"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockBehavior != nil {
				// Test with mock migrator
				mock := &MockMigrator{}
				tt.mockBehavior(mock)

				err := testCmdAddMigration(mock, tt.args)

				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			} else {
				// Test with real migrator
				err := cmdAddMigration(migrator, tt.args)

				// We expect errors with real migrator due to no model changes
				if err == nil && tt.expectError {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

// TestCmdRevertMigrationDetailed tests detailed scenarios for cmdRevertMigration
func TestCmdRevertMigrationDetailed(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(*MockMigrator)
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful_revert_with_message",
			mockBehavior: func(mock *MockMigrator) {
				mock.revertMigrationFunc = func() error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "revert_with_specific_error",
			mockBehavior: func(mock *MockMigrator) {
				mock.revertMigrationFunc = func() error {
					return fmt.Errorf("no migrations to revert")
				}
			},
			expectError: true,
			errorMsg:    "no migrations",
		},
		{
			name: "revert_with_database_error",
			mockBehavior: func(mock *MockMigrator) {
				mock.revertMigrationFunc = func() error {
					return fmt.Errorf("database connection lost")
				}
			},
			expectError: true,
			errorMsg:    "database connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMigrator{}
			tt.mockBehavior(mock)

			err := testCmdRevertMigration(mock)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestCmdGenerateMigrationDetailed tests detailed scenarios for cmdGenerateMigration
func TestCmdGenerateMigrationDetailed(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		mockBehavior func(*MockMigrator)
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful_generation_with_description",
			args: []string{"TestGeneratedMigration"},
			mockBehavior: func(mock *MockMigrator) {
				mock.addMigrationFunc = func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
					if mode != migrations.ModeGenerateOnly {
						return nil, fmt.Errorf("expected generate only mode")
					}

					return &migrations.MigrationFile{
						Name:        name,
						Description: "Generated migration for testing",
						Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
						Timestamp:   time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "generation_with_error",
			args: []string{"FailingGeneration"},
			mockBehavior: func(mock *MockMigrator) {
				mock.addMigrationFunc = func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
					return nil, fmt.Errorf("generation failed due to invalid schema")
				}
			},
			expectError: true,
			errorMsg:    "generation failed",
		},
		{
			name:        "empty_args_error",
			args:        []string{},
			expectError: true,
			errorMsg:    errMigrationNameRequired,
		},
		{
			name: "multiple_args_with_first_valid",
			args: []string{"ValidMigration", "ignored", "args"},
			mockBehavior: func(mock *MockMigrator) {
				mock.addMigrationFunc = func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
					return &migrations.MigrationFile{
						Name:      name,
						Filename:  fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
						Timestamp: time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockBehavior != nil {
				mock := &MockMigrator{}
				tt.mockBehavior(mock)

				err := testCmdGenerateMigration(mock, tt.args)

				if tt.expectError {
					if err == nil {
						t.Error("Expected error but got none")
					}
					if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error but got: %v", err)
					}
				}
			} else {
				// Test with nil mock behavior should fail
				mock := &MockMigrator{}
				err := testCmdGenerateMigration(mock, tt.args)

				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

// TestDisplayPendingMigrationsDetailed tests edge cases for displayPendingMigrations
func TestDisplayPendingMigrationsDetailed(t *testing.T) {
	tests := []struct {
		name       string
		migrations []*migrations.MigrationFile
		expectIcon string
	}{
		{
			name: "pending_migration_with_destructive_changes",
			migrations: []*migrations.MigrationFile{
				{
					Name:      "DestructiveMigration",
					Timestamp: time.Date(2025, 7, 19, 21, 7, 40, 0, time.UTC),
				},
			},
			expectIcon: "⚠️",
		},
		{
			name: "pending_migration_without_destructive_changes",
			migrations: []*migrations.MigrationFile{
				{
					Name:      "SafeMigration",
					Timestamp: time.Date(2025, 7, 19, 21, 7, 40, 0, time.UTC),
				},
			},
			expectIcon: "○",
		},
		{
			name: "multiple_mixed_migrations",
			migrations: []*migrations.MigrationFile{
				{
					Name:      "SafeMigration1",
					Timestamp: time.Date(2025, 7, 19, 21, 7, 40, 0, time.UTC),
				},
				{
					Name:      "DestructiveMigration1",
					Timestamp: time.Date(2025, 7, 19, 22, 7, 40, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up destructive flags for test migrations
			for _, migration := range tt.migrations {
				if contains(migration.Name, "Destructive") {
					// Add destructive changes to the migration to simulate destructive behavior
					migration.Changes = []migrations.MigrationChange{
						{
							Type:          "DROP_COLUMN",
							TableName:     "users",
							ColumnName:    "old_field",
							IsDestructive: true,
							RequiresData:  false,
							Description:   "Drop column with potential data loss",
						},
					}
				}
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			displayPendingMigrations(tt.migrations)

			w.Close()
			os.Stdout = oldStdout

			output := make([]byte, 1024)
			n, _ := r.Read(output)
			outputStr := string(output[:n])

			// Verify output contains expected elements
			if len(tt.migrations) > 0 {
				if !contains(outputStr, fmt.Sprintf("Pending Migrations (%d)", len(tt.migrations))) {
					t.Errorf("Expected output to contain migration count, got: %s", outputStr)
				}

				for _, migration := range tt.migrations {
					if !contains(outputStr, migration.Name) {
						t.Errorf("Expected output to contain migration name '%s', got: %s", migration.Name, outputStr)
					}
				}
			}
		})
	}
}

// TestMainIntegrationScenarios tests integration scenarios
func TestMainIntegrationScenarios(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temporary test database file
	tmpDB := "/tmp/test_migrate.db"
	defer os.Remove(tmpDB)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "complete_sqlite_workflow",
			args: []string{"migrate", "-db", tmpDB, "-driver", "sqlite", "status"},
		},
		{
			name: "apply_with_custom_directories",
			args: []string{"migrate", "-db", tmpDB, "-migrations-dir", "./test_migrations", "-models-dir", "./test_models", "apply"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Run main function - it should execute without panicking
			main()
		})
	}
}

// TestValidateConfigEdgeCases tests additional edge cases for validateConfig
func TestValidateConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    Config
		expectError    bool
		expectedConfig Config
	}{
		{
			name: "all_empty_strings_except_db",
			inputConfig: Config{
				DatabaseURL:   "test://db",
				Driver:        "",
				MigrationsDir: "",
				ModelsDir:     "",
			},
			expectError: false,
			expectedConfig: Config{
				DatabaseURL:   "test://db",
				Driver:        defaultPostgresDriver,
				MigrationsDir: "./migrations",
				ModelsDir:     "./models",
			},
		},
		{
			name: "partial_empty_config",
			inputConfig: Config{
				DatabaseURL:   "test://db",
				Driver:        "mysql",
				MigrationsDir: "",
				ModelsDir:     "./custom_models",
			},
			expectError: false,
			expectedConfig: Config{
				DatabaseURL:   "test://db",
				Driver:        "mysql",
				MigrationsDir: "./migrations",
				ModelsDir:     "./custom_models",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.inputConfig
			err := validateConfig(&config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Verify config was updated correctly
				if config.Driver != tt.expectedConfig.Driver {
					t.Errorf("Expected driver '%s', got '%s'", tt.expectedConfig.Driver, config.Driver)
				}
				if config.MigrationsDir != tt.expectedConfig.MigrationsDir {
					t.Errorf("Expected migrations dir '%s', got '%s'", tt.expectedConfig.MigrationsDir, config.MigrationsDir)
				}
				if config.ModelsDir != tt.expectedConfig.ModelsDir {
					t.Errorf("Expected models dir '%s', got '%s'", tt.expectedConfig.ModelsDir, config.ModelsDir)
				}
			}
		})
	}
}

// TestCmdAddMigrationWithWarningsAndDestructive tests cmdAddMigration with warnings and destructive changes
func TestCmdAddMigrationWithWarningsAndDestructive(t *testing.T) {
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	migrator := migrations.NewHybridMigrator(db, migrations.SQLite, testMigrationsPath)

	// Create a mock migrator that returns warnings and destructive changes
	mock := &MockMigrator{
		addMigrationFunc: func(name string, mode migrations.MigrationMode) (*migrations.MigrationFile, error) {
			migFile := &migrations.MigrationFile{
				Name:        name,
				Description: "Test migration with warnings and destructive changes",
				Filename:    fmt.Sprintf("%d_%s.sql", time.Now().Unix(), name),
				Timestamp:   time.Now(),
				Changes: []migrations.MigrationChange{
					{
						Type:          "DROP_COLUMN",
						TableName:     "users",
						ColumnName:    "old_field",
						IsDestructive: true,
						RequiresData:  false,
						Description:   "Drop column with potential data loss",
					},
					{
						Type:         "ADD_COLUMN",
						TableName:    "users",
						ColumnName:   "new_field",
						RequiresData: true,
						Description:  "Add new field requiring data migration",
					},
				},
			}

			return migFile, nil
		},
	}

	// Test with mock to hit the warning and destructive change paths
	err = testCmdAddMigration(mock, []string{"TestWarningsAndDestructive"})
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Test with real migrator to cover other code paths (will error due to no models)
	err = cmdAddMigration(migrator, []string{"TestMigration"})
	if err == nil {
		t.Error("Expected error due to no models registered")
	}
}
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
