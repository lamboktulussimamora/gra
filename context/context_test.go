package context

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test constants
const (
	// Error message constants
	errStatusCode        = "Expected status code %d, got %d"
	errContentType       = "Expected Content-Type application/json, got %s"
	errUnmarshalResponse = "Failed to unmarshal response: %v"
	errResponseValue     = "Expected %v, got %v"
	errBindJSON          = "BindJSON returned error: %v"
	errExpectedName      = "Expected name %s, got %s"
	errExpectedAge       = "Expected age %d, got %d"
	errExpectedError     = "BindJSON should have returned an error"
	errReadError         = "Expected an error when reading request body fails, got nil"
	errExpectedStatus    = "Expected status %s, got %s"
	errExpectedMessage   = "Expected message %s, got %s"
	errDataConvert       = "Failed to convert response data"
	errExpectedData      = "Expected data %v, got %v"
	errExpectedParam     = "Expected param value %s, got %s"
	errExpectedEmpty     = "Expected empty string for non-existent %s, got %s"
	errExpectedValue     = "Expected value %s, got %v"
	errExpectedNil       = "Expected nil for non-existent key, got %v"
	errExpectedCount     = "Expected count %d, got %v"
	errNoStatus          = "Response should not contain status field (should not be wrapped in APIResponse)"
	errContextNil        = "New() returned nil"
	errContextWriter     = "Context Writer not set correctly"
	errContextRequest    = "Context Request not set correctly"
	errContextParams     = "Context Params not initialized"
	errExpectedItems     = "Expected %d items, got %d"
	errWrongErrorType    = "Expected %v error, got: %v"
	errExpectedArray     = "Expected items to be an array"
	errExpectedNumber    = "Expected count to be a number"

	// HTTP header constants
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

func TestNew(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	c := New(w, r)

	if c == nil {
		t.Fatal(errContextNil)
	}

	if c.Writer != w {
		t.Error(errContextWriter)
	}

	if c.Request != r {
		t.Error(errContextRequest)
	}

	if c.Params == nil {
		t.Error(errContextParams)
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

	contentType := w.Header().Get(headerContentType)
	if contentType != contentTypeJSON {
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
		t.Fatalf(errUnmarshalResponse, err)
	}

	if result["special"] != data["special"] || result["unicode"] != data["unicode"] {
		t.Errorf(errResponseValue, data, result)
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
		t.Fatalf(errUnmarshalResponse, err)
	}

	if result.Name != data.Name || result.Age != data.Age {
		t.Errorf(errResponseValue, data, result)
	}

	if result.Address.Street != data.Address.Street ||
		result.Address.City != data.Address.City ||
		result.Address.Country != data.Address.Country {
		t.Errorf(errResponseValue, data.Address, result.Address)
	}

	if len(result.Hobbies) != len(data.Hobbies) {
		t.Errorf(errResponseValue, data.Hobbies, result.Hobbies)
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
			t.Fatalf(errBindJSON, err)
		}

		if data.Name != "John" {
			t.Errorf(errExpectedName, "John", data.Name)
		}

		if data.Age != 30 {
			t.Errorf(errExpectedAge, 30, data.Age)
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
			t.Fatal(errExpectedError)
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
			t.Fatal(errExpectedError)
		}
	})
}

func TestBindJSONReadError(t *testing.T) {
	// Create a request with a reader that returns an error
	errReader := &errorReader{err: io.ErrUnexpectedEOF}
	r := httptest.NewRequest("POST", "/test", errReader)
	w := httptest.NewRecorder()
	c := New(w, r)

	var data map[string]any
	err := c.BindJSON(&data)

	if err == nil {
		t.Error(errReadError)
	}

	// Make sure we got the expected error
	if err != io.ErrUnexpectedEOF {
		t.Errorf(errWrongErrorType, io.ErrUnexpectedEOF, err)
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	message := "Success message"
	data := map[string]string{"key": "value"}

	c.Success(http.StatusOK, message, data)

	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if response.Status != "success" {
		t.Errorf(errExpectedStatus, "success", response.Status)
	}

	if response.Message != message {
		t.Errorf(errExpectedMessage, message, response.Message)
	}

	responseData, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatal(errDataConvert)
	}

	if val, ok := responseData["key"]; !ok || val != "value" {
		t.Errorf(errExpectedData, map[string]string{"key": "value"}, responseData)
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

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	if response.Status != "error" {
		t.Errorf(errExpectedStatus, "error", response.Status)
	}

	if response.Error != errorMsg {
		t.Errorf(errExpectedMessage, errorMsg, response.Error)
	}
}

func TestGetParam(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	c.Params = map[string]string{"id": "123"}

	param := c.GetParam("id")
	if param != "123" {
		t.Errorf(errExpectedParam, "123", param)
	}

	// Test non-existent param
	param = c.GetParam("unknown")
	if param != "" {
		t.Errorf(errExpectedEmpty, "param", param)
	}
}

func TestGetQuery(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	c := New(w, r)

	query := c.GetQuery("name")
	if query != "John" {
		t.Errorf(errExpectedParam, "John", query)
	}

	// Test non-existent query
	query = c.GetQuery("unknown")
	if query != "" {
		t.Errorf(errExpectedEmpty, "query", query)
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
		t.Errorf(errExpectedValue, value, c.Value(keyInstance))
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
		t.Errorf(errExpectedValue, value, retrievedValue)
	}

	// Test non-existent key
	type unknownKey struct{}
	unknownKeyInstance := unknownKey{}

	retrievedValue = c.Value(unknownKeyInstance)
	if retrievedValue != nil {
		t.Errorf(errExpectedNil, retrievedValue)
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

	contentType := w.Header().Get(headerContentType)
	if contentType != contentTypeJSON {
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

	mapData := map[string]any{
		"items": []string{"item1", "item2"},
		"count": 2,
	}

	c.JSONData(http.StatusCreated, mapData)

	if w.Code != http.StatusCreated {
		t.Errorf(errStatusCode, http.StatusCreated, w.Code)
	}

	var mapResult map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &mapResult); err != nil {
		t.Fatalf(errUnmarshalResponse, err)
	}

	// Check items array
	items, ok := mapResult["items"].([]any)
	if !ok {
		t.Fatal(errExpectedArray)
	}

	if len(items) != 2 {
		t.Errorf(errExpectedItems, 2, len(items))
	}

	// Check count
	count, ok := mapResult["count"].(float64) // JSON numbers are float64 in Go
	if !ok {
		t.Fatal(errExpectedNumber)
	}

	if count != 2 {
		t.Errorf(errExpectedCount, 2, count)
	}

	// Ensure no APIResponse wrapper
	_, hasStatus := mapResult["status"]
	if hasStatus {
		t.Error(errNoStatus)
	}
}

func TestJSONEncodingError(t *testing.T) {
	// Use our custom writer that always returns error on Write
	w := newMockErrorWriter()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	data := map[string]string{
		"key": "value",
	}

	// This should not panic despite the writer returning an error
	c.JSON(http.StatusOK, data)

	// Verify headers were set correctly
	contentType := w.Header().Get(headerContentType)
	if contentType != contentTypeJSON {
		t.Errorf(errContentType, contentType)
	}

	if w.code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.code)
	}
}

func TestJSONDataEncodingError(t *testing.T) {
	// Use our custom writer that always returns error on Write
	w := newMockErrorWriter()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	data := map[string]string{
		"key": "value",
	}

	// This should not panic despite the writer returning an error
	c.JSONData(http.StatusCreated, data)

	// Verify headers were set correctly
	contentType := w.Header().Get(headerContentType)
	if contentType != contentTypeJSON {
		t.Errorf(errContentType, contentType)
	}

	if w.code != http.StatusCreated {
		t.Errorf(errStatusCode, http.StatusCreated, w.code)
	}
}
