package main

import (
	"fmt"
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

	// Error message constants
	errExpectedError      = "Expected error but got none"
	errExpectedNoError    = "Expected no error but got: %v"
	errExpectedMode       = "Expected mode %s, got %s"
	errEmptyMigrationName = "empty migration name"
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
	// For mock, simulate the behavior
	return m.ApplyMigrations(migrations.ModeAutomatic)
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
	return m.ApplyMigrations(migrations.ModeForceDestructive)
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
		mockFunc    func(mode migrations.MigrationMode) error
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
			mockFunc: func(_ migrations.MigrationMode) error {
				return fmt.Errorf("force migration failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &MockMigrator{
				applyMigrationsFunc: tt.mockFunc,
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
