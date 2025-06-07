package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/lamboktulussimamora/gra/orm/schema"
)

// Test entity
type TestUser struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

func main() {
	userType := reflect.TypeOf(TestUser{})
	idField, found := userType.FieldByName("ID")
	if !found {
		fmt.Println("ID field not found")
		return
	}

	fmt.Println("Field tags:")
	fmt.Printf("  db: %q\n", idField.Tag.Get("db"))
	fmt.Printf("  migration: %q\n", idField.Tag.Get("migration"))

	// Test PostgreSQL
	pgColumn := schema.ParseFieldToColumnForDriver(idField, schema.PostgreSQL)
	fmt.Printf("PostgreSQL column: %q\n", pgColumn)

	// Test SQLite
	sqliteColumn := schema.ParseFieldToColumnForDriver(idField, schema.SQLite)
	fmt.Printf("SQLite column: %q\n", sqliteColumn)

	// Test name field
	nameField, found := userType.FieldByName("Name")
	if found {
		fmt.Printf("\nName field PostgreSQL: %q\n", schema.ParseFieldToColumnForDriver(nameField, schema.PostgreSQL))
	}
}
