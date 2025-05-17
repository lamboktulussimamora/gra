// Package cache provides HTTP response caching capabilities.
package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Body         []byte              // The response body
	StatusCode   int                 // The HTTP status code
	Headers      map[string][]string // The HTTP headers
	Expiration   time.Time           // When this entry expires
	LastModified time.Time           // When this entry was last modified
	ETag         string              // Entity Tag for this response
}

// CacheStore defines the interface for cache storage backends
type CacheStore interface {
	// Get retrieves a cached response by key
	Get(key string) (*CacheEntry, bool)
	// Set stores a response in the cache with a key
	Set(key string, entry *CacheEntry, ttl time.Duration)
	// Delete removes an entry from the cache
	Delete(key string)
	// Clear removes all entries from the cache
	Clear()
}

// MemoryStore is an in-memory implementation of CacheStore
type MemoryStore struct {
	items map[string]*CacheEntry
	mutex sync.RWMutex
}

// NewMemoryStore creates a new memory cache store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]*CacheEntry),
	}
}

// Get retrieves an entry from the memory cache
func (s *MemoryStore) Get(key string) (*CacheEntry, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.items[key]
	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.Expiration) {
		delete(s.items, key)
		return nil, false
	}

	return entry, true
}

// Set stores an entry in the memory cache
func (s *MemoryStore) Set(key string, entry *CacheEntry, ttl time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Set expiration time
	entry.Expiration = time.Now().Add(ttl)

	// Generate ETag if not set
	if entry.ETag == "" {
		hash := sha256.Sum256(entry.Body)
		entry.ETag = hex.EncodeToString(hash[:])
	}

	s.items[key] = entry
}

// Delete removes an entry from the memory cache
func (s *MemoryStore) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.items, key)
}

// Clear removes all entries from the memory cache
func (s *MemoryStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = make(map[string]*CacheEntry)
}

// ResponseWriter is a wrapper for http.ResponseWriter that captures the response
type ResponseWriter struct {
	writer    http.ResponseWriter
	body      *bytes.Buffer
	status    int
	headerSet bool
	written   bool
}

// NewResponseWriter creates a new response writer wrapper
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		writer: w,
		body:   &bytes.Buffer{},
		status: http.StatusOK,
	}
}

// Header returns the header map to set before writing a response
func (w *ResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// WriteHeader sends the HTTP status code
func (w *ResponseWriter) WriteHeader(status int) {
	w.status = status
	w.headerSet = true
}

// Write writes the data to the response
func (w *ResponseWriter) Write(b []byte) (int, error) {
	if !w.headerSet {
		w.WriteHeader(http.StatusOK)
	}

	if !w.written {
		w.writer.WriteHeader(w.status)
		w.written = true
	}

	w.body.Write(b)
	return w.writer.Write(b)
}

// Status returns the HTTP status code
func (w *ResponseWriter) Status() int {
	return w.status
}

// Body returns the response body as a byte slice
func (w *ResponseWriter) Body() []byte {
	return w.body.Bytes()
}

// CacheConfig holds configuration options for the cache middleware
type CacheConfig struct {
	// TTL is the default time-to-live for cached items
	TTL time.Duration
	// Methods are the HTTP methods to cache (default: only GET)
	Methods []string
	// Store is the cache store to use
	Store CacheStore
	// KeyGenerator generates cache keys from the request
	KeyGenerator func(*context.Context) string
	// SkipCache determines whether to skip caching for a request
	SkipCache func(*context.Context) bool
	// MaxBodySize is the maximum size of the body to cache (default: 1MB)
	MaxBodySize int64
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:     time.Minute * 5,
		Methods: []string{http.MethodGet},
		Store:   NewMemoryStore(),
		KeyGenerator: func(c *context.Context) string {
			return c.Request.Method + ":" + c.Request.URL.String()
		},
		SkipCache: func(c *context.Context) bool {
			// Skip caching if the request includes Authorization header
			return c.GetHeader("Authorization") != ""
		},
		MaxBodySize: 1024 * 1024, // 1MB
	}
}

// New creates a new cache middleware with default configuration
func New() router.Middleware {
	return WithConfig(DefaultCacheConfig())
}

// initializeConfig sets default values for any unspecified options in the config
func initializeConfig(config *CacheConfig) {
	if config.TTL == 0 {
		config.TTL = DefaultCacheConfig().TTL
	}
	if len(config.Methods) == 0 {
		config.Methods = DefaultCacheConfig().Methods
	}
	if config.Store == nil {
		config.Store = DefaultCacheConfig().Store
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultCacheConfig().KeyGenerator
	}
	if config.SkipCache == nil {
		config.SkipCache = DefaultCacheConfig().SkipCache
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = DefaultCacheConfig().MaxBodySize
	}
}

// isMethodAllowed checks if the HTTP method is allowed for caching
func isMethodAllowed(method string, allowedMethods []string) bool {
	for _, allowed := range allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}

// serveFromCache serves a cached response to the client
func serveFromCache(c *context.Context, entry *CacheEntry) {
	// Serve headers from cache
	for name, values := range entry.Headers {
		for _, value := range values {
			c.SetHeader(name, value)
		}
	}

	// Add cache headers
	c.SetHeader("X-Cache", "HIT")
	c.SetHeader("Age", strconv.FormatInt(int64(time.Since(entry.LastModified).Seconds()), 10))

	// Write status and body
	c.Status(entry.StatusCode)
	w := c.Writer
	if _, err := w.Write(entry.Body); err != nil {
		log.Printf("Error writing cached response: %v", err)
	}
}

// handleConditionalGET checks for conditional GET headers and returns true if 304 Not Modified was sent
func handleConditionalGET(c *context.Context, entry *CacheEntry) bool {
	// Check for conditional GET requests
	ifNoneMatch := c.GetHeader("If-None-Match")
	ifModifiedSince := c.GetHeader("If-Modified-Since")

	// Compare ETag values properly, handling quotes
	if ifNoneMatch != "" {
		// Clean the If-None-Match header to handle quoted ETags
		cleanETag := strings.Trim(ifNoneMatch, "\"")
		entryETag := strings.Trim(entry.ETag, "\"")

		if cleanETag == entryETag {
			c.Status(http.StatusNotModified)
			return true
		}
	}

	if ifModifiedSince != "" {
		if parsedTime, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
			if !entry.LastModified.After(parsedTime) {
				c.Status(http.StatusNotModified)
				return true
			}
		}
	}

	return false
}

// createCacheEntry creates a new cache entry from the response
func createCacheEntry(responseWriter *ResponseWriter, now time.Time) (*CacheEntry, string) {
	headers := make(map[string][]string)

	// Copy headers that should be cached
	for name, values := range responseWriter.Header() {
		// Skip hop-by-hop headers
		if isHopByHopHeader(name) {
			continue
		}
		headers[name] = values
	}

	// Generate ETag
	body := responseWriter.Body()
	hash := sha256.Sum256(body)
	etag := hex.EncodeToString(hash[:])

	entry := &CacheEntry{
		Body:         body,
		StatusCode:   responseWriter.Status(),
		Headers:      headers,
		LastModified: now,
		ETag:         etag,
	}

	return entry, etag
}

// WithConfig creates a new cache middleware with custom configuration
func WithConfig(config CacheConfig) router.Middleware {
	// Initialize configuration with defaults
	initializeConfig(&config)

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Skip cache if the method is not cacheable or if SkipCache returns true
			if !isMethodAllowed(c.Request.Method, config.Methods) || config.SkipCache(c) {
				next(c)
				return
			}

			// Generate cache key
			key := config.KeyGenerator(c)

			// Check if we have a cached response
			if entry, found := config.Store.Get(key); found {
				// Check for conditional GET requests that may result in 304 Not Modified
				if handleConditionalGET(c, entry) {
					return
				}

				// Serve the cached response
				serveFromCache(c, entry)
				return
			}

			// Cache miss, capture the response
			responseWriter := NewResponseWriter(c.Writer)
			c.Writer = responseWriter

			// Call the next handler
			next(c)

			// Don't cache errors or oversized responses
			if responseWriter.Status() >= 400 || int64(len(responseWriter.Body())) > config.MaxBodySize {
				return
			}

			// Create cache entry
			now := time.Now()
			entry, etag := createCacheEntry(responseWriter, now)

			// Add cache headers to response
			c.SetHeader("ETag", etag)
			c.SetHeader("Last-Modified", now.Format(http.TimeFormat))
			c.SetHeader("Cache-Control", fmt.Sprintf("max-age=%d, public", int(config.TTL.Seconds())))
			c.SetHeader("X-Cache", "MISS")

			// Store in cache
			config.Store.Set(key, entry, config.TTL)
		}
	}
}

// isHopByHopHeader determines if the header is a hop-by-hop header
// These headers should not be stored in the cache
func isHopByHopHeader(header string) bool {
	h := strings.ToLower(header)
	switch h {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailers", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

// ClearCache clears the entire cache
func ClearCache(store CacheStore) {
	store.Clear()
}

// InvalidateCache invalidates a specific cache entry
func InvalidateCache(store CacheStore, key string) {
	store.Delete(key)
}
