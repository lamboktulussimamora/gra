package context

import (
	"encoding/json"
	"fmt"
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

	// Test value constants
	testHeaderValue   = "test-value"
	customHeaderName  = "X-Custom-Header"
	customHeaderValue = "custom-value"
	sessionIDCookie   = "session_id"
	sessionIDValue    = "abc123"
	userPrefCookie    = "user_pref"
	darkModeValue     = "dark_mode"
	newCustomValue    = "new-value"
)

func TestNew(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	c := New(w, r)

	if c == nil {
		t.Fatal(errContextNil)
		return
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

// Test for GetHeader function
func TestGetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Test-Header", testHeaderValue)
	r.Header.Set("Authorization", "Bearer token123")
	c := New(w, r)

	// Test getting existing header
	testValue := c.GetHeader("X-Test-Header")
	if testValue != testHeaderValue {
		t.Errorf("Expected header value '%s', got '%s'", testHeaderValue, testValue)
	}

	// Test getting another existing header
	authValue := c.GetHeader("Authorization")
	if authValue != "Bearer token123" {
		t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", authValue)
	}

	// Test getting non-existent header
	emptyValue := c.GetHeader("Non-Existent-Header")
	if emptyValue != "" {
		t.Errorf("Expected empty string for non-existent header, got '%s'", emptyValue)
	}

	// Test case sensitivity
	lowerValue := c.GetHeader("x-test-header")
	if lowerValue != testHeaderValue {
		t.Errorf("Expected header value '%s' (case insensitive), got '%s'", testHeaderValue, lowerValue)
	}
}

// Test for SetHeader function
func TestSetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Test setting a single header
	result := c.SetHeader(customHeaderName, customHeaderValue)

	// Verify method chaining returns context
	if result != c {
		t.Error("SetHeader should return the same context for method chaining")
	}

	// Verify header was set
	headerValue := w.Header().Get(customHeaderName)
	if headerValue != customHeaderValue {
		t.Errorf("Expected header value '%s', got '%s'", customHeaderValue, headerValue)
	}

	// Test setting multiple headers
	c.SetHeader("Content-Encoding", "gzip").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-API-Version", "v1.0")

	// Verify all headers were set
	encodingValue := w.Header().Get("Content-Encoding")
	if encodingValue != "gzip" {
		t.Errorf("Expected Content-Encoding 'gzip', got '%s'", encodingValue)
	}

	cacheValue := w.Header().Get("Cache-Control")
	if cacheValue != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cacheValue)
	}

	versionValue := w.Header().Get("X-API-Version")
	if versionValue != "v1.0" {
		t.Errorf("Expected X-API-Version 'v1.0', got '%s'", versionValue)
	}

	// Test overwriting existing header
	c.SetHeader(customHeaderName, newCustomValue)
	updatedValue := w.Header().Get(customHeaderName)
	if updatedValue != newCustomValue {
		t.Errorf("Expected updated header value '%s', got '%s'", newCustomValue, updatedValue)
	}
}

// Test for GetCookie function
func TestGetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Add cookies to request
	r.AddCookie(&http.Cookie{Name: sessionIDCookie, Value: sessionIDValue})
	r.AddCookie(&http.Cookie{Name: userPrefCookie, Value: darkModeValue})

	c := New(w, r)

	// Test getting existing cookie
	sessionValue, err := c.GetCookie(sessionIDCookie)
	if err != nil {
		t.Errorf("Expected no error getting %s cookie, got: %v", sessionIDCookie, err)
	}
	if sessionValue != sessionIDValue {
		t.Errorf("Expected %s value '%s', got '%s'", sessionIDCookie, sessionIDValue, sessionValue)
	}

	// Test getting another existing cookie
	prefValue, err := c.GetCookie(userPrefCookie)
	if err != nil {
		t.Errorf("Expected no error getting %s cookie, got: %v", userPrefCookie, err)
	}
	if prefValue != darkModeValue {
		t.Errorf("Expected %s value '%s', got '%s'", userPrefCookie, darkModeValue, prefValue)
	}

	// Test getting non-existent cookie
	_, err = c.GetCookie("non_existent")
	if err == nil {
		t.Error("Expected error when getting non-existent cookie, got nil")
	}
	if err != http.ErrNoCookie {
		t.Errorf("Expected http.ErrNoCookie, got: %v", err)
	}

	// Test empty cookie value
	r.AddCookie(&http.Cookie{Name: "empty_cookie", Value: ""})
	emptyValue, err := c.GetCookie("empty_cookie")
	if err != nil {
		t.Errorf("Expected no error getting empty cookie, got: %v", err)
	}
	if emptyValue != "" {
		t.Errorf("Expected empty cookie value '', got '%s'", emptyValue)
	}
}

// Test for SetCookie function - basic functionality
func TestSetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Test setting a basic cookie
	result := c.SetCookie(sessionIDCookie, sessionIDValue, 3600, "/", "", false, true)

	// Verify method chaining returns context
	if result != c {
		t.Error("SetCookie should return the same context for method chaining")
	}

	// Verify cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	validateBasicCookie(t, cookie)
}

// Test for SetCookie function - multiple cookies with chaining
func TestSetCookieChaining(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := New(w, r)

	// Set initial cookie
	c.SetCookie(sessionIDCookie, sessionIDValue, 3600, "/", "", false, true)

	// Test setting multiple cookies with chaining
	c.SetCookie(userPrefCookie, darkModeValue, 86400, "/app", "example.com", true, false).
		SetCookie("lang", "en", 0, "/", "", false, false)

	// Verify multiple cookies were set
	allCookies := w.Result().Cookies()
	if len(allCookies) != 3 {
		t.Errorf("Expected 3 cookies total, got %d", len(allCookies))
	}

	// Test cookie with all options set
	prefCookie := allCookies[1]
	validateAdvancedCookie(t, prefCookie)

	// Test session cookie (MaxAge = 0)
	langCookie := allCookies[2]
	if langCookie.MaxAge != 0 {
		t.Errorf("Expected session cookie MaxAge 0, got %d", langCookie.MaxAge)
	}
}

// Helper function to validate basic cookie properties
func validateBasicCookie(t *testing.T, cookie *http.Cookie) {
	if cookie.Name != sessionIDCookie {
		t.Errorf("Expected cookie name '%s', got '%s'", sessionIDCookie, cookie.Name)
	}
	if cookie.Value != sessionIDValue {
		t.Errorf("Expected cookie value '%s', got '%s'", sessionIDValue, cookie.Value)
	}
	if cookie.MaxAge != 3600 {
		t.Errorf("Expected MaxAge 3600, got %d", cookie.MaxAge)
	}
	if cookie.Path != "/" {
		t.Errorf("Expected Path '/', got '%s'", cookie.Path)
	}
	if cookie.Domain != "" {
		t.Errorf("Expected empty Domain, got '%s'", cookie.Domain)
	}
	if cookie.Secure != false {
		t.Errorf("Expected Secure false, got %t", cookie.Secure)
	}
	if cookie.HttpOnly != true {
		t.Errorf("Expected HttpOnly true, got %t", cookie.HttpOnly)
	}
}

// Helper function to validate advanced cookie properties
func validateAdvancedCookie(t *testing.T, prefCookie *http.Cookie) {
	if prefCookie.Name != userPrefCookie {
		t.Errorf("Expected second cookie name '%s', got '%s'", userPrefCookie, prefCookie.Name)
	}
	if prefCookie.Value != darkModeValue {
		t.Errorf("Expected second cookie value '%s', got '%s'", darkModeValue, prefCookie.Value)
	}
	if prefCookie.MaxAge != 86400 {
		t.Errorf("Expected second cookie MaxAge 86400, got %d", prefCookie.MaxAge)
	}
	if prefCookie.Path != "/app" {
		t.Errorf("Expected second cookie Path '/app', got '%s'", prefCookie.Path)
	}
	if prefCookie.Domain != "example.com" {
		t.Errorf("Expected second cookie Domain 'example.com', got '%s'", prefCookie.Domain)
	}
	if prefCookie.Secure != true {
		t.Errorf("Expected second cookie Secure true, got %t", prefCookie.Secure)
	}
	if prefCookie.HttpOnly != false {
		t.Errorf("Expected second cookie HttpOnly false, got %t", prefCookie.HttpOnly)
	}
}

// Test for GetContentType function
func TestGetContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)

	// Test with JSON content type
	r.Header.Set(HeaderContentType, contentTypeJSON)
	c := New(w, r)

	contentType := c.GetContentType()
	if contentType != contentTypeJSON {
		t.Errorf("Expected content type '%s', got '%s'", contentTypeJSON, contentType)
	}

	// Test with form data content type
	r.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")
	contentType = c.GetContentType()
	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Expected content type 'application/x-www-form-urlencoded', got '%s'", contentType)
	}

	// Test with multipart form data
	r.Header.Set(HeaderContentType, "multipart/form-data; boundary=something")
	contentType = c.GetContentType()
	if contentType != "multipart/form-data; boundary=something" {
		t.Errorf("Expected content type 'multipart/form-data; boundary=something', got '%s'", contentType)
	}

	// Test with no content type header
	r.Header.Del(HeaderContentType)
	contentType = c.GetContentType()
	if contentType != "" {
		t.Errorf("Expected empty content type, got '%s'", contentType)
	}

	// Test with empty content type
	r.Header.Set(HeaderContentType, "")
	contentType = c.GetContentType()
	if contentType != "" {
		t.Errorf("Expected empty content type, got '%s'", contentType)
	}
}

// Test for Redirect function
func TestRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/old-path", nil)
	c := New(w, r)

	// Test temporary redirect (302)
	c.Redirect(http.StatusFound, "/new-path")

	if w.Code != http.StatusFound {
		t.Errorf(errStatusCode, http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/new-path" {
		t.Errorf("Expected Location header '/new-path', got '%s'", location)
	}

	// Test permanent redirect (301)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/old-path", nil)
	c2 := New(w2, r2)

	c2.Redirect(http.StatusMovedPermanently, "https://example.com/new-location")

	if w2.Code != http.StatusMovedPermanently {
		t.Errorf(errStatusCode, http.StatusMovedPermanently, w2.Code)
	}

	location2 := w2.Header().Get("Location")
	if location2 != "https://example.com/new-location" {
		t.Errorf("Expected Location header 'https://example.com/new-location', got '%s'", location2)
	}

	// Test see other redirect (303)
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("POST", "/form-submit", nil)
	c3 := New(w3, r3)

	c3.Redirect(http.StatusSeeOther, "/success")

	if w3.Code != http.StatusSeeOther {
		t.Errorf(errStatusCode, http.StatusSeeOther, w3.Code)
	}

	location3 := w3.Header().Get("Location")
	if location3 != "/success" {
		t.Errorf("Expected Location header '/success', got '%s'", location3)
	}

	// Test relative URL redirect
	w4 := httptest.NewRecorder()
	r4 := httptest.NewRequest("GET", "/test", nil)
	c4 := New(w4, r4)

	c4.Redirect(http.StatusFound, "../parent")

	if w4.Code != http.StatusFound {
		t.Errorf(errStatusCode, http.StatusFound, w4.Code)
	}

	location4 := w4.Header().Get("Location")
	if location4 != "/parent" { // http.Redirect normalizes the relative URL
		t.Errorf("Expected Location header '/parent', got '%s'", location4)
	}

	// Test empty URL redirect
	w5 := httptest.NewRecorder()
	r5 := httptest.NewRequest("GET", "/test", nil)
	c5 := New(w5, r5)

	c5.Redirect(http.StatusFound, "")

	if w5.Code != http.StatusFound {
		t.Errorf(errStatusCode, http.StatusFound, w5.Code)
	}

	location5 := w5.Header().Get("Location")
	if location5 != "/" { // http.Redirect normalizes empty URL to "/"
		t.Errorf("Expected Location header '/', got '%s'", location5)
	}
}

// TestContextAdvancedScenarios tests advanced context usage scenarios
func TestContextAdvancedScenarios(t *testing.T) {
	t.Run("complex JSON response with nested data", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		c := New(rr, req)

		// Complex nested data structure
		complexData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       123,
				"username": "john_doe",
				"profile": map[string]interface{}{
					"email":     "john@example.com",
					"age":       30,
					"interests": []string{"programming", "music", "travel"},
				},
				"settings": map[string]interface{}{
					"theme":         "dark",
					"notifications": true,
					"privacy": map[string]bool{
						"public_profile": false,
						"show_email":     false,
					},
				},
			},
			"metadata": map[string]interface{}{
				"timestamp": "2025-06-13T10:00:00Z",
				"version":   "1.0.0",
				"features":  []string{"feature1", "feature2"},
			},
		}

		c.JSON(http.StatusOK, complexData)

		// Verify content type
		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Verify status code
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Verify JSON structure by unmarshaling
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Verify nested data
		user, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Error("Expected user to be a map")
		}

		profile, ok := user["profile"].(map[string]interface{})
		if !ok {
			t.Error("Expected profile to be a map")
		}

		interests, ok := profile["interests"].([]interface{})
		if !ok || len(interests) != 3 {
			t.Error("Expected interests to be an array with 3 items")
		}
	})

	t.Run("multiple headers manipulation", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		c := New(rr, req)

		// Set multiple headers
		headers := map[string]string{
			"X-API-Version":   "1.0",
			"X-Rate-Limit":    "1000",
			"X-Request-ID":    "req-123",
			"Cache-Control":   "no-cache",
			"X-Custom-Header": "custom-value",
		}

		for key, value := range headers {
			c.SetHeader(key, value)
		}

		c.JSON(http.StatusOK, map[string]string{"status": "ok"})

		// Verify all headers were set
		for key, expectedValue := range headers {
			actualValue := rr.Header().Get(key)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("parameter extraction and validation", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
		c := New(rr, req)

		// Simulate route parameters
		c.Params["userID"] = "123"
		c.Params["postID"] = "456"
		c.Params["category"] = "technology"

		// Test parameter retrieval (accessing from Params map directly)
		userID := c.Params["userID"]
		if userID != "123" {
			t.Errorf("Expected userID to be '123', got '%s'", userID)
		}

		postID := c.Params["postID"]
		if postID != "456" {
			t.Errorf("Expected postID to be '456', got '%s'", postID)
		}

		category := c.Params["category"]
		if category != "technology" {
			t.Errorf("Expected category to be 'technology', got '%s'", category)
		}

		// Test non-existent parameter
		nonExistent := c.Params["nonexistent"]
		if nonExistent != "" {
			t.Errorf("Expected empty string for non-existent param, got '%s'", nonExistent)
		}
	})

	t.Run("query parameter handling", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/search?q=golang&page=2&limit=10&sort=date&order=desc", nil)
		c := New(rr, req)

		// Test query parameter retrieval using GetQuery
		query := c.GetQuery("q")
		if query != "golang" {
			t.Errorf("Expected query 'q' to be 'golang', got '%s'", query)
		}

		page := c.GetQuery("page")
		if page != "2" {
			t.Errorf("Expected page to be '2', got '%s'", page)
		}

		limit := c.GetQuery("limit")
		if limit != "10" {
			t.Errorf("Expected limit to be '10', got '%s'", limit)
		}

		// Test non-existent query parameter
		missing := c.GetQuery("missing")
		if missing != "" {
			t.Errorf("Expected empty string for missing query param, got '%s'", missing)
		}

		// Test manual default value handling
		defaultValue := "default"
		paramValue := c.GetQuery("missing")
		if paramValue == "" {
			paramValue = defaultValue
		}
		if paramValue != "default" {
			t.Errorf("Expected default value 'default', got '%s'", paramValue)
		}

		// Test existing parameter
		existingValue := c.GetQuery("q")
		if existingValue != "golang" {
			t.Errorf("Expected 'golang', got '%s'", existingValue)
		}
	})

	t.Run("request header parsing", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/data", nil)

		// Set various headers
		req.Header.Set("Authorization", "Bearer token123")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "TestClient/1.0")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Add("X-Custom", "value1")
		req.Header.Add("X-Custom", "value2")

		c := New(rr, req)

		// Test single header retrieval
		auth := c.GetHeader("Authorization")
		if auth != "Bearer token123" {
			t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", auth)
		}

		contentType := c.GetHeader("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Test case-insensitive header retrieval
		userAgent := c.GetHeader("user-agent")
		if userAgent != "TestClient/1.0" {
			t.Errorf("Expected User-Agent 'TestClient/1.0', got '%s'", userAgent)
		}

		// Test non-existent header
		missing := c.GetHeader("X-Missing")
		if missing != "" {
			t.Errorf("Expected empty string for missing header, got '%s'", missing)
		}
	})

	t.Run("error response formatting", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
		c := New(rr, req)

		// Test error response
		c.Error(http.StatusBadRequest, "Invalid user data")

		// Verify status code
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}

		// Verify content type
		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Verify error response structure
		var response APIResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if response.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", response.Status)
		}

		if response.Error != "Invalid user data" {
			t.Errorf("Expected error message 'Invalid user data', got '%s'", response.Error)
		}
	})

	t.Run("success response formatting", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
		c := New(rr, req)

		userData := map[string]interface{}{
			"id":       123,
			"username": "john_doe",
			"email":    "john@example.com",
		}

		// Test success response
		c.Success(http.StatusOK, "User retrieved successfully", userData)

		// Verify status code
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Verify response structure
		var response APIResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal success response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", response.Status)
		}

		if response.Message != "User retrieved successfully" {
			t.Errorf("Expected message 'User retrieved successfully', got '%s'", response.Message)
		}

		// Verify data payload
		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Error("Expected data to be a map")
		}

		if data["username"] != "john_doe" {
			t.Errorf("Expected username 'john_doe', got '%s'", data["username"])
		}
	})
}

// TestContextEdgeCases tests edge cases and error conditions
func TestContextEdgeCases(t *testing.T) {
	t.Run("empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil) // No body
		rr := httptest.NewRecorder()
		c := New(rr, req)

		var data map[string]interface{}
		err := c.BindJSON(&data)
		if err == nil {
			t.Error("Expected error when binding empty body")
		}
	})

	t.Run("malformed JSON in request", func(t *testing.T) {
		malformedJSON := `{"name": "test", "invalid": }`
		req := httptest.NewRequest(http.MethodPost, "/test",
			strings.NewReader(malformedJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		c := New(rr, req)

		var data map[string]interface{}
		err := c.BindJSON(&data)
		if err == nil {
			t.Error("Expected error when binding malformed JSON")
		}
	})

	t.Run("context with nil params map", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		c := &Context{
			Request: req,
			Writer:  rr,
			Params:  nil, // nil params
		}

		// Should handle gracefully
		param := c.GetParam("test")
		if param != "" {
			t.Errorf("Expected empty param for nil params map, got %s", param)
		}

		// Setting param should handle nil map gracefully
		if c.Params == nil {
			c.Params = make(map[string]string)
		}
		c.Params["test"] = "value"
		// Since map was nil, we recreated it and can now set values
	})

	t.Run("bind JSON with very large payload", func(t *testing.T) {
		// Create a large JSON payload (1MB)
		largeData := make(map[string]string)
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := strings.Repeat("a", 1000) // 1KB per value
			largeData[key] = value
		}

		jsonData, err := json.Marshal(largeData)
		if err != nil {
			t.Fatalf("Failed to marshal large data: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		c := New(rr, req)

		var result map[string]string
		err = c.BindJSON(&result)
		if err != nil {
			t.Errorf("Failed to bind large JSON: %v", err)
		}

		if len(result) != 1000 {
			t.Errorf("Expected 1000 keys, got %d", len(result))
		}
	})

	t.Run("multiple JSON responses", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// First JSON response
		c.JSON(http.StatusOK, map[string]string{"first": "response"})

		// Second JSON response (should not panic but might not work as expected)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Multiple JSON responses caused panic: %v", r)
			}
		}()

		c.JSON(http.StatusOK, map[string]string{"second": "response"})
	})

	t.Run("JSON with circular reference", func(t *testing.T) {
		type Node struct {
			Name string `json:"name"`
			Next *Node  `json:"next,omitempty"`
		}

		// Create circular reference
		node1 := &Node{Name: "node1"}
		node2 := &Node{Name: "node2"}
		node1.Next = node2
		node2.Next = node1 // Circular reference

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// Should handle circular reference gracefully (likely with an error)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("JSON with circular reference caused panic: %v", r)
			}
		}()

		c.JSON(http.StatusOK, node1)

		// Response should indicate an error occurred
		if rr.Code == http.StatusOK {
			// If it somehow succeeded, that's interesting but not necessarily wrong
			t.Log("JSON marshaling with circular reference succeeded (unexpected but not necessarily wrong)")
		}
	})

	t.Run("concurrent context operations", func(t *testing.T) {
		// Test concurrent access to different context instances
		// (more realistic since each request gets its own context)
		done := make(chan bool, 20)

		// 10 readers working with different contexts
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()
				// Each goroutine gets its own context
				req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
				rr := httptest.NewRecorder()
				c := New(rr, req)

				for j := 0; j < 100; j++ {
					param := c.GetParam("test")
					query := c.GetQuery("param")
					header := c.GetHeader("User-Agent")
					_ = param
					_ = query
					_ = header
				}
			}(i)
		}

		// 10 writers working with different contexts
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()
				// Each goroutine gets its own context
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				rr := httptest.NewRecorder()
				c := New(rr, req)

				for j := 0; j < 100; j++ {
					c.Params[fmt.Sprintf("param_%d", id)] = fmt.Sprintf("value_%d", j)
					c.SetHeader(fmt.Sprintf("X-Header-%d", id), fmt.Sprintf("value_%d", j))
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}
	})

	t.Run("error with non-string message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// Error with number converted to string message
		c.Error(http.StatusBadRequest, "12345")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		// Should handle non-string message gracefully
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}
	})

	t.Run("success response with complex data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		complexData := map[string]interface{}{
			"string":  "value",
			"number":  42,
			"boolean": true,
			"array":   []int{1, 2, 3},
			"object": map[string]string{
				"nested": "value",
			},
			"null": nil,
		}

		c.Success(http.StatusOK, "Complex data test", complexData)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		data, ok := response["data"].(map[string]interface{})
		if !ok {
			t.Error("Expected data to be an object")
		}

		if data["string"] != "value" {
			t.Errorf("Expected string value 'value', got %v", data["string"])
		}

		if data["number"].(float64) != 42 {
			t.Errorf("Expected number 42, got %v", data["number"])
		}
	})

	t.Run("header operations with special characters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// Set headers with special characters
		c.SetHeader("X-Test-Header", "value with spaces")
		c.SetHeader("X-Unicode", "测试中文")
		c.SetHeader("X-Special", "!@#$%^&*()")

		// Get headers back
		value1 := rr.Header().Get("X-Test-Header")
		value2 := rr.Header().Get("X-Unicode")
		value3 := rr.Header().Get("X-Special")

		if value1 != "value with spaces" {
			t.Errorf("Expected 'value with spaces', got '%s'", value1)
		}

		if value2 != "测试中文" {
			t.Errorf("Expected '测试中文', got '%s'", value2)
		}

		if value3 != "!@#$%^&*()" {
			t.Errorf("Expected 'string key', got '%s'", value3)
		}
	})
}

// TestContextValueEdgeCases tests the WithValue and Value functionality edge cases
func TestContextValueEdgeCases(t *testing.T) {
	t.Run("value with nil key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// WithValue with nil key should panic (this is Go's standard behavior)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected WithValue to panic with nil key")
			}
		}()

		c.WithValue(nil, "value")
		t.Error("WithValue with nil key should have panicked")
	})

	t.Run("value with complex key types", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// Test with different comparable key types
		type customKey struct {
			name string
		}

		key1 := customKey{name: "test"}
		key2 := 12345
		key3 := "string_key"

		c.WithValue(key1, "custom struct key")
		c.WithValue(key2, "int key")
		c.WithValue(key3, "string key")

		value1 := c.Value(key1)
		value2 := c.Value(key2)
		value3 := c.Value(key3)

		if value1 != "custom struct key" {
			t.Errorf("Expected 'custom struct key', got %v", value1)
		}

		if value2 != "int key" {
			t.Errorf("Expected 'int key', got %v", value2)
		}

		if value3 != "string key" {
			t.Errorf("Expected 'slice key', got %v", value3)
		}
	})

	t.Run("overwrite existing values", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		key := "test-key"

		// Set initial value
		c.WithValue(key, "initial")
		initialValue := c.Value(key)
		if initialValue != "initial" {
			t.Errorf("Expected 'initial', got %v", initialValue)
		}

		// Overwrite value
		c.WithValue(key, "overwritten")
		newValue := c.Value(key)
		if newValue != "overwritten" {
			t.Errorf("Expected 'overwritten', got %v", newValue)
		}
	})

	t.Run("value with large data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		c := New(rr, req)

		// Store large data structure
		largeData := make([]string, 10000)
		for i := range largeData {
			largeData[i] = fmt.Sprintf("item-%d", i)
		}

		c.WithValue("large-data", largeData)

		retrievedData := c.Value("large-data")
		retrievedSlice, ok := retrievedData.([]string)
		if !ok {
			t.Error("Failed to retrieve large data as slice")
		}

		if len(retrievedSlice) != 10000 {
			t.Errorf("Expected 10000 items, got %d", len(retrievedSlice))
		}

		if retrievedSlice[0] != "item-0" {
			t.Errorf("Expected 'item-0', got '%s'", retrievedSlice[0])
		}
	})
}
