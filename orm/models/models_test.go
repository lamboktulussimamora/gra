package models

import (
	"testing"
	"time"
)

func TestBaseEntity(t *testing.T) {
	t.Run("BaseEntity creation", func(t *testing.T) {
		base := BaseEntity{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if base.ID != 1 {
			t.Errorf("Expected ID 1, got %d", base.ID)
		}
	})

	t.Run("BaseEntity interface methods", func(t *testing.T) {
		base := &BaseEntity{}

		// Test SetID/GetID
		base.SetID(42)
		if base.GetID() != 42 {
			t.Errorf("Expected ID 42, got %d", base.GetID())
		}

		// Test SetCreatedAt/GetCreatedAt
		now := time.Now()
		base.SetCreatedAt(&now)
		if !base.GetCreatedAt().Equal(now) {
			t.Errorf("Expected CreatedAt %v, got %v", now, base.GetCreatedAt())
		}

		// Test SetUpdatedAt/GetUpdatedAt
		base.SetUpdatedAt(&now)
		if !base.GetUpdatedAt().Equal(now) {
			t.Errorf("Expected UpdatedAt %v, got %v", now, base.GetUpdatedAt())
		}

		// Test SetDeletedAt/GetDeletedAt
		base.SetDeletedAt(&now)
		if base.GetDeletedAt() == nil || !base.GetDeletedAt().Equal(now) {
			t.Errorf("Expected DeletedAt %v, got %v", now, base.GetDeletedAt())
		}
	})

	t.Run("BaseEntity soft delete methods", func(t *testing.T) {
		base := &BaseEntity{}

		// Test IsDeleted initially false
		if base.IsDeleted() {
			t.Error("Expected IsDeleted to be false initially")
		}

		// Test SoftDelete
		base.SoftDelete()
		if !base.IsDeleted() {
			t.Error("Expected IsDeleted to be true after SoftDelete")
		}
		if base.DeletedAt == nil {
			t.Error("Expected DeletedAt to be set after SoftDelete")
		}

		// Test Restore
		base.Restore()
		if base.IsDeleted() {
			t.Error("Expected IsDeleted to be false after Restore")
		}
		if base.DeletedAt != nil {
			t.Error("Expected DeletedAt to be nil after Restore")
		}
	})
}

func TestUser(t *testing.T) {
	t.Run("User creation and embedded BaseEntity", func(t *testing.T) {
		user := User{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
			IsActive:  true,
		}

		// Test User-specific fields
		if user.FirstName != "John" {
			t.Errorf("Expected FirstName 'John', got %s", user.FirstName)
		}
		if user.LastName != "Doe" {
			t.Errorf("Expected LastName 'Doe', got %s", user.LastName)
		}
		if user.Email != "john.doe@example.com" {
			t.Errorf("Expected Email 'john.doe@example.com', got %s", user.Email)
		}
		if !user.IsActive {
			t.Error("Expected IsActive to be true")
		}

		// Test embedded BaseEntity access
		if user.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", user.GetID())
		}

		// Test LastLogin pointer field
		now := time.Now()
		user.LastLogin = &now
		if user.LastLogin == nil || !user.LastLogin.Equal(now) {
			t.Errorf("Expected LastLogin %v, got %v", now, user.LastLogin)
		}
	})

	t.Run("User implements IEntity interface", func(t *testing.T) {
		var entity IEntity = &User{}
		entity.SetID(99)
		if entity.GetID() != 99 {
			t.Errorf("Expected ID 99, got %d", entity.GetID())
		}
	})

	t.Run("User TableName", func(t *testing.T) {
		user := User{}
		if user.TableName() != "users" {
			t.Errorf("Expected table name 'users', got %s", user.TableName())
		}
	})
}

func TestProduct(t *testing.T) {
	t.Run("Product creation and embedded BaseEntity", func(t *testing.T) {
		product := Product{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Test Product",
			Description: "A test product",
			Price:       29.99,
			SKU:         "TEST-001",
			CategoryID:  1,
			InStock:     true,
			StockCount:  100,
		}

		// Test Product-specific fields
		if product.Name != "Test Product" {
			t.Errorf("Expected Name 'Test Product', got %s", product.Name)
		}
		if product.Description != "A test product" {
			t.Errorf("Expected Description 'A test product', got %s", product.Description)
		}
		if product.Price != 29.99 {
			t.Errorf("Expected Price 29.99, got %f", product.Price)
		}
		if product.SKU != "TEST-001" {
			t.Errorf("Expected SKU 'TEST-001', got %s", product.SKU)
		}
		if product.CategoryID != 1 {
			t.Errorf("Expected CategoryID 1, got %d", product.CategoryID)
		}
		if !product.InStock {
			t.Error("Expected InStock to be true")
		}
		if product.StockCount != 100 {
			t.Errorf("Expected StockCount 100, got %d", product.StockCount)
		}

		// Test embedded BaseEntity access
		if product.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", product.GetID())
		}
	})

	t.Run("Product implements IEntity interface", func(t *testing.T) {
		var entity IEntity = &Product{}
		entity.SetID(55)
		if entity.GetID() != 55 {
			t.Errorf("Expected ID 55, got %d", entity.GetID())
		}
	})

	t.Run("Product TableName", func(t *testing.T) {
		product := Product{}
		if product.TableName() != "products" {
			t.Errorf("Expected table name 'products', got %s", product.TableName())
		}
	})
}

func TestCategory(t *testing.T) {
	t.Run("Category creation and embedded BaseEntity", func(t *testing.T) {
		category := Category{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Electronics",
			Description: "Electronic products category",
		}

		// Test Category-specific fields
		if category.Name != "Electronics" {
			t.Errorf("Expected Name 'Electronics', got %s", category.Name)
		}
		if category.Description != "Electronic products category" {
			t.Errorf("Expected Description 'Electronic products category', got %s", category.Description)
		}
		if category.ParentID != nil {
			t.Errorf("Expected ParentID to be nil, got %v", category.ParentID)
		}

		// Test embedded BaseEntity access
		if category.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", category.GetID())
		}
	})

	t.Run("Category with parent", func(t *testing.T) {
		parentID := int64(5)
		category := Category{
			BaseEntity: BaseEntity{
				ID: 2,
			},
			Name:     "Smartphones",
			ParentID: &parentID,
		}

		if category.ParentID == nil || *category.ParentID != 5 {
			t.Errorf("Expected ParentID 5, got %v", category.ParentID)
		}
	})

	t.Run("Category implements IEntity interface", func(t *testing.T) {
		var entity IEntity = &Category{}
		entity.SetID(77)
		if entity.GetID() != 77 {
			t.Errorf("Expected ID 77, got %d", entity.GetID())
		}
	})

	t.Run("Category TableName", func(t *testing.T) {
		category := Category{}
		if category.TableName() != "categories" {
			t.Errorf("Expected table name 'categories', got %s", category.TableName())
		}
	})
}

func TestOrder(t *testing.T) {
	t.Run("Order creation and embedded BaseEntity", func(t *testing.T) {
		order := Order{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:      1,
			OrderNumber: "ORD-001",
			Status:      "pending",
			TotalAmount: 199.98,
		}

		// Test Order-specific fields
		if order.UserID != 1 {
			t.Errorf("Expected UserID 1, got %d", order.UserID)
		}
		if order.OrderNumber != "ORD-001" {
			t.Errorf("Expected OrderNumber 'ORD-001', got %s", order.OrderNumber)
		}
		if order.Status != "pending" {
			t.Errorf("Expected Status 'pending', got %s", order.Status)
		}
		if order.TotalAmount != 199.98 {
			t.Errorf("Expected TotalAmount 199.98, got %f", order.TotalAmount)
		}

		// Test embedded BaseEntity access
		if order.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", order.GetID())
		}

		// Test ShippedAt pointer field
		now := time.Now()
		order.ShippedAt = &now
		if order.ShippedAt == nil || !order.ShippedAt.Equal(now) {
			t.Errorf("Expected ShippedAt %v, got %v", now, order.ShippedAt)
		}
	})

	t.Run("Order TableName", func(t *testing.T) {
		order := Order{}
		if order.TableName() != "orders" {
			t.Errorf("Expected table name 'orders', got %s", order.TableName())
		}
	})
}

func TestOrderItem(t *testing.T) {
	t.Run("OrderItem creation and embedded BaseEntity", func(t *testing.T) {
		orderItem := OrderItem{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			OrderID:   1,
			ProductID: 1,
			Quantity:  2,
			UnitPrice: 99.99,
			Total:     199.98,
		}

		// Test OrderItem-specific fields
		if orderItem.OrderID != 1 {
			t.Errorf("Expected OrderID 1, got %d", orderItem.OrderID)
		}
		if orderItem.ProductID != 1 {
			t.Errorf("Expected ProductID 1, got %d", orderItem.ProductID)
		}
		if orderItem.Quantity != 2 {
			t.Errorf("Expected Quantity 2, got %d", orderItem.Quantity)
		}
		if orderItem.UnitPrice != 99.99 {
			t.Errorf("Expected UnitPrice 99.99, got %f", orderItem.UnitPrice)
		}
		if orderItem.Total != 199.98 {
			t.Errorf("Expected Total 199.98, got %f", orderItem.Total)
		}

		// Test embedded BaseEntity access
		if orderItem.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", orderItem.GetID())
		}
	})

	t.Run("OrderItem TableName", func(t *testing.T) {
		orderItem := OrderItem{}
		if orderItem.TableName() != "order_items" {
			t.Errorf("Expected table name 'order_items', got %s", orderItem.TableName())
		}
	})
}

func TestReview(t *testing.T) {
	t.Run("Review creation and embedded BaseEntity", func(t *testing.T) {
		review := Review{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:     1,
			ProductID:  1,
			Rating:     5,
			Title:      "Great product!",
			Comment:    "This product exceeded my expectations.",
			IsVerified: true,
		}

		// Test Review-specific fields
		if review.UserID != 1 {
			t.Errorf("Expected UserID 1, got %d", review.UserID)
		}
		if review.ProductID != 1 {
			t.Errorf("Expected ProductID 1, got %d", review.ProductID)
		}
		if review.Rating != 5 {
			t.Errorf("Expected Rating 5, got %d", review.Rating)
		}
		if review.Title != "Great product!" {
			t.Errorf("Expected Title 'Great product!', got %s", review.Title)
		}
		if review.Comment != "This product exceeded my expectations." {
			t.Errorf("Expected Comment 'This product exceeded my expectations.', got %s", review.Comment)
		}
		if !review.IsVerified {
			t.Error("Expected IsVerified to be true")
		}

		// Test embedded BaseEntity access
		if review.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", review.GetID())
		}
	})

	t.Run("Review TableName", func(t *testing.T) {
		review := Review{}
		if review.TableName() != "reviews" {
			t.Errorf("Expected table name 'reviews', got %s", review.TableName())
		}
	})
}

func TestRole(t *testing.T) {
	t.Run("Role creation and embedded BaseEntity", func(t *testing.T) {
		role := Role{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "admin",
			Description: "Administrator role",
		}

		// Test Role-specific fields
		if role.Name != "admin" {
			t.Errorf("Expected Name 'admin', got %s", role.Name)
		}
		if role.Description != "Administrator role" {
			t.Errorf("Expected Description 'Administrator role', got %s", role.Description)
		}

		// Test embedded BaseEntity access
		if role.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", role.GetID())
		}
	})

	t.Run("Role TableName", func(t *testing.T) {
		role := Role{}
		if role.TableName() != "roles" {
			t.Errorf("Expected table name 'roles', got %s", role.TableName())
		}
	})
}

func TestUserRole(t *testing.T) {
	t.Run("UserRole creation and embedded BaseEntity", func(t *testing.T) {
		userRole := UserRole{
			BaseEntity: BaseEntity{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID: 1,
			RoleID: 1,
		}

		// Test UserRole-specific fields
		if userRole.UserID != 1 {
			t.Errorf("Expected UserID 1, got %d", userRole.UserID)
		}
		if userRole.RoleID != 1 {
			t.Errorf("Expected RoleID 1, got %d", userRole.RoleID)
		}

		// Test embedded BaseEntity access
		if userRole.GetID() != 1 {
			t.Errorf("Expected ID 1, got %d", userRole.GetID())
		}
	})

	t.Run("UserRole TableName", func(t *testing.T) {
		userRole := UserRole{}
		if userRole.TableName() != "user_roles" {
			t.Errorf("Expected table name 'user_roles', got %s", userRole.TableName())
		}
	})
}
