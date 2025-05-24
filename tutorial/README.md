# Tutorial: Building a RESTful API with GRA

This tutorial will guide you through building a complete RESTful API using the GRA framework. We'll cover all the essential concepts, including routing, middleware, validation, error handling, and more.

## What We'll Build

We'll create a simple API for managing a collection of books. It will support:

- Listing all books
- Getting a single book by ID
- Adding a new book
- Updating an existing book
- Deleting a book

## Prerequisites

- Go 1.18 or later
- Basic knowledge of Go and RESTful API concepts

## Step 1: Set Up Your Project

Create a new directory for your project and initialize a Go module:

```bash
mkdir books-api
cd books-api
go mod init books-api
```

Install the GRA framework:

```bash
go get github.com/lamboktulussimamora/gra
```

## Step 2: Create Your Data Model

Create a file named `model.go` with the following content:

```go
package main

import "time"

// Book represents a book in our collection
type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title" validate:"required"`
	Author      string    `json:"author" validate:"required"`
	ISBN        string    `json:"isbn" validate:"required,isbn"`
	Pages       int       `json:"pages" validate:"required,gt=0"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BookRequest represents the incoming request data when creating or updating a book
type BookRequest struct {
	Title       string    `json:"title" validate:"required"`
	Author      string    `json:"author" validate:"required"`
	ISBN        string    `json:"isbn" validate:"required,isbn"`
	Pages       int       `json:"pages" validate:"required,gt=0"`
	PublishedAt time.Time `json:"published_at"`
}
```

## Step 3: Create an In-Memory Store

For simplicity, we'll use an in-memory store to hold our book data. Create a file named `store.go`:

```go
package main

import (
	"sync"
	"time"
)

// BookStore is an in-memory store for books
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

// NewBookStore creates a new BookStore
func NewBookStore() *BookStore {
	return &BookStore{
		books:  make(map[int]Book),
		nextID: 1,
	}
}

// GetBooks returns all books
func (s *BookStore) GetBooks() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	books := make([]Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	return books
}

// GetBook returns a book by ID
func (s *BookStore) GetBook(id int) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	book, exists := s.books[id]
	return book, exists
}

// AddBook adds a new book
func (s *BookStore) AddBook(req BookRequest) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	book := Book{
		ID:          s.nextID,
		Title:       req.Title,
		Author:      req.Author,
		ISBN:        req.ISBN,
		Pages:       req.Pages,
		PublishedAt: req.PublishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	s.books[s.nextID] = book
	s.nextID++
	
	return book
}

// UpdateBook updates an existing book
func (s *BookStore) UpdateBook(id int, req BookRequest) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	book, exists := s.books[id]
	if !exists {
		return Book{}, false
	}
	
	book.Title = req.Title
	book.Author = req.Author
	book.ISBN = req.ISBN
	book.Pages = req.Pages
	book.PublishedAt = req.PublishedAt
	book.UpdatedAt = time.Now()
	
	s.books[id] = book
	return book, true
}

// DeleteBook deletes a book by ID
func (s *BookStore) DeleteBook(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.books[id]; !exists {
		return false
	}
	
	delete(s.books, id)
	return true
}
```

## Step 4: Create API Handlers

Create a file named `handlers.go`:

```go
package main

import (
	"net/http"
	"strconv"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/validator"
)

// Handlers contains all the book API handlers
type Handlers struct {
	store     *BookStore
	validator *validator.Validator
}

// NewHandlers creates a new Handlers instance
func NewHandlers(store *BookStore) *Handlers {
	return &Handlers{
		store:     store,
		validator: validator.New(),
	}
}

// RegisterRoutes registers all the book API routes
func (h *Handlers) RegisterRoutes(r *gra.Router) {
	// Create a books group
	books := r.Group("/books")
	
	// Register routes
	books.GET("/", h.listBooks)
	books.GET("/:id", h.getBook)
	books.POST("/", h.createBook)
	books.PUT("/:id", h.updateBook)
	books.DELETE("/:id", h.deleteBook)
}

// listBooks returns all books
func (h *Handlers) listBooks(c *gra.Context) {
	books := h.store.GetBooks()
	c.Success(http.StatusOK, "Books retrieved successfully", books)
}

// getBook returns a book by ID
func (h *Handlers) getBook(c *gra.Context) {
	id, err := strconv.Atoi(c.GetParam("id"))
	if err != nil {
		c.Error(http.StatusBadRequest, "Invalid book ID")
		return
	}
	
	book, exists := h.store.GetBook(id)
	if !exists {
		c.Error(http.StatusNotFound, "Book not found")
		return
	}
	
	c.Success(http.StatusOK, "Book retrieved successfully", book)
}

// createBook creates a new book
func (h *Handlers) createBook(c *gra.Context) {
	var req BookRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate request
	if err := h.validator.Validate(req); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		c.Error(http.StatusBadRequest, "Validation error", validationErrors...)
		return
	}
	
	// Add book to store
	book := h.store.AddBook(req)
	
	c.Success(http.StatusCreated, "Book created successfully", book)
}

// updateBook updates an existing book
func (h *Handlers) updateBook(c *gra.Context) {
	id, err := strconv.Atoi(c.GetParam("id"))
	if err != nil {
		c.Error(http.StatusBadRequest, "Invalid book ID")
		return
	}
	
	var req BookRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate request
	if err := h.validator.Validate(req); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		c.Error(http.StatusBadRequest, "Validation error", validationErrors...)
		return
	}
	
	// Update book in store
	book, exists := h.store.UpdateBook(id, req)
	if !exists {
		c.Error(http.StatusNotFound, "Book not found")
		return
	}
	
	c.Success(http.StatusOK, "Book updated successfully", book)
}

// deleteBook deletes a book
func (h *Handlers) deleteBook(c *gra.Context) {
	id, err := strconv.Atoi(c.GetParam("id"))
	if err != nil {
		c.Error(http.StatusBadRequest, "Invalid book ID")
		return
	}
	
	// Delete book from store
	if !h.store.DeleteBook(id) {
		c.Error(http.StatusNotFound, "Book not found")
		return
	}
	
	c.Success(http.StatusOK, "Book deleted successfully", nil)
}
```

## Step 5: Create Main Application

Create a file named `main.go`:

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
)

func main() {
	// Create a new router
	r := gra.New()
	
	// Apply middleware
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS("*"),
		middleware.SecureHeaders(),
	)
	
	// Create book store and handlers
	store := NewBookStore()
	handlers := NewHandlers(store)
	
	// Register routes
	handlers.RegisterRoutes(r)
	
	// Add home route
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to the Books API", map[string]interface{}{
			"version":      "1.0",
			"documentation": "/docs",
		})
	})
	
	// Add health check
	r.GET("/health", func(c *gra.Context) {
		c.Success(http.StatusOK, "Service is healthy", nil)
	})
	
	// Start server
	fmt.Println("Starting server on :8080")
	if err := gra.Run(":8080", r); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
```

## Step 6: Run Your API

Run the API with the following command:

```bash
go run .
```

Your API should now be running at `http://localhost:8080`.

## Step 7: Test Your API

You can test your API using curl or a tool like Postman:

### List all books

```bash
curl http://localhost:8080/books
```

### Create a new book

```bash
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Go Programming Language",
    "author": "Alan A. A. Donovan & Brian W. Kernighan",
    "isbn": "978-0134190440",
    "pages": 380,
    "published_at": "2015-10-30T00:00:00Z"
  }'
```

### Get a book by ID

```bash
curl http://localhost:8080/books/1
```

### Update a book

```bash
curl -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Go Programming Language",
    "author": "Alan A. A. Donovan & Brian W. Kernighan",
    "isbn": "978-0134190440",
    "pages": 400,
    "published_at": "2015-11-01T00:00:00Z"
  }'
```

### Delete a book

```bash
curl -X DELETE http://localhost:8080/books/1
```

## Conclusion

Congratulations! You've built a complete RESTful API using the GRA framework. You've learned how to:

- Set up routes and route groups
- Handle HTTP requests
- Validate request data
- Return standardized JSON responses
- Implement CRUD operations
- Apply middleware for logging, recovery, CORS, and secure headers

To expand on this example, you could:

- Add authentication using the JWT middleware
- Implement pagination for the book list
- Add filtering and sorting options
- Replace the in-memory store with a database
- Add API versioning

For more advanced usage, check out the [API Reference](../api-reference/) and [Examples](../examples/) sections.
