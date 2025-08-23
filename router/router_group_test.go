package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

// DB-free tests to cover group prefix normalization and nested grouping

func TestNormalizePrefix(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"/api", "/api"},
		{"api", "/api"},
		{"/api/", "/api"},
		{"v1/", "/v1"},
	}
	for _, c := range cases {
		got := normalizePrefix(c.in)
		if got != c.out {
			t.Fatalf("normalizePrefix(%q)=%q want %q", c.in, got, c.out)
		}
	}
}

func TestGroupAndNestedGroupRoutes(t *testing.T) {
	r := New()
	api := r.Group("/api/")      // should normalize to "/api"
	v1 := api.Group("v1/")       // should normalize and append to "/api/v1"
	items := v1.Group("/items/") // "/api/v1/items"

	called := false
	items.GET("/list", func(c *context.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/list", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !called {
		t.Fatalf("expected 200 with handler called, code=%d called=%v", w.Code, called)
	}
}
