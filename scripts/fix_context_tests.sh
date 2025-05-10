#!/bin/bash
# This script updates the context_test.go file with proper constants

cat > /Users/lamboktulussimamora/Projects/gra/context/context_test.go.new << 'EOF'
// Package context provides the Context type for handling HTTP requests and responses.
package context

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test error message constants
const (
	errStatusCode        = "Expected status code %d, got %d"
	errContentType       = "Expected Content-Type application/json, got %s"
	errUnmarshalResponse = "Failed to unmarshal response: %v"
	errResponseValue     = "Expected %v, got %v"
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
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
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
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(errContentType, contentType)
	}

	var result TestData
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf(errResponseValue, data, result)
	}
}

func TestJSONData(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Test with struct
	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := TestData{
		Name: "Jane",
		Age:  25,
	}

	c.JSONData(http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(errContentType, contentType)
	}

	var result TestData
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf(errResponseValue, data, result)
	}

	// Test with map
	w = httptest.NewRecorder()
	c = New(w, r)
	
	mapData := map[string]interface{}{
		"items": []string{"item1", "item2"},
		"count": 2,
	}

	c.JSONData(http.StatusCreated, mapData)

	if w.Code != http.StatusCreated {
		t.Errorf(errStatusCode, http.StatusCreated, w.Code)
	}

	var mapResult map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &mapResult); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	// Check items array
	items, ok := mapResult["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be an array")
	}
	
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	// Check count
	count, ok := mapResult["count"].(float64) // JSON numbers are float64 in Go
	if !ok {
		t.Fatalf("Expected count to be a number")
	}
	
	if count != 2 {
		t.Errorf("Expected count 2, got %v", count)
	}
	
	// Ensure no APIResponse wrapper
	_, hasStatus := mapResult["status"]
	if hasStatus {
		t.Error("Response should not contain status field (should not be wrapped in APIResponse)")
	}
}

func TestBindJSON(t *testing.T) {
	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	jsonData := `{"name":"Alice","age":25}`
	r := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
	w := httptest.NewRecorder()
	c := New(w, r)

	var data TestData
	if err := c.BindJSON(&data); err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	if data.Name != "Alice" || data.Age != 25 {
		t.Errorf(errResponseValue, "Alice/25", data)
	}
}

func TestBindJSONError(t *testing.T) {
	// Invalid JSON
	jsonData := `{"name":"Alice",age:25}`
	r := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))
	w := httptest.NewRecorder()
	c := New(w, r)

	var data map[string]interface{}
	err := c.BindJSON(&data)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	data := map[string]string{"key": "value"}
	message := "Success message"

	c.Success(http.StatusOK, message, data)

	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(errContentType, contentType)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, response.Message)
	}

	// Convert response.Data to map for comparison
	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map, got %T", response.Data)
	}

	value, exists := dataMap["key"]
	if !exists || value != "value" {
		t.Errorf("Expected data to have key 'key' with value 'value', got %v", dataMap)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	errorMsg := "Error message"

	c.Error(http.StatusBadRequest, errorMsg)

	if w.Code != http.StatusBadRequest {
		t.Errorf(errStatusCode, http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(errContentType, contentType)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if response.Error != errorMsg {
		t.Errorf("Expected error '%s', got '%s'", errorMsg, response.Error)
	}
}

func TestGetParam(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Set a parameter
	c.Params["id"] = "123"

	value := c.GetParam("id")
	if value != "123" {
		t.Errorf("Expected parameter 'id' to be '123', got '%s'", value)
	}

	// Non-existent parameter should return empty string
	value = c.GetParam("unknown")
	if value != "" {
		t.Errorf("Expected non-existent parameter to return empty string, got '%s'", value)
	}
}

func TestGetQuery(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	c := New(w, r)

	value := c.GetQuery("name")
	if value != "John" {
		t.Errorf("Expected query parameter 'name' to be 'John', got '%s'", value)
	}

	value = c.GetQuery("age")
	if value != "30" {
		t.Errorf("Expected query parameter 'age' to be '30', got '%s'", value)
	}

	// Non-existent parameter should return empty string
	value = c.GetQuery("unknown")
	if value != "" {
		t.Errorf("Expected non-existent query parameter to return empty string, got '%s'", value)
	}
}

func TestContextValues(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Test setting and getting a value
	key := "user_id"
	value := "12345"

	c.WithValue(key, value)
	result := c.Value(key)

	if result != value {
		t.Errorf("Expected context value '%s' for key '%s', got '%v'", value, key, result)
	}

	// Test getting a non-existent value
	result = c.Value("non_existent")
	if result != nil {
		t.Errorf("Expected nil for non-existent context key, got '%v'", result)
	}
}
EOF

# Move the new file to replace the old one
mv /Users/lamboktulussimamora/Projects/gra/context/context_test.go.new /Users/lamboktulussimamora/Projects/gra/context/context_test.go

echo "Successfully updated context_test.go with constants"
