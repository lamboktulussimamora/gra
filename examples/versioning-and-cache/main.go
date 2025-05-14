package main

import (
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/cache"
	"github.com/lamboktulussimamora/gra/middleware"
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

// Sample data
var productsV1 = []ProductV1{
	{ID: "1", Name: "Product 1", Price: 100},
	{ID: "2", Name: "Product 2", Price: 200},
	{ID: "3", Name: "Product 3", Price: 300},
}

var productsV2 = []ProductV2{
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

func main() {
	// Create a new GRA application
	r := gra.New()

	// Set up API versioning
	v := versioning.New().
		WithSupportedVersions("1", "2").
		WithDefaultVersion("1")

	// Set up caching with a 30-second TTL for demonstration purposes
	cacheConfig := cache.DefaultCacheConfig()
	cacheConfig.TTL = 30 * time.Second

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

	// Start the server
	gra.Run(":8080", r)
}

// health is a simple health check endpoint without versioning
func health(c *gra.Context) {
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
		c.Error(http.StatusInternalServerError, "API version not found")
		return
	}

	// Different response based on version
	switch versionInfo.Version {
	case "1":
		c.Success(http.StatusOK, "Products retrieved successfully", productsV1)
	case "2":
		c.Success(http.StatusOK, "Products retrieved successfully", productsV2)
	default:
		c.Error(http.StatusBadRequest, "Unsupported API version")
	}
}

// getProduct returns a specific product based on API version
func getProduct(c *gra.Context) {
	// Get product ID from path parameters
	id := c.GetParam("id")
	if id == "" {
		c.Error(http.StatusBadRequest, "Product ID is required")
		return
	}

	// Get API version from context
	versionInfo, exists := versioning.GetAPIVersion(c)
	if !exists {
		c.Error(http.StatusInternalServerError, "API version not found")
		return
	}

	// Different response based on version
	switch versionInfo.Version {
	case "1":
		for _, p := range productsV1 {
			if p.ID == id {
				c.Success(http.StatusOK, "Product retrieved successfully", p)
				return
			}
		}
	case "2":
		for _, p := range productsV2 {
			if p.ID == id {
				c.Success(http.StatusOK, "Product retrieved successfully", p)
				return
			}
		}
	default:
		c.Error(http.StatusBadRequest, "Unsupported API version")
		return
	}

	// Product not found
	c.Error(http.StatusNotFound, "Product not found")
}
