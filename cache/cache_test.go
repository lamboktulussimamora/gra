package cache

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// Test constants for better maintainability
const (
	// Test values
	testBody        = "test body"
	testMessage     = "Hello"
	testKey         = "test-key"
	testKey1        = "key1"
	testKey2        = "key2"
	testBearerToken = "Bearer token"
	testValue       = "value"

	// Header names
	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	headerXTest         = "X-Test"
	headerXCache        = "X-Cache"
	headerETag          = "ETag"
	headerLastModified  = "Last-Modified"
	headerIfNoneMatch   = "If-None-Match"

	// Header values
	valApplicationJSON = "application/json"
	valCacheHit        = "HIT"
	valCacheMiss       = "MISS"

	// Error messages
	errCacheEntryExists = "Expected cache entry to %s"
	errEntryBody        = "Expected body %s, got %s"
	errHandlerCallCount = "Expected handler to be called %s, got %d"
	errStatus           = "Expected status %d, got %d"
	errHeader           = "Expected %s header to be %s, got %s"
	errHeaderPresence   = "Expected %s header to be %s"
	errResponseBodyDiff = "Expected same response body, got different bodies"
	errHopByHopHeader   = "Expected %s %s to be a hop-by-hop header"
	errNoHeader         = "Expected no %s header, got %s"
	errEntriesCleared   = "Expected all cache entries to be cleared"
	errInvalidatedEntry = "Expected %s to be invalidated"
	errEntryStillExists = "Expected %s to still exist"

	// Test durations
	shortTTL           = 100 * time.Millisecond
	expirationWaitTime = 150 * time.Millisecond
	standardTTL        = time.Minute
)

// createTestEntry creates a standard cache entry for testing
func createTestEntry() *CacheEntry {
	return &CacheEntry{
		Body:         []byte(testBody),
		StatusCode:   http.StatusOK,
		Headers:      map[string][]string{headerContentType: {valApplicationJSON}},
		LastModified: time.Now(),
	}
}

// testCacheEntryExists verifies if a cache entry exists or not
func testCacheEntryExists(t *testing.T, store *MemoryStore, key string, shouldExist bool, reason string) {
	t.Helper()
	_, exists := store.Get(key)
	if shouldExist && !exists {
		t.Errorf(errCacheEntryExists, "exist")
	} else if !shouldExist && exists {
		t.Errorf(errCacheEntryExists, reason)
	}
}

func TestMemoryStore(t *testing.T) {
	t.Run("Basic Cache Operations", func(t *testing.T) {
		store := NewMemoryStore()
		entry := createTestEntry()

		t.Run("Set and Get", func(t *testing.T) {
			// Set entry with short TTL
			store.Set(testKey, entry, shortTTL)

			// Get entry immediately - should exist
			retrievedEntry, exists := store.Get(testKey)
			if !exists {
				t.Errorf(errCacheEntryExists, "exist")
			}

			if string(retrievedEntry.Body) != string(entry.Body) {
				t.Errorf(errEntryBody, string(entry.Body), string(retrievedEntry.Body))
			}

			// Wait for entry to expire
			time.Sleep(expirationWaitTime)

			// Get entry after expiration - should not exist
			testCacheEntryExists(t, store, testKey, false, "be expired")
		})

		t.Run("Delete", func(t *testing.T) {
			// Test Delete
			store.Set(testKey, entry, standardTTL)
			store.Delete(testKey)
			testCacheEntryExists(t, store, testKey, false, "be deleted")
		})

		t.Run("Clear", func(t *testing.T) {
			// Test Clear
			store.Set(testKey1, entry, standardTTL)
			store.Set(testKey2, entry, standardTTL)
			store.Clear()

			testCacheEntryExists(t, store, testKey1, false, "be deleted after clear")
			testCacheEntryExists(t, store, testKey2, false, "be deleted after clear")
		})
	})
}

func TestResponseWriter(t *testing.T) {
	const testContent = "Hello, world!"

	// Create original response writer
	w := httptest.NewRecorder()

	// Create wrapped response writer
	rw := NewResponseWriter(w)

	t.Run("Default Status", func(t *testing.T) {
		if rw.Status() != http.StatusOK {
			t.Errorf(errStatus, http.StatusOK, rw.Status())
		}
	})

	t.Run("WriteHeader", func(t *testing.T) {
		rw.WriteHeader(http.StatusCreated)
		if rw.Status() != http.StatusCreated {
			t.Errorf(errStatus, http.StatusCreated, rw.Status())
		}
	})

	t.Run("Write Content", func(t *testing.T) {
		content := []byte(testContent)
		n, err := rw.Write(content)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if n != len(content) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(content), n)
		}

		// Test Body method
		body := rw.Body()
		if string(body) != testContent {
			t.Errorf(errEntryBody, testContent, string(body))
		}

		// Check original writer
		if w.Body.String() != testContent {
			t.Errorf("Expected original writer to have content %s, got %s",
				testContent, w.Body.String())
		}
	})
}

// createTestHandler creates a handler for cache testing that increments a counter when called
func createTestHandler(handlerCalled *int) func(c *context.Context) {
	return func(c *context.Context) {
		(*handlerCalled)++
		c.SetHeader(headerContentType, valApplicationJSON)
		c.SetHeader(headerXTest, testValue)
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": testMessage})
	}
}

// setupCacheTest creates common test objects for cache middleware tests
func setupCacheTest() (*MemoryStore, CacheConfig, *int) {
	store := NewMemoryStore()
	config := DefaultCacheConfig()
	config.Store = store
	config.TTL = time.Minute
	handlerCalled := new(int)
	return store, config, handlerCalled
}

// testCacheHeader checks for the presence of expected cache headers
func testCacheHeader(t *testing.T, w http.ResponseWriter, expectedValue string) {
	t.Helper()
	if w.Header().Get(headerXCache) != expectedValue {
		t.Errorf(errHeader, headerXCache, expectedValue, w.Header().Get(headerXCache))
	}
}

// setupRequest creates and returns an HTTP test request and recorder
func setupRequest(method, path string, headers map[string]string) (*httptest.ResponseRecorder, *context.Context) {
	req := httptest.NewRequest(method, path, nil)
	// Add any headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	return w, context.New(w, req)
}

// testFirstRequest performs the first request in cache tests (always a miss)
func testFirstRequest(t *testing.T, middleware router.HandlerFunc) (*httptest.ResponseRecorder, string) {
	w, c := setupRequest(http.MethodGet, "/test", nil)
	middleware(c)

	// Check ETag and Last-Modified headers
	etag := w.Header().Get(headerETag)
	if etag == "" {
		t.Errorf(errHeaderPresence, headerETag, "set")
		t.Logf("Headers received: %v", w.Header())
		// Return with a default etag to prevent test failure cascade
		etag = "default-etag-for-test"
	}
	if w.Header().Get(headerLastModified) == "" {
		t.Errorf(errHeaderPresence, headerLastModified, "set")
	}

	return w, etag
}

func TestCacheMiddleware(t *testing.T) {
	_, config, handlerCalled := setupCacheTest()
	handler := createTestHandler(handlerCalled)
	middleware := WithConfig(config)(handler)

	t.Run("Cache Miss and Hit", func(t *testing.T) {
		// First request - Cache Miss
		w1, _ := testFirstRequest(t, middleware)

		if *handlerCalled != 1 {
			t.Errorf(errHandlerCallCount, "once", *handlerCalled)
		}
		testCacheHeader(t, w1, valCacheMiss)

		// Second request - Cache Hit
		w2, c2 := setupRequest(http.MethodGet, "/test", nil)
		middleware(c2)

		// Handler should not be called again
		if *handlerCalled != 1 {
			t.Errorf(errHandlerCallCount, "once", *handlerCalled)
		}
		testCacheHeader(t, w2, valCacheHit)
	})

	t.Run("Conditional GET", func(t *testing.T) {
		// Reset store and handler counter to ensure clean state
		store := NewMemoryStore()
		config := DefaultCacheConfig()
		config.Store = store
		handlerCalled := new(int)
		*handlerCalled = 0
		handler := createTestHandler(handlerCalled)
		localMiddleware := WithConfig(config)(handler)

		// Step 1: Make initial request to populate cache
		w1, c1 := setupRequest(http.MethodGet, "/test", nil)
		localMiddleware(c1)

		// Get the ETag from the first response
		etag := w1.Header().Get(headerETag)
		t.Logf("Original ETag from server: %s", etag)

		if etag == "" {
			t.Fatal("No ETag in the first response. Check cache middleware implementation.")
		}

		// Step 2: Make conditional request with If-None-Match header
		// Ensure ETag is properly quoted
		quotedETag := etag
		if !strings.HasPrefix(etag, "\"") && !strings.HasSuffix(etag, "\"") {
			quotedETag = "\"" + etag + "\""
		}

		t.Logf("Using If-None-Match: %s", quotedETag)
		headers := map[string]string{headerIfNoneMatch: quotedETag}
		w2, c2 := setupRequest(http.MethodGet, "/test", headers)

		// Execute the conditional request
		localMiddleware(c2)

		// Verify response is 304 Not Modified
		t.Logf("Response status code: %d", w2.Code)
		t.Logf("Response headers: %v", w2.Header())

		if w2.Code != http.StatusNotModified {
			t.Errorf(errStatus, http.StatusNotModified, w2.Code)
		} else {
			// Test passed, handler shouldn't be called twice
			if *handlerCalled > 1 {
				t.Errorf(errHandlerCallCount, "once (for the first request only)", *handlerCalled)
			}
		}
	})
}

func TestSkipCache(t *testing.T) {
	store := NewMemoryStore()
	config := DefaultCacheConfig()
	config.Store = store

	var handlerCalled int
	handler := func(c *context.Context) {
		handlerCalled++
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": testMessage})
	}

	middleware := WithConfig(config)(handler)

	// Request with Authorization header should skip cache
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set(headerAuthorization, testBearerToken)
	w1 := httptest.NewRecorder()
	c1 := context.New(w1, req1)

	middleware(c1)

	if handlerCalled != 1 {
		t.Errorf(errHandlerCallCount, "once", handlerCalled)
	}

	// Should not be cached, so header should not exist
	if w1.Header().Get(headerXCache) != "" {
		t.Errorf(errNoHeader, headerXCache, w1.Header().Get(headerXCache))
	}

	// Same request should still skip cache and call handler again
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set(headerAuthorization, testBearerToken)
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
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{"message": testMessage})
	}

	middleware := WithConfig(config)(handler)

	// POST request should not be cached
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"data":"value"}`))
	w := httptest.NewRecorder()
	c := context.New(w, req)

	middleware(c)

	if handlerCalled != 1 {
		t.Errorf(errHandlerCallCount, "once", handlerCalled)
	}

	// Should not be cached
	if w.Header().Get(headerXCache) != "" {
		t.Errorf(errNoHeader, headerXCache, w.Header().Get(headerXCache))
	}
}

func TestClearAndInvalidateCache(t *testing.T) {
	store := NewMemoryStore()

	// Add test entries
	entry := createTestEntry()

	store.Set(testKey1, entry, time.Minute)
	store.Set(testKey2, entry, time.Minute)

	// Test InvalidateCache
	t.Run("InvalidateCache", func(t *testing.T) {
		InvalidateCache(store, testKey1)
		_, exists1 := store.Get(testKey1)
		_, exists2 := store.Get(testKey2)

		if exists1 {
			t.Error("Expected key1 to be invalidated")
		}
		if !exists2 {
			t.Error("Expected key2 to still exist")
		}
	})

	// Test ClearCache
	t.Run("ClearCache", func(t *testing.T) {
		ClearCache(store)
		_, exists2 := store.Get(testKey2)
		if exists2 {
			t.Error("Expected all entries to be cleared")
		}
	})
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
			t.Errorf(errHopByHopHeader, header, "")
		}
	}

	endToEndHeaders := []string{
		headerContentType,
		"User-Agent",
		"Accept",
		headerAuthorization,
		"X-Custom-Header",
	}

	for _, header := range endToEndHeaders {
		if isHopByHopHeader(header) {
			t.Errorf(errHopByHopHeader, header, "not")
		}
	}
}
