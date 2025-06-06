// Package models provides entity definitions and base types for the ORM layer.
package models

import (
	"time"
)

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	ID        int64      `db:"id" json:"id" sql:"primary_key;auto_increment"`
	CreatedAt time.Time  `db:"created_at" json:"created_at" sql:"not_null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at" sql:"not_null;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty" sql:"index"`
}

// IEntity defines the interface that all entities must implement
type IEntity interface {
	GetID() int64
	SetID(id int64)
	GetCreatedAt() time.Time
	SetCreatedAt(t *time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(t *time.Time)
	GetDeletedAt() *time.Time
	SetDeletedAt(t *time.Time)
}

// GetID returns the entity's ID
func (b *BaseEntity) GetID() int64 {
	return b.ID
}

// SetID sets the entity's ID
func (b *BaseEntity) SetID(id int64) {
	b.ID = id
}

// GetCreatedAt returns the entity's creation time
func (b *BaseEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// SetCreatedAt sets the entity's creation time
func (b *BaseEntity) SetCreatedAt(t *time.Time) {
	if t != nil {
		b.CreatedAt = *t
	}
}

// GetUpdatedAt returns the entity's last update time
func (b *BaseEntity) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// SetUpdatedAt sets the entity's last update time
func (b *BaseEntity) SetUpdatedAt(t *time.Time) {
	if t != nil {
		b.UpdatedAt = *t
	}
}

// GetDeletedAt returns the entity's deletion time (soft delete)
func (b *BaseEntity) GetDeletedAt() *time.Time {
	return b.DeletedAt
}

// SetDeletedAt sets the entity's deletion time (soft delete)
func (b *BaseEntity) SetDeletedAt(t *time.Time) {
	b.DeletedAt = t
}

// IsDeleted checks if the entity is soft deleted
func (b *BaseEntity) IsDeleted() bool {
	return b.DeletedAt != nil
}

// SoftDelete marks the entity as deleted
func (b *BaseEntity) SoftDelete() {
	now := time.Now()
	b.DeletedAt = &now
}

// Restore removes the soft delete mark
func (b *BaseEntity) Restore() {
	b.DeletedAt = nil
}
