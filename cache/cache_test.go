package cache

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/context"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	// Test Set and Get
	entry := &CacheEntry{
		Body:         []byte("test body"),
		StatusCode:   200,
		Headers:      map[string][]string{"Content-Type": {"application/json"}},
		LastModified: time.Now(),
	}

	// Set entry with 100ms TTL
	store.Set("test-key", entry, 100*time.Millisecond)

	// Get entry immediately - should exist
	retrievedEntry, exists := store.Get("test-key")
	if !exists {
		t.Error("Expected cache entry to exist")
	}

	if string(retrievedEntry.Body) != string(entry.Body) {
		t.Errorf("Expected body %s, got %s", string(entry.Body), string(retrievedEntry.Body))
	}

	// Wait for entry to expire
	time.Sleep(150 * time.Millisecond)

	// Get entry after expiration - should not exist
	_, exists = store.Get("test-key")
	if exists {
		t.Error("Expected cache entry to be expired")
	}

	// Test Delete
	store.Set("test-key", entry, time.Minute)
	store.Delete("test-key")
	_, exists = store.Get("test-key")
	if exists {
		t.Error("Expected cache entry to be deleted")
	}

	// Test Clear
	store.Set("key1", entry, time.Minute)
	store.Set("key2", entry, time.Minute)
	store.Clear()

	_, exists1 := store.Get("key1")
	_, exists2 := store.Get("key2")
	if exists1 || exists2 {
		t.Error("Expected all cache entries to be cleared")
	}
}

func TestResponseWriter(t *testing.T) {
	// Create original response writer
	w := httptest.NewRecorder()

	// Create wrapped response writer
	rw := NewResponseWriter(w)

	// Test default status
	if rw.Status() != http.StatusOK {
		t.Errorf("Expected default status %d, got %d", http.StatusOK, rw.Status())
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	if rw.Status() != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rw.Status())
	}

	// Test Write
	content := []byte("Hello, world!")
	n, err := rw.Write(content)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(content) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(content), n)
	}

	// Test Body
	body := rw.Body()
	if string(body) != string(content) {
		t.Errorf("Expected body %s, got %s", string(content), string(body))
	}

	// Check that content was written to original writer
	if w.Body.String() != string(content) {
		t.Errorf("Expected original writer to have content %s, got %s", string(content), w.Body.String())
	}
}

func TestCacheMiddleware(t *testing.T) {
	// Create cache store
	store := NewMemoryStore()

	// Create cache config
	config := DefaultCacheConfig()
	config.Store = store
	config.TTL = time.Minute

	// Create test handler
	var handlerCalled int
	handler := func(c *context.Context) {
		handlerCalled++
		c.SetHeader("Content-Type", "application/json")
		c.SetHeader("X-Test", "value")
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": "Hello"})
	}

	// Create middleware
	middleware := WithConfig(config)(handler)

	// First request should miss cache
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	c1 := context.New(w1, req1)

	middleware(c1)

	if handlerCalled != 1 {
		t.Errorf("Expected handler to be called once, got %d", handlerCalled)
	}

	// Check X-Cache header
	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("Expected X-Cache: MISS, got %s", w1.Header().Get("X-Cache"))
	}

	// Check ETag and Last-Modified headers
	if w1.Header().Get("ETag") == "" {
		t.Error("Expected ETag header to be set")
	}
	if w1.Header().Get("Last-Modified") == "" {
		t.Error("Expected Last-Modified header to be set")
	}

	// Second request should hit cache
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	c2 := context.New(w2, req2)

	middleware(c2)

	// Handler should not be called again
	if handlerCalled != 1 {
		t.Errorf("Expected handler to still be called once, got %d", handlerCalled)
	}

	// Check X-Cache header
	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Expected X-Cache: HIT, got %s", w2.Header().Get("X-Cache"))
	}

	// Response body should be the same
	if w1.Body.String() != w2.Body.String() {
		t.Errorf("Expected same response body, got different bodies")
	}

	// Test conditional GET with If-None-Match
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("If-None-Match", w1.Header().Get("ETag"))
	w3 := httptest.NewRecorder()
	c3 := context.New(w3, req3)

	middleware(c3)

	if w3.Code != http.StatusNotModified {
		t.Errorf("Expected status %d, got %d", http.StatusNotModified, w3.Code)
	}
}

func TestSkipCache(t *testing.T) {
	store := NewMemoryStore()
	config := DefaultCacheConfig()
	config.Store = store

	var handlerCalled int
	handler := func(c *context.Context) {
		handlerCalled++
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": "Hello"})
	}

	middleware := WithConfig(config)(handler)

	// Request with Authorization header should skip cache
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("Authorization", "Bearer token")
	w1 := httptest.NewRecorder()
	c1 := context.New(w1, req1)

	middleware(c1)

	if handlerCalled != 1 {
		t.Errorf("Expected handler to be called once, got %d", handlerCalled)
	}

	// Should not be cached, so header should not exist
	if w1.Header().Get("X-Cache") != "" {
		t.Errorf("Expected no X-Cache header, got %s", w1.Header().Get("X-Cache"))
	}

	// Same request should still skip cache and call handler again
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("Authorization", "Bearer token")
	w2 := httptest.NewRecorder()
	c2 := context.New(w2, req2)

	middleware(c2)

	if handlerCalled != 2 {
		t.Errorf("Expected handler to be called twice, got %d", handlerCalled)
	}
}

func TestNonGetMethod(t *testing.T) {
	store := NewMemoryStore()
	config := DefaultCacheConfig()
	config.Store = store

	var handlerCalled int
	handler := func(c *context.Context) {
		handlerCalled++
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": "Hello"})
	}

	middleware := WithConfig(config)(handler)

	// POST request should not be cached
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"data":"value"}`))
	w := httptest.NewRecorder()
	c := context.New(w, req)

	middleware(c)

	if handlerCalled != 1 {
		t.Errorf("Expected handler to be called once, got %d", handlerCalled)
	}

	// Should not be cached
	if w.Header().Get("X-Cache") != "" {
		t.Errorf("Expected no X-Cache header, got %s", w.Header().Get("X-Cache"))
	}
}

func TestClearAndInvalidateCache(t *testing.T) {
	store := NewMemoryStore()

	// Add test entries
	entry := &CacheEntry{
		Body:         []byte("test body"),
		StatusCode:   200,
		Headers:      map[string][]string{"Content-Type": {"application/json"}},
		LastModified: time.Now(),
	}

	store.Set("key1", entry, time.Minute)
	store.Set("key2", entry, time.Minute)

	// Test InvalidateCache
	InvalidateCache(store, "key1")
	_, exists1 := store.Get("key1")
	_, exists2 := store.Get("key2")

	if exists1 {
		t.Error("Expected key1 to be invalidated")
	}
	if !exists2 {
		t.Error("Expected key2 to still exist")
	}

	// Test ClearCache
	ClearCache(store)
	_, exists2 = store.Get("key2")
	if exists2 {
		t.Error("Expected all entries to be cleared")
	}
}

func TestHopByHopHeaders(t *testing.T) {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, header := range hopByHopHeaders {
		if !isHopByHopHeader(header) {
			t.Errorf("Expected %s to be a hop-by-hop header", header)
		}
	}

	endToEndHeaders := []string{
		"Content-Type",
		"User-Agent",
		"Accept",
		"Authorization",
		"X-Custom-Header",
	}

	for _, header := range endToEndHeaders {
		if isHopByHopHeader(header) {
			t.Errorf("Expected %s not to be a hop-by-hop header", header)
		}
	}
}
