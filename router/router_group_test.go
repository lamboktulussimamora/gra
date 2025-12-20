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

func TestGroupMiddleware_ScopedToGroupRoutes(t *testing.T) {
	r := New()

	order := make([]string, 0, 4)
	globalMW := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			order = append(order, "global")
			next(c)
		}
	}
	groupMW := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			order = append(order, "group")
			next(c)
		}
	}

	r.Use(globalMW)

	api := r.Group("/api")
	api.Use(groupMW)

	api.GET("/hello", func(c *context.Context) { c.Status(http.StatusOK) })
	r.GET("/public", func(c *context.Context) { c.Status(http.StatusOK) })

	// Public route should only execute global middleware.
	order = order[:0]
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/public", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if len(order) != 1 || order[0] != "global" {
		t.Fatalf("expected only global middleware, got %v", order)
	}

	// Group route should execute global then group middleware.
	order = order[:0]
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if len(order) != 2 || order[0] != "global" || order[1] != "group" {
		t.Fatalf("expected middleware order [global group], got %v", order)
	}
}
