package cache

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// TestCacheCustomConfiguration tests advanced cache configuration scenarios
func TestCacheCustomConfiguration(t *testing.T) {
	t.Run("CustomKeyGenerator", func(t *testing.T) {
		store := NewMemoryStore()
		config := DefaultCacheConfig()
		config.Store = store
		config.KeyGenerator = func(c *context.Context) string {
			userID := c.Request.Header.Get("X-User-ID")
			if userID == "" {
				userID = "anonymous"
			}
			return fmt.Sprintf("user:%s:path:%s", userID, c.Request.URL.Path)
		}

		callCount := 0
		testHandler := func(c *context.Context) {
			callCount++
			c.JSON(200, map[string]string{"message": "custom key test"})
		}

		r := router.New()
		r.Use(WithConfig(config))
		r.GET("/test", testHandler)

		// First request with User-ID header
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.Header.Set("X-User-ID", "user123")
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if callCount != 1 {
			t.Errorf("Expected handler to be called once, got %d", callCount)
		}

		// Second request with same user ID - should hit cache
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-User-ID", "user123")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if callCount != 1 {
			t.Errorf("Expected handler to still be called once (cache hit), got %d", callCount)
		}

		// Third request with different user ID - should miss cache
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.Header.Set("X-User-ID", "user456")
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)

		if callCount != 2 {
			t.Errorf("Expected handler to be called twice (cache miss for different user), got %d", callCount)
		}
	})

	t.Run("SkipCacheCondition", func(t *testing.T) {
		store := NewMemoryStore()
		config := DefaultCacheConfig()
		config.Store = store
		config.SkipCache = func(c *context.Context) bool {
			return c.Request.Header.Get("X-No-Cache") == "true"
		}

		callCount := 0
		testHandler := func(c *context.Context) {
			callCount++
			c.JSON(200, map[string]string{"message": "skip test"})
		}

		r := router.New()
		r.Use(WithConfig(config))
		r.GET("/test", testHandler)

		// First request - should cache
		req1 := httptest.NewRequest("GET", "/test", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		// Second request - should hit cache
		req2 := httptest.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if callCount != 1 {
			t.Errorf("Expected handler to be called once (cache hit), got %d", callCount)
		}

		// Third request with no-cache header - should skip cache
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.Header.Set("X-No-Cache", "true")
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)

		if callCount != 2 {
			t.Errorf("Expected handler to be called twice (cache skipped), got %d", callCount)
		}
	})

	t.Run("MaxBodySizeLimit", func(t *testing.T) {
		store := NewMemoryStore()
		config := DefaultCacheConfig()
		config.Store = store
		config.MaxBodySize = 20 // Very small limit

		callCount := 0
		testHandler := func(c *context.Context) {
			callCount++
			// Return a large response that exceeds MaxBodySize
			c.JSON(200, map[string]string{
				"message": "This is a very long message that should exceed the max body size limit for caching in this test case",
			})
		}

		r := router.New()
		r.Use(WithConfig(config))
		r.GET("/test", testHandler)

		// First request - should not cache due to large response
		req1 := httptest.NewRequest("GET", "/test", nil)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		// Second request - should not hit cache (not cached due to size)
		req2 := httptest.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if callCount != 2 {
			t.Errorf("Expected handler to be called twice (not cached due to size), got %d", callCount)
		}
	})
}

// TestCacheStoreAdvanced tests advanced store operations
func TestCacheStoreAdvanced(t *testing.T) {
	t.Run("ExpirationHandling", func(t *testing.T) {
		store := NewMemoryStore()

		// Create entries with different expiration times
		validEntry := &Entry{
			Body:       []byte("valid"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
		}

		expiredEntry := &Entry{
			Body:       []byte("expired"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
		}

		// Set valid entry with long TTL
		store.Set("valid-key", validEntry, time.Hour)

		// Set expired entry with very short TTL
		store.Set("expired-key", expiredEntry, time.Nanosecond)

		// Wait a bit to ensure expiration
		time.Sleep(time.Millisecond)

		// Valid entry should be retrievable
		if _, exists := store.Get("valid-key"); !exists {
			t.Error("Expected valid entry to be retrievable")
		}

		// Expired entry should not be retrievable
		if _, exists := store.Get("expired-key"); exists {
			t.Error("Expected expired entry to not be retrievable")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		store := NewMemoryStore()

		entry := &Entry{
			Body:       []byte("test"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
			Expiration: time.Now().Add(time.Hour),
		}

		done := make(chan bool, 3)

		// Concurrent writes
		go func() {
			for i := 0; i < 50; i++ {
				store.Set(fmt.Sprintf("write-key-%d", i), entry, time.Hour)
			}
			done <- true
		}()

		// Concurrent reads
		go func() {
			for i := 0; i < 50; i++ {
				_, _ = store.Get(fmt.Sprintf("read-key-%d", i))
			}
			done <- true
		}()

		// Concurrent deletes
		go func() {
			for i := 0; i < 25; i++ {
				store.Delete(fmt.Sprintf("delete-key-%d", i))
			}
			done <- true
		}()

		// Wait for completion
		for i := 0; i < 3; i++ {
			<-done
		}

		// Test passes if no race conditions occur
	})

	t.Run("ClearAllEntries", func(t *testing.T) {
		store := NewMemoryStore()

		entry := &Entry{
			Body:       []byte("test"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
			Expiration: time.Now().Add(time.Hour),
		}

		// Add multiple entries
		keys := []string{"key1", "key2", "key3", "key4"}
		for _, key := range keys {
			store.Set(key, entry, time.Hour)
		}

		// Verify entries exist
		for _, key := range keys {
			if _, exists := store.Get(key); !exists {
				t.Errorf("Expected key %s to exist before clear", key)
			}
		}

		// Clear all
		store.Clear()

		// Verify all entries are gone
		for _, key := range keys {
			if _, exists := store.Get(key); exists {
				t.Errorf("Expected key %s to be cleared", key)
			}
		}
	})
}

// TestCacheUtilityFunctions tests utility functions
func TestCacheUtilityFunctions(t *testing.T) {
	t.Run("ClearCacheFunction", func(t *testing.T) {
		store := NewMemoryStore()

		entry := &Entry{
			Body:       []byte("test"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
			Expiration: time.Now().Add(time.Hour),
		}

		store.Set("test-key", entry, time.Hour)

		// Verify entry exists
		if _, exists := store.Get("test-key"); !exists {
			t.Error("Expected entry to exist before clear")
		}

		// Clear using utility function
		ClearCache(store)

		// Verify entry is gone
		if _, exists := store.Get("test-key"); exists {
			t.Error("Expected entry to be cleared")
		}
	})

	t.Run("InvalidateCacheFunction", func(t *testing.T) {
		store := NewMemoryStore()

		entry := &Entry{
			Body:       []byte("test"),
			StatusCode: 200,
			Headers:    make(map[string][]string),
			Expiration: time.Now().Add(time.Hour),
		}

		store.Set("test-key", entry, time.Hour)
		store.Set("other-key", entry, time.Hour)

		// Verify entries exist
		if _, exists := store.Get("test-key"); !exists {
			t.Error("Expected test-key to exist before invalidation")
		}
		if _, exists := store.Get("other-key"); !exists {
			t.Error("Expected other-key to exist before invalidation")
		}

		// Invalidate specific key
		InvalidateCache(store, "test-key")

		// Verify only specific key is gone
		if _, exists := store.Get("test-key"); exists {
			t.Error("Expected test-key to be invalidated")
		}
		if _, exists := store.Get("other-key"); !exists {
			t.Error("Expected other-key to still exist")
		}
	})

	t.Run("HopByHopHeaderDetection", func(t *testing.T) {
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
				t.Errorf("Expected %s to be detected as hop-by-hop header", header)
			}
			// Test case insensitive
			if !isHopByHopHeader(strings.ToLower(header)) {
				t.Errorf("Expected %s (lowercase) to be detected as hop-by-hop header", header)
			}
		}

		// Test non-hop-by-hop headers
		normalHeaders := []string{
			"Content-Type",
			"Cache-Control",
			"ETag",
			"Last-Modified",
			"Authorization",
		}

		for _, header := range normalHeaders {
			if isHopByHopHeader(header) {
				t.Errorf("Expected %s to NOT be detected as hop-by-hop header", header)
			}
		}
	})
}
