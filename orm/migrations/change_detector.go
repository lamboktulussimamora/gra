package migrations

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

const foreignKeyConstraintType = "FOREIGN KEY"

// ChangeDetector detects schema changes between model snapshots and database state
type ChangeDetector struct {
	registry  *ModelRegistry
	inspector *DatabaseInspector
}

// NewChangeDetector creates a new change detector
func NewChangeDetector(registry *ModelRegistry, inspector *DatabaseInspector) *ChangeDetector {
	return &ChangeDetector{
		registry:  registry,
		inspector: inspector,
	}
}

// DetectChanges compares current model state with database and returns migration changes
func (cd *ChangeDetector) DetectChanges() (*MigrationPlan, error) {
	// Get current model snapshots
	modelSnapshots := cd.registry.GetModels()

	// Get current database schema
	dbSchema, err := cd.inspector.GetCurrentSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to read database schema: %w", err)
	}

	// Compare and generate changes
	changes, err := cd.inspector.CompareWithModelSnapshot(dbSchema, modelSnapshots)
	if err != nil {
		return nil, fmt.Errorf("failed to compare schemas: %w", err)
	}

	// Create migration plan
	plan := &MigrationPlan{
		Changes:        changes,
		ModelSnapshots: modelSnapshots,
		DatabaseSchema: dbSchema,
		PlanChecksum:   cd.calculatePlanChecksum(changes),
		HasDestructive: cd.hasDestructiveChanges(changes),
		RequiresReview: cd.requiresManualReview(changes),
	}

	// Sort changes by dependency order
	cd.sortChangesByDependency(plan.Changes)

	return plan, nil
}

// MigrationPlan represents a complete migration plan
type MigrationPlan struct {
	Changes        []MigrationChange
	ModelSnapshots map[string]*ModelSnapshot
	DatabaseSchema map[string]*TableSchema
	PlanChecksum   string
	HasDestructive bool
	RequiresReview bool
	Warnings       []string
	Errors         []string
}

// calculatePlanChecksum creates a checksum for the entire migration plan
func (cd *ChangeDetector) calculatePlanChecksum(changes []MigrationChange) string {
	hasher := md5.New()

	// Sort changes for consistent checksum
	sortedChanges := make([]MigrationChange, len(changes))
	copy(sortedChanges, changes)
	sort.Slice(sortedChanges, func(i, j int) bool {
		return cd.compareChanges(sortedChanges[i], sortedChanges[j])
	})

	for _, change := range sortedChanges {
		hasher.Write([]byte(cd.changeToString(change)))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// changeToString converts a migration change to a string for hashing
func (cd *ChangeDetector) changeToString(change MigrationChange) string {
	parts := []string{
		string(change.Type),
		change.TableName,
		change.ModelName,
		change.ColumnName,
		change.IndexName,
	}
	return strings.Join(parts, "|")
}

// compareChanges provides ordering for migration changes
func (cd *ChangeDetector) compareChanges(a, b MigrationChange) bool {
	// Primary sort by type priority
	aPriority := cd.getChangeTypePriority(a.Type)
	bPriority := cd.getChangeTypePriority(b.Type)

	if aPriority != bPriority {
		return aPriority < bPriority
	}

	// Secondary sort by table name
	if a.TableName != b.TableName {
		return a.TableName < b.TableName
	}

	// Tertiary sort by column/index name
	if a.ColumnName != b.ColumnName {
		return a.ColumnName < b.ColumnName
	}

	return a.IndexName < b.IndexName
}

// getChangeTypePriority returns priority order for change types
func (cd *ChangeDetector) getChangeTypePriority(changeType ChangeType) int {
	priorities := map[ChangeType]int{
		CreateTable: 1,
		AddColumn:   2,
		AlterColumn: 3,
		CreateIndex: 4,
		DropIndex:   5,
		DropColumn:  6,
		DropTable:   7,
	}

	if priority, exists := priorities[changeType]; exists {
		return priority
	}
	return 999
}

// hasDestructiveChanges checks if any changes are potentially destructive
func (cd *ChangeDetector) hasDestructiveChanges(changes []MigrationChange) bool {
	destructiveTypes := map[ChangeType]bool{
		DropTable:   true,
		DropColumn:  true,
		AlterColumn: true, // Can be destructive depending on the change
	}

	for _, change := range changes {
		if destructiveTypes[change.Type] {
			return true
		}
	}
	return false
}

// requiresManualReview determines if changes need manual review
func (cd *ChangeDetector) requiresManualReview(changes []MigrationChange) bool {
	for _, change := range changes {
		switch change.Type {
		case DropTable, DropColumn:
			return true
		case AlterColumn:
			// Check if it's a potentially data-losing change
			if cd.isDataLosingAlterColumn(change) {
				return true
			}
		}
	}
	return false
}

// isDataLosingAlterColumn checks if a column alteration might lose data
func (cd *ChangeDetector) isDataLosingAlterColumn(change MigrationChange) bool {
	if change.Type != AlterColumn {
		return false
	}

	oldColumn, okOld := change.OldValue.(*DatabaseColumnInfo)
	newColumn, okNew := change.NewValue.(*ColumnInfo)

	if !okOld || !okNew {
		return false
	}

	// Check for potentially data-losing changes
	// 1. Making column non-nullable when it was nullable
	if oldColumn.IsNullable && !newColumn.IsNullable {
		return true
	}

	// 2. Reducing string length
	if oldColumn.MaxLength != nil && newColumn.MaxLength != nil {
		if *newColumn.MaxLength < *oldColumn.MaxLength {
			return true
		}
	}

	// 3. Changing data type to incompatible type
	if cd.isIncompatibleTypeChange(oldColumn.DataType, newColumn.DataType) {
		return true
	}

	return false
}

// isIncompatibleTypeChange checks if a type change is incompatible
func (cd *ChangeDetector) isIncompatibleTypeChange(oldType, newType string) bool {
	oldType = strings.ToUpper(strings.TrimSpace(oldType))
	newType = strings.ToUpper(strings.TrimSpace(newType))

	// Define incompatible type changes
	incompatibleChanges := map[string][]string{
		"TEXT":      {"INTEGER", "BIGINT", "BOOLEAN", "TIMESTAMP", "DATE"},
		"VARCHAR":   {"INTEGER", "BIGINT", "BOOLEAN", "TIMESTAMP", "DATE"},
		"INTEGER":   {"BOOLEAN", "TIMESTAMP", "DATE"},
		"BIGINT":    {"BOOLEAN", "TIMESTAMP", "DATE"},
		"BOOLEAN":   {"INTEGER", "BIGINT", "TEXT", "VARCHAR", "TIMESTAMP", "DATE"},
		"TIMESTAMP": {"INTEGER", "BIGINT", "BOOLEAN"},
		"DATE":      {"INTEGER", "BIGINT", "BOOLEAN"},
	}

	if incompatibleTypes, exists := incompatibleChanges[oldType]; exists {
		for _, incompatible := range incompatibleTypes {
			if strings.HasPrefix(newType, incompatible) {
				return true
			}
		}
	}

	return false
}

// sortChangesByDependency sorts changes in dependency order
func (cd *ChangeDetector) sortChangesByDependency(changes []MigrationChange) {
	sort.Slice(changes, func(i, j int) bool {
		return cd.compareChanges(changes[i], changes[j])
	})
}

// ValidateMigrationPlan performs validation checks on a migration plan
func (cd *ChangeDetector) ValidateMigrationPlan(plan *MigrationPlan) error {
	var errors []string
	var warnings []string

	// Check for circular dependencies
	if err := cd.checkCircularDependencies(plan.Changes); err != nil {
		errors = append(errors, fmt.Sprintf("Circular dependency detected: %v", err))
	}

	// Check for orphaned foreign keys
	orphanedFKs := cd.findOrphanedForeignKeys(plan.Changes)
	for _, fk := range orphanedFKs {
		warnings = append(warnings, fmt.Sprintf("Foreign key %s references table that will be dropped", fk))
	}

	// Check for data loss potential
	dataLossChanges := cd.findDataLossChanges(plan.Changes)
	for _, change := range dataLossChanges {
		warnings = append(warnings, fmt.Sprintf("Potential data loss in %s.%s", change.TableName, change.ColumnName))
	}

	plan.Warnings = warnings
	plan.Errors = errors

	if len(errors) > 0 {
		return fmt.Errorf("migration plan validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// checkCircularDependencies checks for circular dependencies in migration changes
func (cd *ChangeDetector) checkCircularDependencies(changes []MigrationChange) error {
	// Build dependency graph
	dependencies := make(map[string][]string)

	for _, change := range changes {
		if change.Type == CreateTable {
			// Tables with foreign keys depend on their referenced tables
			if snapshot, ok := change.NewValue.(*ModelSnapshot); ok {
				for _, constraint := range snapshot.Constraints {
					if constraint.Type == foreignKeyConstraintType && constraint.ReferencedTable != "" {
						dependencies[snapshot.TableName] = append(dependencies[snapshot.TableName], constraint.ReferencedTable)
					}
				}
			}
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for table := range dependencies {
		if !visited[table] {
			if cd.hasCycleDFS(table, dependencies, visited, recursionStack) {
				return fmt.Errorf("circular dependency involving table %s", table)
			}
		}
	}

	return nil
}

// hasCycleDFS performs DFS to detect cycles
func (cd *ChangeDetector) hasCycleDFS(
	table string,
	dependencies map[string][]string,
	visited map[string]bool,
	recursionStack map[string]bool,
) bool {
	visited[table] = true
	recursionStack[table] = true

	for _, dependency := range dependencies[table] {
		if !visited[dependency] {
			if cd.hasCycleDFS(dependency, dependencies, visited, recursionStack) {
				return true
			}
		} else if recursionStack[dependency] {
			return true
		}
	}

	recursionStack[table] = false
	return false
}

// findOrphanedForeignKeys finds foreign keys that reference tables being dropped
func (cd *ChangeDetector) findOrphanedForeignKeys(changes []MigrationChange) []string {
	var orphaned []string

	// Find tables being dropped
	droppedTables := make(map[string]bool)
	for _, change := range changes {
		if change.Type == DropTable {
			droppedTables[change.TableName] = true
		}
	}

	// Check for foreign keys referencing dropped tables
	for _, change := range changes {
		if change.Type == CreateTable || change.Type == AddColumn {
			var constraints map[string]*ConstraintInfo

			if snapshot, ok := change.NewValue.(*ModelSnapshot); ok {
				constraints = snapshot.Constraints
			} else if column, ok := change.NewValue.(*ColumnInfo); ok && len(column.Constraints) > 0 {
				// Handle individual column constraints
				constraints = column.Constraints
			}

			for constraintName, constraint := range constraints {
				if constraint.Type == foreignKeyConstraintType && droppedTables[constraint.ReferencedTable] {
					orphaned = append(orphaned, constraintName)
				}
			}
		}
	}

	return orphaned
}

// findDataLossChanges identifies changes that might cause data loss
func (cd *ChangeDetector) findDataLossChanges(changes []MigrationChange) []MigrationChange {
	var dataLossChanges []MigrationChange

	for _, change := range changes {
		switch change.Type {
		case DropTable, DropColumn:
			dataLossChanges = append(dataLossChanges, change)
		case AlterColumn:
			if cd.isDataLosingAlterColumn(change) {
				dataLossChanges = append(dataLossChanges, change)
			}
		}
	}

	return dataLossChanges
}

// GetChangeSummary returns a human-readable summary of changes
func (cd *ChangeDetector) GetChangeSummary(plan *MigrationPlan) string {
	if len(plan.Changes) == 0 {
		return "No changes detected"
	}

	summary := make(map[ChangeType]int)
	for _, change := range plan.Changes {
		summary[change.Type]++
	}

	var parts []string
	if count, exists := summary[CreateTable]; exists {
		parts = append(parts, fmt.Sprintf("%d table(s) to create", count))
	}
	if count, exists := summary[DropTable]; exists {
		parts = append(parts, fmt.Sprintf("%d table(s) to drop", count))
	}
	if count, exists := summary[AddColumn]; exists {
		parts = append(parts, fmt.Sprintf("%d column(s) to add", count))
	}
	if count, exists := summary[DropColumn]; exists {
		parts = append(parts, fmt.Sprintf("%d column(s) to drop", count))
	}
	if count, exists := summary[AlterColumn]; exists {
		parts = append(parts, fmt.Sprintf("%d column(s) to alter", count))
	}
	if count, exists := summary[CreateIndex]; exists {
		parts = append(parts, fmt.Sprintf("%d index(es) to create", count))
	}
	if count, exists := summary[DropIndex]; exists {
		parts = append(parts, fmt.Sprintf("%d index(es) to drop", count))
	}

	result := strings.Join(parts, ", ")

	if plan.HasDestructive {
		result += " (includes destructive changes)"
	}

	if plan.RequiresReview {
		result += " (requires manual review)"
	}

	return result
}
