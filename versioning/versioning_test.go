package versioning

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

func TestPathVersionStrategy(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedVer   string
		expectedError bool
	}{
		{
			name:          "Valid v1 path",
			path:          "/v1/users",
			expectedVer:   "1",
			expectedError: false,
		},
		{
			name:          "Valid v2 path",
			path:          "/v2/users",
			expectedVer:   "2",
			expectedError: false,
		},
		{
			name:          "No version in path",
			path:          "/users",
			expectedVer:   "",
			expectedError: true,
		},
		{
			name:          "Custom prefix",
			path:          "/api-v3/users",
			expectedVer:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with test path
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			c := context.New(w, req)

			// Create strategy and extract version
			strategy := &PathVersionStrategy{Prefix: "v"}
			version, err := strategy.ExtractVersion(c)

			// Check results
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if version != tt.expectedVer {
				t.Errorf("Expected version %s but got %s", tt.expectedVer, version)
			}
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
		{
			name:          "Valid version parameter",
			query:         "?version=1",
			paramName:     "version",
			expectedVer:   "1",
			expectedError: false,
		},
		{
			name:          "Valid v parameter",
			query:         "?v=2",
			paramName:     "",
			expectedVer:   "2",
			expectedError: false,
		},
		{
			name:          "Custom parameter name",
			query:         "?api_version=3",
			paramName:     "api_version",
			expectedVer:   "3",
			expectedError: false,
		},
		{
			name:          "Missing parameter",
			query:         "?other=value",
			paramName:     "version",
			expectedVer:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with test query
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			w := httptest.NewRecorder()
			c := context.New(w, req)

			// Create strategy and extract version
			strategy := &QueryVersionStrategy{ParamName: tt.paramName}
			version, err := strategy.ExtractVersion(c)

			// Check results
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if version != tt.expectedVer {
				t.Errorf("Expected version %s but got %s", tt.expectedVer, version)
			}
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
		{
			name:          "Default header name",
			headerName:    "",
			headerValue:   "1",
			expectedVer:   "1",
			expectedError: false,
		},
		{
			name:          "Custom header name",
			headerName:    "X-API-Version",
			headerValue:   "2",
			expectedVer:   "2",
			expectedError: false,
		},
		{
			name:          "Missing header",
			headerName:    "X-API-Version",
			headerValue:   "",
			expectedVer:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with test header
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			headerName := tt.headerName
			if headerName == "" {
				headerName = "Accept-Version"
			}
			if tt.headerValue != "" {
				req.Header.Set(headerName, tt.headerValue)
			}

			w := httptest.NewRecorder()
			c := context.New(w, req)

			// Create strategy and extract version
			strategy := &HeaderVersionStrategy{HeaderName: tt.headerName}
			version, err := strategy.ExtractVersion(c)

			// Check results
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if version != tt.expectedVer {
				t.Errorf("Expected version %s but got %s", tt.expectedVer, version)
			}
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
		{
			name:          "Valid vendor media type",
			accept:        "application/vnd.api.v1+json",
			expectedVer:   "1",
			expectedError: false,
		},
		{
			name:          "Multiple media types with valid one",
			accept:        "application/json, application/vnd.api.v2+json",
			expectedVer:   "2",
			expectedError: false,
		},
		{
			name:          "No valid vendor media type",
			accept:        "application/json",
			expectedVer:   "",
			expectedError: true,
		},
		{
			name:          "Missing Accept header",
			accept:        "",
			expectedVer:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with test Accept header
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			w := httptest.NewRecorder()
			c := context.New(w, req)

			// Create strategy and extract version
			strategy := &MediaTypeVersionStrategy{MediaTypePrefix: "application/vnd."}
			version, err := strategy.ExtractVersion(c)

			// Check results
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if version != tt.expectedVer {
				t.Errorf("Expected version %s but got %s", tt.expectedVer, version)
			}
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
		{
			name:             "Valid version",
			path:             "/v1/users",
			supportedVers:    []string{"1", "2"},
			defaultVer:       "1",
			strictVersioning: false,
			expectedStatus:   http.StatusOK,
			expectedVersion:  "1",
		},
		{
			name:             "Unsupported version",
			path:             "/v3/users",
			supportedVers:    []string{"1", "2"},
			defaultVer:       "1",
			strictVersioning: false,
			expectedStatus:   http.StatusBadRequest,
			expectedVersion:  "",
		},
		{
			name:             "No version with non-strict mode",
			path:             "/users",
			supportedVers:    []string{"1", "2"},
			defaultVer:       "1",
			strictVersioning: false,
			expectedStatus:   http.StatusOK,
			expectedVersion:  "1",
		},
		{
			name:             "No version with strict mode",
			path:             "/users",
			supportedVers:    []string{"1", "2"},
			defaultVer:       "1",
			strictVersioning: true,
			expectedStatus:   http.StatusBadRequest,
			expectedVersion:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create versioning options
			v := New().
				WithStrategy(&PathVersionStrategy{Prefix: "v"}).
				WithSupportedVersions(tt.supportedVers...).
				WithDefaultVersion(tt.defaultVer).
				WithStrictVersioning(tt.strictVersioning)

			// Create test handler
			var capturedVersion string
			handler := func(c *context.Context) {
				if versionInfo, exists := GetAPIVersion(c); exists {
					capturedVersion = versionInfo.Version
				}
				c.Status(http.StatusOK)
			}

			// Create middleware chain
			middleware := v.Middleware()(handler)

			// Create request with test path
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			c := context.New(w, req)

			// Execute middleware and handler
			middleware(c)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			// Check captured version if expecting success
			if tt.expectedStatus == http.StatusOK && capturedVersion != tt.expectedVersion {
				t.Errorf("Expected version %s but got %s", tt.expectedVersion, capturedVersion)
			}
		})
	}
}

func TestGetAPIVersion(t *testing.T) {
	// Create context
	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	w := httptest.NewRecorder()
	c := context.New(w, req)

	// Case 1: No version info in context
	info, exists := GetAPIVersion(c)
	if exists {
		t.Errorf("Expected no version info but got: %+v", info)
	}

	// Case 2: Version info in context
	expectedInfo := VersionInfo{
		Version:     "2",
		IsSupported: true,
	}
	c.WithValue("API-Version", expectedInfo)

	info, exists = GetAPIVersion(c)
	if !exists {
		t.Errorf("Expected version info but got none")
	}

	if info.Version != expectedInfo.Version {
		t.Errorf("Expected version %s but got %s", expectedInfo.Version, info.Version)
	}
}
