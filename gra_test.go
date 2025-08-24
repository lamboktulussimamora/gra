package gra

import "testing"

// TestNew_CreatesRouter ensures New() constructs a non-nil router
func TestNew_CreatesRouter(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatalf("expected non-nil router from gra.New()")
	}
}
