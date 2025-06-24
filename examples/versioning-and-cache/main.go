package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/cache"
	"github.com/lamboktulussimamora/gra/middleware"
	"github.com/lamboktulussimamora/gra/router"
	"github.com/lamboktulussimamora/gra/versioning"
)

// ProductV1 represents a product in API v1
type ProductV1 struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// ProductV2 adds additional fields for API v2
type ProductV2 struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description"` // Added in v2
	Categories  []string `json:"categories"`  // Added in v2
	CreatedAt   string   `json:"created_at"`  // Added in v2
}

// GetSampleDataV1 returns sample data for API v1
func GetSampleDataV1() []ProductV1 {
	return []ProductV1{
		{ID: "1", Name: "Product 1", Price: 100},
		{ID: "2", Name: "Product 2", Price: 200},
		{ID: "3", Name: "Product 3", Price: 300},
	}
}

// GetSampleDataV2 returns sample data for API v2
func GetSampleDataV2() []ProductV2 {
	return []ProductV2{
		{
			ID:          "1",
			Name:        "Product 1 Enhanced",
			Price:       100,
			Description: "This is product 1 with enhanced description",
			Categories:  []string{"electronics", "gadgets"},
			CreatedAt:   "2023-01-15T10:00:00Z",
		},
		{
			ID:          "2",
			Name:        "Product 2 Enhanced",
			Price:       200,
			Description: "This is product 2 with enhanced description",
			Categories:  []string{"accessories", "lifestyle"},
			CreatedAt:   "2023-02-20T11:30:00Z",
		},
		{
			ID:          "3",
			Name:        "Product 3 Enhanced",
			Price:       300,
			Description: "This is product 3 with enhanced description",
			Categories:  []string{"home", "kitchen"},
			CreatedAt:   "2023-03-25T09:15:00Z",
		},
	}
}

// Sample data (now using functions)
var productsV1 = GetSampleDataV1()
var productsV2 = GetSampleDataV2()

// AppConfig holds the configuration for the application
type AppConfig struct {
	CacheTTL          time.Duration
	SupportedVersions []string
	DefaultVersion    string
	VersionHeaderName string
}

// DefaultAppConfig returns the default configuration
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		CacheTTL:          30 * time.Second,
		SupportedVersions: []string{"1", "2"},
		DefaultVersion:    "1",
		VersionHeaderName: "API-Version",
	}
}

// SetupRouter creates and configures a new GRA router with versioning and caching
func SetupRouter(config *AppConfig) *router.Router {
	// Create a new GRA application
	r := gra.New()

	// Set up API versioning with header strategy
	v := versioning.New().
		WithStrategy(&versioning.HeaderVersionStrategy{HeaderName: config.VersionHeaderName}).
		WithSupportedVersions(config.SupportedVersions...).
		WithDefaultVersion(config.DefaultVersion)

	// Set up caching
	cacheConfig := cache.DefaultCacheConfig()
	cacheConfig.TTL = config.CacheTTL

	// Add global middleware
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
		v.Middleware(),                // Apply versioning middleware
		cache.WithConfig(cacheConfig), // Apply cache middleware
		middleware.SecureHeaders(),    // Add secure headers
	)

	// Define API routes with versioning
	api := r.Group("/api")
	{
		// Products endpoint
		api.GET("/products", getProducts)
		api.GET("/products/:id", getProduct)

		// Add more routes as needed
		api.GET("/health", health)
	}

	return r
}

// FindProductV1 finds a product by ID in V1 data
func FindProductV1(id string, products []ProductV1) *ProductV1 {
	for _, p := range products {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

// FindProductV2 finds a product by ID in V2 data
func FindProductV2(id string, products []ProductV2) *ProductV2 {
	for _, p := range products {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

func main() {
	fmt.Println("🚀 Starting GRA Versioning and Cache Example")
	fmt.Println("===========================================")

	config := DefaultAppConfig()
	r := SetupRouter(config)

	fmt.Println("✓ API versioning configured (supported: v1, v2, default: v1)")
	fmt.Printf("✓ Cache configured with %v TTL\n", config.CacheTTL)
	fmt.Println("✓ Middleware configured (logger, recovery, versioning, cache, security)")
	fmt.Println("✓ Routes configured:")
	fmt.Println("  - GET /api/products (versioned)")
	fmt.Println("  - GET /api/products/:id (versioned)")
	fmt.Println("  - GET /api/health (health check)")
	fmt.Println()
	fmt.Println("📡 Server starting on http://localhost:8080")
	fmt.Println("🔧 Available API versions: v1 (default), v2")
	fmt.Println("💡 Usage examples:")
	fmt.Println("  - curl http://localhost:8080/api/products")
	fmt.Println("  - curl -H \"API-Version: 2\" http://localhost:8080/api/products")
	fmt.Println("  - curl http://localhost:8080/api/products/1")
	fmt.Println("  - curl http://localhost:8080/api/health")
	fmt.Println()

	// Start the server
	log.Fatal(gra.Run(":8080", r))
}

// health is a simple health check endpoint without versioning
func health(c *gra.Context) {
	log.Println("💚 Health check requested")
	c.Success(http.StatusOK, "Service is healthy", map[string]string{
		"status": "up",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// getProducts returns all products based on API version
func getProducts(c *gra.Context) {
	// Get API version from context
	versionInfo, exists := versioning.GetAPIVersion(c)
	if !exists {
		log.Println("⚠️  API version not found in request context")
		c.Error(http.StatusInternalServerError, "API version not found")
		return
	}

	log.Printf("📋 Fetching products for API version: %s", versionInfo.Version)

	// Different response based on version
	switch versionInfo.Version {
	case "1":
		log.Printf("✓ Returning %d products (v1 format)", len(productsV1))
		c.Success(http.StatusOK, "Products retrieved successfully", productsV1)
	case "2":
		log.Printf("✓ Returning %d products (v2 format with enhanced data)", len(productsV2))
		c.Success(http.StatusOK, "Products retrieved successfully", productsV2)
	default:
		log.Printf("❌ Unsupported API version requested: %s", versionInfo.Version)
		c.Error(http.StatusBadRequest, "Unsupported API version")
	}
}

// getProduct returns a specific product based on API version
func getProduct(c *gra.Context) {
	// Get product ID from path parameters
	id := c.GetParam("id")
	if id == "" {
		log.Println("❌ Product ID missing in request")
		c.Error(http.StatusBadRequest, "Product ID is required")
		return
	}

	// Get API version from context
	versionInfo, exists := versioning.GetAPIVersion(c)
	if !exists {
		log.Println("⚠️  API version not found in request context")
		c.Error(http.StatusInternalServerError, "API version not found")
		return
	}

	log.Printf("🔍 Searching for product ID: %s (API version: %s)", id, versionInfo.Version)

	// Different response based on version
	switch versionInfo.Version {
	case "1":
		if p := FindProductV1(id, productsV1); p != nil {
			log.Printf("✓ Found product (v1): %s", p.Name)
			c.Success(http.StatusOK, "Product retrieved successfully", p)
			return
		}
	case "2":
		if p := FindProductV2(id, productsV2); p != nil {
			log.Printf("✓ Found product (v2): %s", p.Name)
			c.Success(http.StatusOK, "Product retrieved successfully", p)
			return
		}
	default:
		log.Printf("❌ Unsupported API version requested: %s", versionInfo.Version)
		c.Error(http.StatusBadRequest, "Unsupported API version")
		return
	}

	// Product not found
	log.Printf("❌ Product not found: %s", id)
	c.Error(http.StatusNotFound, "Product not found")
}
