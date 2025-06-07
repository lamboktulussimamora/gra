// Package testutils provides common utilities for testing
package testutils

import (
	"database/sql"
	"testing"
)

// CloseDB safely closes a database connection and logs any errors
func CloseDB(t *testing.T, db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}
}

// CloseDBAnyway safely closes a database connection ignoring errors (for defer use)
func CloseDBAnyway(db *sql.DB) {
	if db != nil {
		_ = db.Close()
	}
}
