package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

const (
	// Route paths for benchmarking
	pathUserWithID = "/api/users/:id"
	pathResource   = "/api/resource/"
)

func BenchmarkRouterSimple(b *testing.B) {
	r := New()

	handler := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
	}

	r.GET("/", handler)
	r.GET("/api/users", handler)
	r.POST("/api/users", handler)
	r.GET(pathUserWithID, handler)
	r.PUT(pathUserWithID, handler)
	r.DELETE(pathUserWithID, handler)

	// Simple route
	req1 := httptest.NewRequest("GET", "/", nil)

	// Route with parameter
	req2 := httptest.NewRequest("GET", "/api/users/123", nil)

	b.ResetTimer()
	b.Run("SimpleRoute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req1)
		}
	})

	b.Run("ParameterizedRoute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req2)
		}
	})
}

func BenchmarkRouterComplex(b *testing.B) {
	r := New()

	handler := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
	}

	// Register a larger number of routes to test routing performance with many routes
	for i := 0; i < 100; i++ {
		r.GET(pathResource+string(rune(i)), handler)
		r.GET(pathResource+string(rune(i))+"/:id", handler)
		r.PUT(pathResource+string(rune(i))+"/:id", handler)
		r.DELETE(pathResource+string(rune(i))+"/:id", handler)
	}

	// Add some more specific routes
	r.GET(pathUserWithID+"/profile", handler)
	r.GET(pathUserWithID+"/posts/:postID/comments", handler)
	r.GET(pathUserWithID+"/posts/:postID/comments/:commentID", handler)

	// Requests to test
	req1 := httptest.NewRequest("GET", pathResource+"A", nil)
	req2 := httptest.NewRequest("GET", pathResource+"Z/123", nil)
	req3 := httptest.NewRequest("GET", "/api/users/123/posts/456/comments/789", nil)

	b.ResetTimer()
	b.Run("ManyRoutes_Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req1)
		}
	})

	b.Run("ManyRoutes_WithParameter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req2)
		}
	})

	b.Run("DeepNestedParameters", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req3)
		}
	})
}
