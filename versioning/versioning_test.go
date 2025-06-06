package versioning

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

// Path constants for testing
const (
	pathV1Users = "/v1/users"
	pathV2Users = "/v2/users"
	pathV3Users = "/v3/users"
	pathUsers   = "/users"

	// Version constants
	version1 = "1"
	version2 = "2"
	version3 = "3"

	// Parameter names
	paramVersion    = "version"
	paramAPIVersion = "api_version"

	// Header names
	headerAcceptVersion = "Accept-Version"
	headerXAPIVersion   = "X-API-Version"
	headerAccept        = "Accept"

	// Media types
	mediaTypeJSON           = "application/json"
	mediaTypeVndPrefix      = "application/vnd."
	mediaTypeVndAPIV1       = "application/vnd.api.v1+json"
	mediaTypeVndAPIV2       = "application/vnd.api.v2+json"
	mediaTypeMultipleWithV2 = "application/json, application/vnd.api.v2+json"

	// API version context key
	apiVersionKey = "API-Version"

	// Error messages
	errExpectedError         = "Expected error but got nil"
	errExpectedNoError       = "Expected no error but got: %v"
	errExpectedVersion       = "Expected version %s but got %s"
	errExpectedNoVersionInfo = "Expected no version info but got: %+v"
	errExpectedVersionInfo   = "Expected version info but got none"
	errExpectedStatus        = "Expected status %d but got %d"
)

// Test helper functions

// setupPathRequest creates a request for path versioning tests
func setupPathRequest(path string) (*http.Request, *httptest.ResponseRecorder, *context.Context) {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	return req, w, context.New(w, req)
}

// setupQueryRequest creates a request with query parameters for versioning tests
func setupQueryRequest(query string) (*http.Request, *httptest.ResponseRecorder, *context.Context) {
	req := httptest.NewRequest(http.MethodGet, pathUsers+query, nil)
	w := httptest.NewRecorder()
	return req, w, context.New(w, req)
}

// setupHeaderRequest prepares a request with the appropriate version header
func setupHeaderRequest(headerName, headerValue string) (*http.Request, *httptest.ResponseRecorder, *context.Context) {
	req := httptest.NewRequest(http.MethodGet, pathUsers, nil)

	// Set default header name if empty
	effectiveHeaderName := headerName
	if effectiveHeaderName == "" {
		effectiveHeaderName = headerAcceptVersion
	}

	// Set header value if provided
	if headerValue != "" {
		req.Header.Set(effectiveHeaderName, headerValue)
	}

	w := httptest.NewRecorder()
	return req, w, context.New(w, req)
}

// setupMediaTypeRequest creates a request with an Accept header for versioning tests
func setupMediaTypeRequest(acceptValue string) (*http.Request, *httptest.ResponseRecorder, *context.Context) {
	req := httptest.NewRequest(http.MethodGet, pathUsers, nil)
	if acceptValue != "" {
		req.Header.Set(headerAccept, acceptValue)
	}

	w := httptest.NewRecorder()
	return req, w, context.New(w, req)
}

// checkVersionResult validates version extraction results
func checkVersionResult(t *testing.T, version string, err error, expectedVer string, expectedError bool) {
	t.Helper()

	// Check error cases
	if expectedError {
		if err == nil {
			t.Error(errExpectedError)
		}
	} else if err != nil {
		t.Errorf(errExpectedNoError, err)
	}

	// Check version
	if version != expectedVer {
		t.Errorf(errExpectedVersion, expectedVer, version)
	}
}

// setupVersioningOptions creates versioning options with the provided settings
func setupVersioningOptions(supportedVersions []string, defaultVersion string, strictVersioning bool) *Options {
	return New().
		WithStrategy(&PathVersionStrategy{Prefix: "v"}).
		WithSupportedVersions(supportedVersions...).
		WithDefaultVersion(defaultVersion).
		WithStrictVersioning(strictVersioning)
}

// checkVersioningResults validates the middleware execution results
func checkVersioningResults(t *testing.T, statusCode int, capturedVersion string, expectedStatus int, expectedVersion string) {
	t.Helper()

	// Check status code
	if statusCode != expectedStatus {
		t.Errorf(errExpectedStatus, expectedStatus, statusCode)
	}

	// Check captured version if expecting success
	if expectedStatus == http.StatusOK && capturedVersion != expectedVersion {
		t.Errorf(errExpectedVersion, expectedVersion, capturedVersion)
	}
}

// createTestContext creates a context for testing version info
func createTestContext() *context.Context {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, pathV1Users, nil)
	return context.New(w, req)
}

// checkVersionInfo validates the version information
func checkVersionInfo(t *testing.T, info VersionInfo, exists bool, shouldExist bool, expectedInfo VersionInfo) {
	t.Helper()

	if !shouldExist && exists {
		t.Errorf(errExpectedNoVersionInfo, info)
		return
	}

	if shouldExist && !exists {
		t.Error(errExpectedVersionInfo)
		return
	}

	if shouldExist && exists && info.Version != expectedInfo.Version {
		t.Errorf(errExpectedVersion, expectedInfo.Version, info.Version)
	}
}

func TestPathVersionStrategy(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedVer   string
		expectedError bool
	}{
		{"Valid v1 path", pathV1Users, version1, false},
		{"Valid v2 path", pathV2Users, version2, false},
		{"No version in path", pathUsers, "", true},
		{"Custom prefix", "/api-v3/users", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, c := setupPathRequest(tt.path)
			strategy := &PathVersionStrategy{Prefix: "v"}
			version, err := strategy.ExtractVersion(c)
			checkVersionResult(t, version, err, tt.expectedVer, tt.expectedError)
		})
	}
}

func TestQueryVersionStrategy(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		paramName     string
		expectedVer   string
		expectedError bool
	}{
		{"Valid version parameter", "?version=1", paramVersion, version1, false},
		{"Valid v parameter", "?v=2", "", version2, false},
		{"Custom parameter name", "?api_version=3", paramAPIVersion, version3, false},
		{"Missing parameter", "?other=value", paramVersion, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, c := setupQueryRequest(tt.query)
			strategy := &QueryVersionStrategy{ParamName: tt.paramName}
			version, err := strategy.ExtractVersion(c)
			checkVersionResult(t, version, err, tt.expectedVer, tt.expectedError)
		})
	}
}

func TestHeaderVersionStrategy(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		expectedVer   string
		expectedError bool
	}{
		{"Default header name", "", version1, version1, false},
		{"Custom header name", headerXAPIVersion, version2, version2, false},
		{"Missing header", headerXAPIVersion, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, c := setupHeaderRequest(tt.headerName, tt.headerValue)
			strategy := &HeaderVersionStrategy{HeaderName: tt.headerName}
			version, err := strategy.ExtractVersion(c)
			checkVersionResult(t, version, err, tt.expectedVer, tt.expectedError)
		})
	}
}

func TestMediaTypeVersionStrategy(t *testing.T) {
	tests := []struct {
		name          string
		accept        string
		expectedVer   string
		expectedError bool
	}{
		{"Valid vendor media type", mediaTypeVndAPIV1, version1, false},
		{"Multiple media types with valid one", mediaTypeMultipleWithV2, version2, false},
		{"No valid vendor media type", mediaTypeJSON, "", true},
		{"Missing Accept header", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, c := setupMediaTypeRequest(tt.accept)
			strategy := &MediaTypeVersionStrategy{MediaTypePrefix: mediaTypeVndPrefix}
			version, err := strategy.ExtractVersion(c)
			checkVersionResult(t, version, err, tt.expectedVer, tt.expectedError)
		})
	}
}

func TestVersioningMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		supportedVers    []string
		defaultVer       string
		strictVersioning bool
		expectedStatus   int
		expectedVersion  string
	}{
		{"Valid version", pathV1Users, []string{version1, version2}, version1, false, http.StatusOK, version1},
		{"Unsupported version", pathV3Users, []string{version1, version2}, version1, false, http.StatusBadRequest, ""},
		{"No version with non-strict mode", pathUsers, []string{version1, version2}, version1, false, http.StatusOK, version1},
		{"No version with strict mode", pathUsers, []string{version1, version2}, version1, true, http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := setupVersioningOptions(tt.supportedVers, tt.defaultVer, tt.strictVersioning)
			var capturedVersion string
			handler := func(c *context.Context) {
				if versionInfo, exists := GetAPIVersion(c); exists {
					capturedVersion = versionInfo.Version
				}
				c.Status(http.StatusOK)
			}
			middleware := v.Middleware()(handler)
			_, w, c := setupPathRequest(tt.path)
			middleware(c)
			checkVersioningResults(t, w.Code, capturedVersion, tt.expectedStatus, tt.expectedVersion)
		})
	}
}

func TestGetAPIVersion(t *testing.T) {
	c := createTestContext()

	info, exists := GetAPIVersion(c)
	checkVersionInfo(t, info, exists, false, VersionInfo{})

	expectedInfo := VersionInfo{
		Version:     version2,
		IsSupported: true,
	}
	c.WithValue(apiVersionKey, expectedInfo)

	info, exists = GetAPIVersion(c)
	checkVersionInfo(t, info, exists, true, expectedInfo)
}
