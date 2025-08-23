package models

import (
	"testing"
	"time"
)

func TestBaseEntity_GettersSettersAndSoftDelete(t *testing.T) {
	var b BaseEntity

	// ID
	b.SetID(42)
	if b.GetID() != 42 {
		t.Fatalf("ID getter/setter mismatch: got %d", b.GetID())
	}

	// CreatedAt
	now := time.Now().Add(-time.Hour)
	b.SetCreatedAt(&now)
	if !b.GetCreatedAt().Equal(now) {
		t.Fatalf("CreatedAt mismatch")
	}

	// UpdatedAt
	later := time.Now()
	b.SetUpdatedAt(&later)
	if !b.GetUpdatedAt().Equal(later) {
		t.Fatalf("UpdatedAt mismatch")
	}

	// DeletedAt / SoftDelete / Restore
	if b.IsDeleted() {
		t.Fatalf("should not be deleted initially")
	}
	b.SoftDelete()
	if !b.IsDeleted() || b.GetDeletedAt() == nil {
		t.Fatalf("expected soft deleted state")
	}
	b.Restore()
	if b.IsDeleted() || b.GetDeletedAt() != nil {
		t.Fatalf("expected restored state")
	}
}

func TestTableNames(t *testing.T) {
	if (User{}).TableName() != "users" {
		t.Fatalf("users table name")
	}
	if (Product{}).TableName() != "products" {
		t.Fatalf("products table name")
	}
	if (Category{}).TableName() != "categories" {
		t.Fatalf("categories table name")
	}
	if (Order{}).TableName() != "orders" {
		t.Fatalf("orders table name")
	}
	if (OrderItem{}).TableName() != "order_items" {
		t.Fatalf("order_items table name")
	}
	if (Review{}).TableName() != "reviews" {
		t.Fatalf("reviews table name")
	}
	if (Role{}).TableName() != "roles" {
		t.Fatalf("roles table name")
	}
	if (UserRole{}).TableName() != "user_roles" {
		t.Fatalf("user_roles table name")
	}
}
