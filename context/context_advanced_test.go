package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextComprehensiveScenarios(t *testing.T) {
	// Test header manipulation with multiple values
	t.Run("ComplexHeaderManipulation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer token123")
		req.Header.Set("User-Agent", "TestClient/1.0")
		req.Header.Add("X-Custom", "value1")
		req.Header.Add("X-Custom", "value2")

		w := httptest.NewRecorder()
		c := New(w, req)

		// Test setting multiple response headers
		c.SetHeader("Cache-Control", "no-cache").
			SetHeader("Access-Control-Allow-Origin", "*").
			SetHeader("X-API-Version", "v1")

		// Verify request headers
		auth := c.GetHeader("Authorization")
		if auth != "Bearer token123" {
			t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", auth)
		}

		userAgent := c.GetHeader("User-Agent")
		if userAgent != "TestClient/1.0" {
			t.Errorf("Expected User-Agent header 'TestClient/1.0', got '%s'", userAgent)
		}

		// Note: GetHeader only returns first value for multi-value headers
		customHeader := c.GetHeader("X-Custom")
		if customHeader != "value1" {
			t.Errorf("Expected X-Custom header 'value1', got '%s'", customHeader)
		}

		// Test content type helper
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		contentType := c.GetContentType()
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", contentType)
		}

		// Verify response headers were set
		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Error("Cache-Control header not set correctly")
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("CORS header not set correctly")
		}
		if w.Header().Get("X-API-Version") != "v1" {
			t.Error("API version header not set correctly")
		}
	})

	// Test parameter handling with complex scenarios
	t.Run("ParameterHandling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Simulate route parameters (normally set by router)
		c.Params["id"] = "123"
		c.Params["category"] = "electronics"

		// Test parameter retrieval
		userID := c.GetParam("id")
		if userID != "123" {
			t.Errorf("Expected user ID '123', got '%s'", userID)
		}

		category := c.GetParam("category")
		if category != "electronics" {
			t.Errorf("Expected category 'electronics', got '%s'", category)
		}

		// Test non-existent parameter
		nonExistent := c.GetParam("non-existent")
		if nonExistent != "" {
			t.Errorf("Expected empty string for non-existent param, got '%s'", nonExistent)
		}
	})

	// Test query parameter handling
	t.Run("QueryParameterHandling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/search?page=2&limit=10&sort=name&filter=active&filter=verified", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test basic query parameter retrieval
		page := c.GetQuery("page")
		if page != "2" {
			t.Errorf("Expected page '2', got '%s'", page)
		}

		limit := c.GetQuery("limit")
		if limit != "10" {
			t.Errorf("Expected limit '10', got '%s'", limit)
		}

		// Test non-existent query parameter
		nonExistent := c.GetQuery("non-existent")
		if nonExistent != "" {
			t.Errorf("Expected empty string for non-existent query param, got '%s'", nonExistent)
		}

		// Test sorting parameter
		sort := c.GetQuery("sort")
		if sort != "name" {
			t.Errorf("Expected sort 'name', got '%s'", sort)
		}

		// Note: GetQuery only returns first value for multi-value params
		filter := c.GetQuery("filter")
		if filter != "active" {
			t.Errorf("Expected first filter 'active', got '%s'", filter)
		}
	})

	// Test request body binding with various scenarios
	t.Run("RequestBodyBinding", func(t *testing.T) {
		// Test successful JSON binding
		testData := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
			"tags":  []string{"developer", "go"},
		}
		jsonData, _ := json.Marshal(testData)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := New(w, req)

		var result map[string]interface{}
		err := c.BindJSON(&result)
		if err != nil {
			t.Errorf("Expected no error binding JSON, got: %v", err)
		}

		if result["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", result["name"])
		}

		if result["email"] != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%v'", result["email"])
		}

		// Verify numeric conversion
		if age, ok := result["age"].(float64); !ok || age != 30 {
			t.Errorf("Expected age 30, got %v (type: %T)", result["age"], result["age"])
		}

		// Test binding to struct
		type User struct {
			Name  string   `json:"name"`
			Email string   `json:"email"`
			Age   int      `json:"age"`
			Tags  []string `json:"tags"`
		}

		req2 := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req2.Header.Set("Content-Type", "application/json")
		c2 := New(httptest.NewRecorder(), req2)

		var user User
		err = c2.BindJSON(&user)
		if err != nil {
			t.Errorf("Expected no error binding to struct, got: %v", err)
		}

		if user.Name != "John Doe" {
			t.Errorf("Expected struct name 'John Doe', got '%s'", user.Name)
		}

		if len(user.Tags) != 2 || user.Tags[0] != "developer" {
			t.Errorf("Expected tags to be parsed correctly, got %v", user.Tags)
		}
	})

	// Test error responses
	t.Run("ErrorResponses", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test error response
		c.Error(400, "Invalid request data")

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Error unmarshaling response: %v", err)
		}

		if response.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", response.Status)
		}

		if response.Error != "Invalid request data" {
			t.Errorf("Expected error message 'Invalid request data', got '%s'", response.Error)
		}
	})

	// Test success responses
	t.Run("SuccessResponses", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test success response with data
		data := map[string]interface{}{
			"id":   1,
			"name": "Test Item",
		}
		c.Success(200, "Item retrieved successfully", data)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Error unmarshaling response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		if response.Message != "Item retrieved successfully" {
			t.Errorf("Expected message 'Item retrieved successfully', got '%s'", response.Message)
		}

		// Verify data structure
		if responseData, ok := response.Data.(map[string]interface{}); ok {
			if responseData["name"] != "Test Item" {
				t.Errorf("Expected data name 'Test Item', got '%v'", responseData["name"])
			}
		} else {
			t.Error("Expected data to be a map")
		}
	})

	// Test JSONData response (direct data without APIResponse wrapper)
	t.Run("JSONDataResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test direct JSON data response
		data := []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		}
		c.JSONData(200, data)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		// Should not be wrapped in APIResponse
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Error unmarshaling direct JSON data: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 items in response, got %d", len(response))
		}

		if response[0]["name"] != "Item 1" {
			t.Errorf("Expected first item name 'Item 1', got '%v'", response[0]["name"])
		}
	})

	// Test cookie handling
	t.Run("CookieHandling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
		req.AddCookie(&http.Cookie{Name: "user_pref", Value: "theme=dark"})

		w := httptest.NewRecorder()
		c := New(w, req)

		// Test getting existing cookie
		sessionID, err := c.GetCookie("session_id")
		if err != nil {
			t.Errorf("Expected no error getting cookie, got: %v", err)
		}
		if sessionID != "abc123" {
			t.Errorf("Expected session ID 'abc123', got '%s'", sessionID)
		}

		// Test getting non-existent cookie
		_, err = c.GetCookie("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent cookie")
		}

		// Test setting cookie
		c.SetCookie("new_cookie", "value123", 3600, "/", "", false, true)

		// Verify cookie was set in response
		cookies := w.Header()["Set-Cookie"]
		found := false
		for _, cookie := range cookies {
			if strings.Contains(cookie, "new_cookie=value123") {
				found = true
				if !strings.Contains(cookie, "HttpOnly") {
					t.Error("Expected HttpOnly flag to be set")
				}
				if !strings.Contains(cookie, "Max-Age=3600") {
					t.Error("Expected Max-Age to be set")
				}
				break
			}
		}
		if !found {
			t.Error("Expected new_cookie to be set in response")
		}
	})

	// Test redirect functionality
	t.Run("RedirectHandling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/old-path", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test redirect
		c.Redirect(302, "/new-path")

		if w.Code != 302 {
			t.Errorf("Expected status code 302, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/new-path" {
			t.Errorf("Expected Location header '/new-path', got '%s'", location)
		}
	})

	// Test context value storage and retrieval using WithValue/Value
	t.Run("ContextValueStorage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test setting and getting context values using WithValue/Value
		c = c.WithValue("user_id", "123")
		c = c.WithValue("role", "admin")

		userID := c.Value("user_id")
		if userID != "123" {
			t.Errorf("Expected user_id '123', got '%v'", userID)
		}

		role := c.Value("role")
		if role != "admin" {
			t.Errorf("Expected role 'admin', got '%v'", role)
		}

		// Test getting non-existent value
		nonExistent := c.Value("non-existent")
		if nonExistent != nil {
			t.Errorf("Expected nil for non-existent value, got '%v'", nonExistent)
		}
	})

	// Test status setting and chaining
	t.Run("StatusChaining", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, req)

		// Test status chaining
		result := c.Status(201).SetHeader("X-Custom", "test")
		if result != c {
			t.Error("Expected Status() to return the same context for chaining")
		}

		// Note: we can't easily test if status was set since it's written to ResponseWriter
		// and we can't inspect it without triggering the write
	})
}

// Test binding with invalid JSON and error conditions
func TestContextBindingErrors(t *testing.T) {
	t.Run("InvalidJSONBinding", func(t *testing.T) {
		invalidJSON := `{"name": "John", "age": invalid}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := New(w, req)

		var result map[string]interface{}
		err := c.BindJSON(&result)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("EmptyBodyBinding", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := New(w, req)

		var result map[string]interface{}
		err := c.BindJSON(&result)
		if err == nil {
			t.Error("Expected error for empty JSON body")
		}
	})

	t.Run("ReadErrorBinding", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", &errorReader{err: fmt.Errorf("read error")})
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := New(w, req)

		var result map[string]interface{}
		err := c.BindJSON(&result)
		if err == nil {
			t.Error("Expected error when body read fails")
		}
	})
}
