package migrations

import (
	"crypto/sha256"
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
	// Get current model snapshots and database schema
	modelSnapshots, dbSchema, err := cd.gatherCurrentState()
	if err != nil {
		return nil, err
	}

	// Compare and generate changes
	changes, err := cd.generateChanges(dbSchema, modelSnapshots)
	if err != nil {
		return nil, err
	}

	// Create and configure migration plan
	plan := cd.createMigrationPlan(changes, modelSnapshots, dbSchema)

	return plan, nil
}

// gatherCurrentState retrieves current model snapshots and database schema
func (cd *ChangeDetector) gatherCurrentState() (map[string]*ModelSnapshot, map[string]*TableSchema, error) {
	modelSnapshots := cd.registry.GetModels()

	dbSchema, err := cd.inspector.GetCurrentSchema()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read database schema: %w", err)
	}

	return modelSnapshots, dbSchema, nil
}

// generateChanges compares schemas and generates migration changes
func (cd *ChangeDetector) generateChanges(dbSchema map[string]*TableSchema, modelSnapshots map[string]*ModelSnapshot) ([]MigrationChange, error) {
	changes, err := cd.inspector.CompareWithModelSnapshot(dbSchema, modelSnapshots)
	if err != nil {
		return nil, fmt.Errorf("failed to compare schemas: %w", err)
	}
	return changes, nil
}

// createMigrationPlan creates and configures a migration plan
func (cd *ChangeDetector) createMigrationPlan(changes []MigrationChange, modelSnapshots map[string]*ModelSnapshot, dbSchema map[string]*TableSchema) *MigrationPlan {
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

	return plan
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
	hasher := sha256.New()

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

	oldColumn, newColumn, ok := cd.extractColumnInfoFromChange(change)
	if !ok {
		return false
	}

	return cd.hasDataLosingColumnChanges(oldColumn, newColumn)
}

// extractColumnInfoFromChange extracts old and new column info from a change
func (cd *ChangeDetector) extractColumnInfoFromChange(change MigrationChange) (*DatabaseColumnInfo, *ColumnInfo, bool) {
	oldColumn, okOld := change.OldValue.(*DatabaseColumnInfo)
	newColumn, okNew := change.NewValue.(*ColumnInfo)

	if !okOld || !okNew {
		return nil, nil, false
	}

	return oldColumn, newColumn, true
}

// hasDataLosingColumnChanges checks for potentially data-losing changes
func (cd *ChangeDetector) hasDataLosingColumnChanges(oldColumn *DatabaseColumnInfo, newColumn *ColumnInfo) bool {
	// Check for potentially data-losing changes
	if cd.isNullabilityChangeDataLosing(oldColumn, newColumn) {
		return true
	}

	if cd.isLengthReductionDataLosing(oldColumn, newColumn) {
		return true
	}

	if cd.isIncompatibleTypeChange(oldColumn.DataType, newColumn.DataType) {
		return true
	}

	return false
}

// isNullabilityChangeDataLosing checks if nullability change might lose data
func (cd *ChangeDetector) isNullabilityChangeDataLosing(oldColumn *DatabaseColumnInfo, newColumn *ColumnInfo) bool {
	// Making column non-nullable when it was nullable
	return oldColumn.IsNullable && !newColumn.IsNullable
}

// isLengthReductionDataLosing checks if length reduction might lose data
func (cd *ChangeDetector) isLengthReductionDataLosing(oldColumn *DatabaseColumnInfo, newColumn *ColumnInfo) bool {
	// Reducing string length
	if oldColumn.MaxLength != nil && newColumn.MaxLength != nil {
		return *newColumn.MaxLength < *oldColumn.MaxLength
	}
	return false
}

// isIncompatibleTypeChange checks if a type change is incompatible
func (cd *ChangeDetector) isIncompatibleTypeChange(oldType, newType string) bool {
	oldType = strings.ToUpper(strings.TrimSpace(oldType))
	newType = strings.ToUpper(strings.TrimSpace(newType))

	// Define incompatible type changes
	incompatibleChanges := cd.getIncompatibleTypeMap()

	return cd.checkTypeIncompatibility(oldType, newType, incompatibleChanges)
}

// getIncompatibleTypeMap returns a map of incompatible type changes
func (cd *ChangeDetector) getIncompatibleTypeMap() map[string][]string {
	return map[string][]string{
		"TEXT":      {"INTEGER", "BIGINT", "BOOLEAN", "TIMESTAMP", "DATE"},
		"VARCHAR":   {"INTEGER", "BIGINT", "BOOLEAN", "TIMESTAMP", "DATE"},
		"INTEGER":   {"BOOLEAN", "TIMESTAMP", "DATE"},
		"BIGINT":    {"BOOLEAN", "TIMESTAMP", "DATE"},
		"BOOLEAN":   {"INTEGER", "BIGINT", "TEXT", "VARCHAR", "TIMESTAMP", "DATE"},
		"TIMESTAMP": {"INTEGER", "BIGINT", "BOOLEAN"},
		"DATE":      {"INTEGER", "BIGINT", "BOOLEAN"},
	}
}

// checkTypeIncompatibility checks if the new type is incompatible with the old type
func (cd *ChangeDetector) checkTypeIncompatibility(oldType, newType string, incompatibleChanges map[string][]string) bool {
	incompatibleTypes, exists := incompatibleChanges[oldType]
	if !exists {
		return false
	}

	for _, incompatible := range incompatibleTypes {
		if strings.HasPrefix(newType, incompatible) {
			return true
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
	warnings := make([]string, 0, len(plan.Changes))

	// Perform all validation checks
	errors = cd.performValidationChecks(plan.Changes, errors)
	warnings = cd.collectValidationWarnings(plan.Changes, warnings)

	plan.Warnings = warnings
	plan.Errors = errors

	if len(errors) > 0 {
		return fmt.Errorf("migration plan validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// performValidationChecks runs all error-level validation checks
func (cd *ChangeDetector) performValidationChecks(changes []MigrationChange, errors []string) []string {
	// Check for circular dependencies
	if err := cd.checkCircularDependencies(changes); err != nil {
		errors = append(errors, fmt.Sprintf("Circular dependency detected: %v", err))
	}
	return errors
}

// collectValidationWarnings gathers all warning-level issues
func (cd *ChangeDetector) collectValidationWarnings(changes []MigrationChange, warnings []string) []string {
	// Check for orphaned foreign keys
	orphanedFKs := cd.findOrphanedForeignKeys(changes)
	for _, fk := range orphanedFKs {
		warnings = append(warnings, fmt.Sprintf("Foreign key %s references table that will be dropped", fk))
	}

	// Check for data loss potential
	dataLossChanges := cd.findDataLossChanges(changes)
	for _, change := range dataLossChanges {
		warnings = append(warnings, fmt.Sprintf("Potential data loss in %s.%s", change.TableName, change.ColumnName))
	}

	return warnings
}

// checkCircularDependencies checks for circular dependencies in migration changes
func (cd *ChangeDetector) checkCircularDependencies(changes []MigrationChange) error {
	// Build dependency graph
	dependencies := cd.buildDependencyGraph(changes)

	// Check for cycles using DFS
	return cd.detectCyclesInDependencyGraph(dependencies)
}

// buildDependencyGraph creates a dependency graph from migration changes
func (cd *ChangeDetector) buildDependencyGraph(changes []MigrationChange) map[string][]string {
	dependencies := make(map[string][]string)

	for _, change := range changes {
		if change.Type == CreateTable {
			cd.addTableDependencies(change, dependencies)
		}
	}

	return dependencies
}

// addTableDependencies adds foreign key dependencies for a table
func (cd *ChangeDetector) addTableDependencies(change MigrationChange, dependencies map[string][]string) {
	snapshot, ok := change.NewValue.(*ModelSnapshot)
	if !ok {
		return
	}

	for _, constraint := range snapshot.Constraints {
		if constraint.Type == foreignKeyConstraintType && constraint.ReferencedTable != "" {
			dependencies[snapshot.TableName] = append(dependencies[snapshot.TableName], constraint.ReferencedTable)
		}
	}
}

// detectCyclesInDependencyGraph uses DFS to detect circular dependencies
func (cd *ChangeDetector) detectCyclesInDependencyGraph(dependencies map[string][]string) error {
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
	// Preallocate with a reasonable guess (number of changes)
	orphaned := make([]string, 0, len(changes))

	// Find tables being dropped
	droppedTables := cd.identifyDroppedTables(changes)

	// Check for foreign keys referencing dropped tables
	orphaned = cd.findForeignKeysReferencingDroppedTables(changes, droppedTables, orphaned)

	return orphaned
}

// identifyDroppedTables creates a set of tables being dropped
func (cd *ChangeDetector) identifyDroppedTables(changes []MigrationChange) map[string]bool {
	droppedTables := make(map[string]bool)
	for _, change := range changes {
		if change.Type == DropTable {
			droppedTables[change.TableName] = true
		}
	}
	return droppedTables
}

// findForeignKeysReferencingDroppedTables checks changes for foreign keys referencing dropped tables
func (cd *ChangeDetector) findForeignKeysReferencingDroppedTables(changes []MigrationChange, droppedTables map[string]bool, orphaned []string) []string {
	for _, change := range changes {
		if cd.isChangeWithConstraints(change) {
			orphaned = cd.checkConstraintsForOrphanedForeignKeys(change, droppedTables, orphaned)
		}
	}
	return orphaned
}

// isChangeWithConstraints checks if a change type can have constraints
func (cd *ChangeDetector) isChangeWithConstraints(change MigrationChange) bool {
	return change.Type == CreateTable || change.Type == AddColumn
}

// checkConstraintsForOrphanedForeignKeys examines constraints in a change for orphaned foreign keys
func (cd *ChangeDetector) checkConstraintsForOrphanedForeignKeys(change MigrationChange, droppedTables map[string]bool, orphaned []string) []string {
	constraints := cd.extractConstraintsFromChange(change)

	for constraintName, constraint := range constraints {
		if cd.isForeignKeyReferencingDroppedTable(constraint, droppedTables) {
			orphaned = append(orphaned, constraintName)
		}
	}

	return orphaned
}

// extractConstraintsFromChange gets constraints from a migration change
func (cd *ChangeDetector) extractConstraintsFromChange(change MigrationChange) map[string]*ConstraintInfo {
	if snapshot, ok := change.NewValue.(*ModelSnapshot); ok {
		return snapshot.Constraints
	}

	if column, ok := change.NewValue.(*ColumnInfo); ok && len(column.Constraints) > 0 {
		return column.Constraints
	}

	return nil
}

// isForeignKeyReferencingDroppedTable checks if a constraint is a foreign key referencing a dropped table
func (cd *ChangeDetector) isForeignKeyReferencingDroppedTable(constraint *ConstraintInfo, droppedTables map[string]bool) bool {
	return constraint.Type == foreignKeyConstraintType && droppedTables[constraint.ReferencedTable]
}

// findDataLossChanges identifies changes that might cause data loss
func (cd *ChangeDetector) findDataLossChanges(changes []MigrationChange) []MigrationChange {
	dataLossChanges := make([]MigrationChange, 0, len(changes))

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
