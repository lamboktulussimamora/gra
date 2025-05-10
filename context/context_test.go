package context

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	c := New(w, r)

	if c == nil {
		t.Fatal("New() returned nil")
	}

	if c.Writer != w {
		t.Error("Context Writer not set correctly")
	}

	if c.Request != r {
		t.Error("Context Request not set correctly")
	}

	if c.Params == nil {
		t.Error("Context Params not initialized")
	}
}

func TestStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	c.Status(http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestJSONBasic(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := TestData{
		Name: "John",
		Age:  30,
	}

	c.JSON(http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result TestData
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf("Expected %v, got %v", data, result)
	}
}

func TestJSONSpecialChars(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	data := map[string]string{
		"special": "special\"chars'<>&",
		"unicode": "你好世界",
	}

	c.JSON(http.StatusOK, data)

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["special"] != data["special"] || result["unicode"] != data["unicode"] {
		t.Errorf("Expected %v, got %v", data, result)
	}
}

func TestJSONComplex(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type Person struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Address Address  `json:"address"`
		Hobbies []string `json:"hobbies"`
	}

	data := Person{
		Name: "Alice",
		Age:  28,
		Address: Address{
			Street:  "123 Main St",
			City:    "Anytown",
			Country: "USA",
		},
		Hobbies: []string{"reading", "hiking", "coding"},
	}

	c.JSON(http.StatusOK, data)

	var result Person
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf("Names or ages don't match. Expected %v, got %v", data, result)
	}

	if result.Address.Street != data.Address.Street ||
		result.Address.City != data.Address.City ||
		result.Address.Country != data.Address.Country {
		t.Errorf("Addresses don't match. Expected %v, got %v", data.Address, result.Address)
	}

	if len(result.Hobbies) != len(data.Hobbies) {
		t.Errorf("Hobbies length mismatch. Expected %v, got %v", data.Hobbies, result.Hobbies)
	}
}

func TestBindJSON(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30}`
		r := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()
		c := New(w, r)

		type TestData struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var data TestData
		err := c.BindJSON(&data)

		if err != nil {
			t.Fatalf("BindJSON returned error: %v", err)
		}

		if data.Name != "John" {
			t.Errorf("Expected name John, got %s", data.Name)
		}

		if data.Age != 30 {
			t.Errorf("Expected age 30, got %d", data.Age)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{"name":"John","age":}`
		r := httptest.NewRequest("POST", "/test", strings.NewReader(invalidJSON))
		w := httptest.NewRecorder()
		c := New(w, r)

		type TestData struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var data TestData
		err := c.BindJSON(&data)

		if err == nil {
			t.Fatal("BindJSON should have returned an error for invalid JSON")
		}
	})

	t.Run("Empty Body", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		c := New(w, r)

		type TestData struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var data TestData
		err := c.BindJSON(&data)

		if err == nil {
			t.Fatal("BindJSON should have returned an error for empty body")
		}
	})
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	message := "Success message"
	data := map[string]string{"key": "value"}

	c.Success(http.StatusOK, message, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status success, got %s", response.Status)
	}

	if response.Message != message {
		t.Errorf("Expected message %s, got %s", message, response.Message)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to convert response data")
	}

	if val, ok := responseData["key"]; !ok || val != "value" {
		t.Errorf("Expected data {key: value}, got %v", responseData)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	errorMsg := "Error message"

	c.Error(http.StatusBadRequest, errorMsg)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status error, got %s", response.Status)
	}

	if response.Error != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, response.Error)
	}
}

func TestGetParam(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	c.Params = map[string]string{"id": "123"}

	param := c.GetParam("id")
	if param != "123" {
		t.Errorf("Expected param value 123, got %s", param)
	}

	// Test non-existent param
	param = c.GetParam("unknown")
	if param != "" {
		t.Errorf("Expected empty string for non-existent param, got %s", param)
	}
}

func TestGetQuery(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	c := New(w, r)

	query := c.GetQuery("name")
	if query != "John" {
		t.Errorf("Expected query value John, got %s", query)
	}

	// Test non-existent query
	query = c.GetQuery("unknown")
	if query != "" {
		t.Errorf("Expected empty string for non-existent query, got %s", query)
	}
}

func TestWithValue(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	type key struct{}
	keyInstance := key{}
	value := "test value"

	c.WithValue(keyInstance, value)

	if c.Value(keyInstance) != value {
		t.Errorf("Expected value %s, got %v", value, c.Value(keyInstance))
	}
}

func TestValue(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	type key struct{}
	keyInstance := key{}
	value := "test value"

	c.WithValue(keyInstance, value)

	retrievedValue := c.Value(keyInstance)
	if retrievedValue != value {
		t.Errorf("Expected value %s, got %v", value, retrievedValue)
	}

	// Test non-existent key
	type unknownKey struct{}
	unknownKeyInstance := unknownKey{}

	retrievedValue = c.Value(unknownKeyInstance)
	if retrievedValue != nil {
		t.Errorf("Expected nil for non-existent key, got %v", retrievedValue)
	}
}
