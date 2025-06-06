package migrations

import (
	"fmt"
	"reflect"
	"time"
)

// ChangeType represents the type of migration change.
type ChangeType string

const (
	// CreateTable indicates a table creation operation.
	CreateTable ChangeType = "CreateTable"
	// DropTable indicates a table drop operation.
	DropTable ChangeType = "DropTable"
	// AddColumn indicates a column addition operation.
	AddColumn ChangeType = "AddColumn"
	// DropColumn indicates a column drop operation.
	DropColumn ChangeType = "DropColumn"
	// AlterColumn indicates a column alteration operation.
	AlterColumn ChangeType = "AlterColumn"
	// RenameColumn indicates a column rename operation.
	RenameColumn ChangeType = "RenameColumn"
	// AddIndex indicates an index addition operation.
	AddIndex ChangeType = "AddIndex"
	// CreateIndex is an alias for AddIndex.
	CreateIndex ChangeType = "CreateIndex" // Alias for AddIndex
	// DropIndex indicates an index drop operation.
	DropIndex ChangeType = "DropIndex"
	// AddConstraint indicates a constraint addition operation.
	AddConstraint ChangeType = "AddConstraint"
	// DropConstraint indicates a constraint drop operation.
	DropConstraint ChangeType = "DropConstraint"
)

// MigrationMode defines how migrations should be applied.
type MigrationMode int

const (
	// ModeAutomatic applies only safe changes.
	ModeAutomatic MigrationMode = iota
	// ModeInteractive prompts for destructive changes.
	ModeInteractive
	// ModeGenerateOnly generates SQL files, doesn't apply them.
	ModeGenerateOnly
	// ModeForceDestructive applies all changes automatically.
	ModeForceDestructive

	// Automatic is an alias for ModeAutomatic.
	// Automatic applies only safe changes (alias for ModeAutomatic).
	Automatic = ModeAutomatic
	// Interactive is an alias for ModeInteractive.
	// Interactive prompts for destructive changes (alias for ModeInteractive).
	Interactive = ModeInteractive
	// GenerateOnly is an alias for ModeGenerateOnly.
	// GenerateOnly generates SQL files, doesn't apply them (alias for ModeGenerateOnly).
	GenerateOnly = ModeGenerateOnly
	// ForceDestructive is an alias for ModeForceDestructive.
	// ForceDestructive applies all changes automatically (alias for ModeForceDestructive).
	ForceDestructive = ModeForceDestructive
)

// String returns the string representation of MigrationMode
func (m MigrationMode) String() string {
	switch m {
	case ModeAutomatic:
		return "Automatic"
	case ModeInteractive:
		return "Interactive"
	case ModeGenerateOnly:
		return "GenerateOnly"
	case ModeForceDestructive:
		return "ForceDestructive"
	default:
		return "Unknown"
	}
}

// ParseMigrationMode parses a string into MigrationMode
func ParseMigrationMode(s string) MigrationMode {
	switch s {
	case "Automatic":
		return ModeAutomatic
	case "Interactive":
		return ModeInteractive
	case "GenerateOnly":
		return ModeGenerateOnly
	case "ForceDestructive":
		return ModeForceDestructive
	default:
		return ModeAutomatic // Default fallback
	}
}

// ColumnInfo represents database column information
type ColumnInfo struct {
	Name         string
	Type         string
	SQLType      string
	DataType     string // Additional field for DataType
	Nullable     bool
	IsNullable   bool // Additional field for IsNullable
	Default      *string
	DefaultValue *string // Additional field for DefaultValue
	IsPrimaryKey bool
	IsUnique     bool
	IsIdentity   bool // Additional field for auto-increment/identity columns
	IsForeignKey bool
	References   *ForeignKeyInfo
	Size         int
	MaxLength    *int                       // Change to pointer for nil comparison
	Precision    *int                       // Change to pointer for nil comparison
	Scale        *int                       // Change to pointer for nil comparison
	Constraints  map[string]*ConstraintInfo // Additional field for Constraints
}

// ForeignKeyInfo represents foreign key relationship
type ForeignKeyInfo struct {
	Table  string
	Column string
}

// IndexInfo represents database index information
type IndexInfo struct {
	Name     string
	Columns  []string
	Unique   bool
	IsUnique bool   // Additional field for IsUnique
	Type     string // "btree", "hash", etc.
}

// ConstraintInfo represents database constraint information
type ConstraintInfo struct {
	Name              string
	Type              string // "CHECK", "UNIQUE", "FOREIGN_KEY"
	SQL               string
	ReferencedTable   string   // Additional field for ReferencedTable
	Columns           []string // Additional field for Columns
	ReferencedColumns []string // Additional field for ReferencedColumns
}

// ModelSnapshot represents the complete schema of a table
type ModelSnapshot struct {
	TableName   string
	ModelType   reflect.Type
	Columns     map[string]*ColumnInfo // Using pointers for consistency
	Indexes     map[string]IndexInfo
	Constraints map[string]*ConstraintInfo // Using pointers for consistency
	Checksum    string
}

// MigrationChange represents a single change to be applied
type MigrationChange struct {
	Type          ChangeType
	TableName     string
	ColumnName    string
	IndexName     string // For index operations
	ModelName     string // Model name for reference
	OldColumn     *ColumnInfo
	NewColumn     *ColumnInfo
	OldTable      *ModelSnapshot
	NewTable      *ModelSnapshot
	OldValue      interface{} // For alter operations
	NewValue      interface{} // For alter operations
	SQL           []string
	DownSQL       []string
	IsDestructive bool
	RequiresData  bool
	Description   string
}

// MigrationFile represents a generated migration file
type MigrationFile struct {
	Version     string
	Name        string
	Description string
	UpSQL       []string
	DownSQL     []string
	Filename    string
	FilePath    string
	Timestamp   time.Time
	Changes     []MigrationChange
	Checksum    string
	Mode        MigrationMode
	// ParsedHasDestructive stores the destructive flag parsed from file metadata
	// when Changes slice is not available (e.g., when loading from disk)
	ParsedHasDestructive *bool
}

// HasDestructiveChanges returns true if any change is destructive
func (mf *MigrationFile) HasDestructiveChanges() bool {
	// If we have Changes populated, use them for calculation
	if len(mf.Changes) > 0 {
		for _, change := range mf.Changes {
			if change.IsDestructive {
				return true
			}
		}
		return false
	}

	// If Changes are not available (e.g., when loaded from disk),
	// use the parsed flag from file metadata
	if mf.ParsedHasDestructive != nil {
		return *mf.ParsedHasDestructive
	}

	// Default to false if neither Changes nor parsed flag is available
	return false
}

// HasDestructive is an alias for HasDestructiveChanges
func (mf *MigrationFile) HasDestructive() bool {
	return mf.HasDestructiveChanges()
}

// RequiresReview returns true if the migration requires manual review
func (mf *MigrationFile) RequiresReview() bool {
	return mf.HasDestructiveChanges()
}

// GetWarnings returns warnings about the migration
func (mf *MigrationFile) GetWarnings() []string {
	var warnings []string
	for _, change := range mf.Changes {
		if change.IsDestructive {
			warnings = append(warnings, fmt.Sprintf("Destructive change: %s on %s.%s",
				change.Type, change.TableName, change.ColumnName))
		}
		if change.RequiresData {
			warnings = append(warnings, fmt.Sprintf("Data migration required: %s",
				change.Description))
		}
	}
	return warnings
}

// Warnings is an alias for GetWarnings
func (mf *MigrationFile) Warnings() []string {
	return mf.GetWarnings()
}

// Errors returns any errors found during migration planning
func (mf *MigrationFile) Errors() []string {
	var errors []string
	// For now, errors are determined by validation logic
	// This could be expanded to include specific error conditions
	return errors
}

// ModelRegistry manages registered models for migration operations.
type ModelRegistry struct {
	models map[string]*ModelSnapshot
	driver DatabaseDriver
}

// DatabaseDriver represents the type of database (e.g., PostgreSQL, MySQL, SQLite).
type DatabaseDriver string

const (
	// PostgreSQL is the constant for the PostgreSQL database driver.
	PostgreSQL DatabaseDriver = "postgres"
	// MySQL is the constant for the MySQL database driver.
	MySQL DatabaseDriver = "mysql"
	// SQLite is the constant for the SQLite database driver.
	SQLite DatabaseDriver = "sqlite3"
)
