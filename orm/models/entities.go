package models

import (
	"time"
)

// User entity represents a system user
type User struct {
	BaseEntity
	FirstName string     `db:"first_name" json:"first_name" validate:"required,min=2,max=50" sql:"size:50;not_null"`
	LastName  string     `db:"last_name" json:"last_name" validate:"required,min=2,max=50" sql:"size:50;not_null"`
	Email     string     `db:"email" json:"email" validate:"required,email" sql:"size:255;not_null;unique"`
	Password  string     `db:"password" json:"-" validate:"required,min=6" sql:"size:255;not_null"`
	IsActive  bool       `db:"is_active" json:"is_active" sql:"default:true"`
	LastLogin *time.Time `db:"last_login" json:"last_login,omitempty" sql:""`

	// Navigation properties (excluded from database)
	Roles   []*Role   `json:"roles,omitempty" sql:"-"`
	Orders  []*Order  `json:"orders,omitempty" sql:"-"`
	Reviews []*Review `json:"reviews,omitempty" sql:"-"`
}

// Product entity represents a product in the catalog
type Product struct {
	BaseEntity
	Name        string  `db:"name" json:"name" validate:"required,min=2,max=200" sql:"size:200;not_null"`
	Description string  `db:"description" json:"description" sql:"type:TEXT"`
	Price       float64 `db:"price" json:"price" validate:"required,min=0" sql:"type:DECIMAL(10,2);not_null"`
	SKU         string  `db:"sku" json:"sku" validate:"required" sql:"size:100;not_null;unique"`
	CategoryID  int64   `db:"category_id" json:"category_id" validate:"required" sql:"foreign_key:categories(id);not_null"`
	InStock     bool    `db:"in_stock" json:"in_stock" sql:"default:true"`
	StockCount  int     `db:"stock_count" json:"stock_count" sql:"default:0"`

	// Navigation properties (excluded from database)
	Category   *Category    `json:"category,omitempty" sql:"-"`
	OrderItems []*OrderItem `json:"order_items,omitempty" sql:"-"`
	Reviews    []*Review    `json:"reviews,omitempty" sql:"-"`
}

// Category entity represents a product category
type Category struct {
	BaseEntity
	Name        string `db:"name" json:"name" validate:"required,min=2,max=100" sql:"size:100;not_null;unique"`
	Description string `db:"description" json:"description" sql:"type:TEXT"`
	ParentID    *int64 `db:"parent_id" json:"parent_id,omitempty" sql:"foreign_key:categories(id)"`

	// Navigation properties (excluded from database)
	Parent   *Category   `json:"parent,omitempty" sql:"-"`
	Children []*Category `json:"children,omitempty" sql:"-"`
	Products []*Product  `json:"products,omitempty" sql:"-"`
}

// Order entity represents a customer order
type Order struct {
	BaseEntity
	UserID      int64      `db:"user_id" json:"user_id" validate:"required" sql:"foreign_key:users(id);not_null"`
	OrderNumber string     `db:"order_number" json:"order_number" validate:"required" sql:"size:50;not_null;unique"`
	Status      string     `db:"status" json:"status" validate:"required" sql:"size:20;not_null;default:'pending'"`
	TotalAmount float64    `db:"total_amount" json:"total_amount" validate:"required,min=0" sql:"type:DECIMAL(10,2);not_null"`
	ShippedAt   *time.Time `db:"shipped_at" json:"shipped_at,omitempty" sql:""`

	// Navigation properties (excluded from database)
	User       *User        `json:"user,omitempty" sql:"-"`
	OrderItems []*OrderItem `json:"order_items,omitempty" sql:"-"`
}

// OrderItem entity represents an item within an order
type OrderItem struct {
	BaseEntity
	OrderID   int64   `db:"order_id" json:"order_id" validate:"required" sql:"foreign_key:orders(id);not_null"`
	ProductID int64   `db:"product_id" json:"product_id" validate:"required" sql:"foreign_key:products(id);not_null"`
	Quantity  int     `db:"quantity" json:"quantity" validate:"required,min=1" sql:"not_null"`
	UnitPrice float64 `db:"unit_price" json:"unit_price" validate:"required,min=0" sql:"type:DECIMAL(10,2);not_null"`
	Total     float64 `db:"total" json:"total" validate:"required,min=0" sql:"type:DECIMAL(10,2);not_null"`

	// Navigation properties (excluded from database)
	Order   *Order   `json:"order,omitempty" sql:"-"`
	Product *Product `json:"product,omitempty" sql:"-"`
}

// Review entity represents a product review
type Review struct {
	BaseEntity
	UserID     int64  `db:"user_id" json:"user_id" validate:"required" sql:"foreign_key:users(id);not_null"`
	ProductID  int64  `db:"product_id" json:"product_id" validate:"required" sql:"foreign_key:products(id);not_null"`
	Rating     int    `db:"rating" json:"rating" validate:"required,min=1,max=5" sql:"not_null"`
	Title      string `db:"title" json:"title" validate:"required,min=5,max=200" sql:"size:200;not_null"`
	Comment    string `db:"comment" json:"comment" validate:"required,min=10" sql:"type:TEXT;not_null"`
	IsVerified bool   `db:"is_verified" json:"is_verified" sql:"default:false"`

	// Navigation properties (excluded from database)
	User    *User    `json:"user,omitempty" sql:"-"`
	Product *Product `json:"product,omitempty" sql:"-"`
}

// Role entity represents a user role
type Role struct {
	BaseEntity
	Name        string `db:"name" json:"name" validate:"required,min=2,max=50" sql:"size:50;not_null;unique"`
	Description string `db:"description" json:"description" sql:"type:TEXT"`

	// Navigation properties (excluded from database)
	Users []*User `json:"users,omitempty" sql:"-"`
}

// UserRole entity represents the many-to-many relationship between users and roles
type UserRole struct {
	BaseEntity
	UserID int64 `db:"user_id" json:"user_id" validate:"required" sql:"foreign_key:users(id);not_null"`
	RoleID int64 `db:"role_id" json:"role_id" validate:"required" sql:"foreign_key:roles(id);not_null"`

	// Navigation properties (excluded from database)
	User *User `json:"user,omitempty" sql:"-"`
	Role *Role `json:"role,omitempty" sql:"-"`
}

// TableName returns the table name for the User entity.
func (User) TableName() string { return "users" }

// TableName returns the table name for the Product entity.
func (Product) TableName() string { return "products" }

// TableName returns the table name for the Category entity.
func (Category) TableName() string { return "categories" }

// TableName returns the table name for the Order entity.
func (Order) TableName() string { return "orders" }

// TableName returns the table name for the OrderItem entity.
func (OrderItem) TableName() string { return "order_items" }

// TableName returns the table name for the Review entity.
func (Review) TableName() string { return "reviews" }

// TableName returns the table name for the Role entity.
func (Role) TableName() string { return "roles" }

// TableName returns the table name for the UserRole entity.
func (UserRole) TableName() string { return "user_roles" }
