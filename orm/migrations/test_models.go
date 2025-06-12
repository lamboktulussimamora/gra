package migrations

import (
	"time"
)

// Test models for auto migration
type AutoTestUser struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type AutoTestProduct struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// TableName returns the table name for AutoTestUser
func (AutoTestUser) TableName() string {
	return "auto_test_user"
}

// TableName returns the table name for AutoTestProduct
func (AutoTestProduct) TableName() string {
	return "auto_test_product"
}
