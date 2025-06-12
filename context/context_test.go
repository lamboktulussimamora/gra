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
