// Example: Entity Framework Core-like Migration Lifecycle
// This demonstrates the complete migration lifecycle using GRA's EF migration system
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("sqlite3", "./test_migrations/example.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database connection: %v", closeErr)
		}
	}()

	// Create EF Migration Manager
	config := migrations.DefaultEFMigrationConfig()
	config.Logger = log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags)
	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize migration schema (like EF Core's initial setup)
	if err := manager.EnsureSchema(); err != nil {
		log.Fatal("Failed to initialize migration schema:", err)
	}

	// ========================================
	// EF CORE MIGRATION LIFECYCLE DEMONSTRATION
	// ========================================

	fmt.Println("\n🚀 MIGRATION LIFECYCLE DEMO")
	fmt.Println("=====================================")

	// 1. ADD-MIGRATION: Create initial migration
	fmt.Println("\n1️⃣  ADDING INITIAL MIGRATION (Add-Migration CreateUsersTable)")
	createUsersSQL := `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_users_email ON users(email);
	`

	dropUsersSQL := `
	DROP INDEX IF EXISTS idx_users_email;
	DROP TABLE IF EXISTS users;
	`

	migration1 := manager.AddMigration(
		"CreateUsersTable",
		"Initial migration to create users table",
		createUsersSQL,
		dropUsersSQL,
	)

	// 2. ADD-MIGRATION: Add another migration
	fmt.Println("\n2️⃣  ADDING SECOND MIGRATION (Add-Migration AddUserProfiles)")
	createProfilesSQL := `
	CREATE TABLE user_profiles (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		bio TEXT,
		avatar_url VARCHAR(500),
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_profiles_user_id ON user_profiles(user_id);
	`

	dropProfilesSQL := `
	DROP INDEX IF EXISTS idx_profiles_user_id;
	DROP TABLE IF EXISTS user_profiles;
	`

	_ = manager.AddMigration(
		"AddUserProfiles",
		"Add user profiles table with foreign key to users",
		createProfilesSQL,
		dropProfilesSQL,
	)

	// 3. GET-MIGRATION: View migration history before applying
	fmt.Println("\n3️⃣  CHECKING MIGRATION STATUS (Get-Migration)")
	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("Failed to get migration history:", err)
	}

	printMigrationStatus(history)

	// 4. UPDATE-DATABASE: Apply all pending migrations
	fmt.Println("\n4️⃣  APPLYING MIGRATIONS (Update-Database)")
	if err := manager.UpdateDatabase(); err != nil {
		log.Fatal("Failed to update database:", err)
	}

	// 5. GET-MIGRATION: View status after applying
	fmt.Println("\n5️⃣  CHECKING STATUS AFTER UPDATE (Get-Migration)")
	history, err = manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("Failed to get migration history:", err)
	}

	printMigrationStatus(history)

	// 6. ADD-MIGRATION: Add another migration
	fmt.Println("\n6️⃣  ADDING THIRD MIGRATION (Add-Migration AddUserSettings)")
	createSettingsSQL := `
	CREATE TABLE user_settings (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		setting_key VARCHAR(100) NOT NULL,
		setting_value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, setting_key)
	);
	CREATE INDEX idx_settings_user_key ON user_settings(user_id, setting_key);
	`

	dropSettingsSQL := `
	DROP INDEX IF EXISTS idx_settings_user_key;
	DROP TABLE IF EXISTS user_settings;
	`

	migration3 := manager.AddMigration(
		"AddUserSettings",
		"Add user settings table for user preferences",
		createSettingsSQL,
		dropSettingsSQL,
	)

	// 7. UPDATE-DATABASE: Apply specific migration
	fmt.Println("\n7️⃣  APPLYING SPECIFIC MIGRATION (Update-Database AddUserSettings)")
	if err := manager.UpdateDatabase(migration3.ID); err != nil {
		log.Fatal("Failed to update database to specific migration:", err)
	}

	// 8. ROLLBACK: Demonstrate rollback functionality
	fmt.Println("\n8️⃣  ROLLING BACK MIGRATION (Update-Database CreateUsersTable)")
	if err := manager.RollbackMigration(migration1.ID); err != nil {
		log.Fatal("Failed to rollback migration:", err)
	}

	// 9. FINAL STATUS: Check final state
	fmt.Println("\n9️⃣  FINAL MIGRATION STATUS")
	history, err = manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("Failed to get final migration history:", err)
	}

	printMigrationStatus(history)

	// 10. AUTOMATIC MIGRATION: Generate migration from entity
	fmt.Println("\n🔟 AUTOMATIC MIGRATION GENERATION")
	demonstrateAutoMigration(manager)

	fmt.Println("\n✅ MIGRATION LIFECYCLE DEMO COMPLETED!")
	fmt.Println("=====================================")
}

// printMigrationStatus displays the current migration status
func printMigrationStatus(history *migrations.MigrationHistory) {
	fmt.Printf("📊 Migration Status:\n")
	fmt.Printf("   Applied: %d migrations\n", len(history.Applied))
	fmt.Printf("   Pending: %d migrations\n", len(history.Pending))
	fmt.Printf("   Failed:  %d migrations\n", len(history.Failed))

	if len(history.Applied) > 0 {
		fmt.Println("\n   ✅ Applied Migrations:")
		for _, m := range history.Applied {
			fmt.Printf("      • %s (%s) - %s\n", m.ID, m.AppliedAt.Format("2006-01-02 15:04:05"), m.Description)
		}
	}

	if len(history.Pending) > 0 {
		fmt.Println("\n   ⏳ Pending Migrations:")
		for _, m := range history.Pending {
			fmt.Printf("      • %s - %s\n", m.ID, m.Description)
		}
	}

	if len(history.Failed) > 0 {
		fmt.Println("\n   ❌ Failed Migrations:")
		for _, m := range history.Failed {
			fmt.Printf("      • %s - %s\n", m.ID, m.Description)
		}
	}
}

// User entity for automatic migration demo
type User struct {
	ID        int    `db:"id" migrations:"primary_key,auto_increment"`
	Email     string `db:"email" migrations:"unique,not_null,type:varchar(255)"`
	Name      string `db:"name" migrations:"not_null,type:varchar(100)"`
	Age       int    `db:"age" migrations:"null,type:integer"`
	IsActive  bool   `db:"is_active" migrations:"default:true"`
	CreatedAt string `db:"created_at" migrations:"default:CURRENT_TIMESTAMP,type:timestamp"`
}

// demonstrateAutoMigration shows automatic migration generation from entities
func demonstrateAutoMigration(manager *migrations.EFMigrationManager) {
	user := User{}

	fmt.Println("🤖 Generating migration from User entity...")

	// Use the available CreateAutoMigrations method
	entities := []interface{}{user}
	err := manager.CreateAutoMigrations(entities, "AutoGenerateUserEntity")
	if err != nil {
		log.Printf("Failed to generate auto migration: %v", err)
		return
	}

	fmt.Printf("✅ Generated auto migration for User entity\n")

	// Apply the auto-generated migration
	fmt.Println("🚀 Applying auto-generated migration...")
	if err := manager.UpdateDatabase(); err != nil {
		log.Printf("Failed to apply auto migration: %v", err)
	} else {
		fmt.Println("✅ Auto-generated migration applied successfully!")
	}
}
