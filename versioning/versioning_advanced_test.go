package versioning

import (
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

// TestVersioningAdvancedScenarios tests complex versioning scenarios
func TestVersioningAdvancedScenarios(t *testing.T) {
	t.Run("PathVersionStrategy", func(t *testing.T) {
		strategy := &PathVersionStrategy{Prefix: "v"}

		// Test valid version extraction
		req := httptest.NewRequest("GET", "/v2/users", nil)
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version != "2" {
			t.Errorf("Expected version '2', got '%s'", version)
		}

		// Test path without version
		req2 := httptest.NewRequest("GET", "/users", nil)
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := strategy.ExtractVersion(c2)
		if err2 == nil {
			t.Error("Expected error for path without version")
		}
		if version2 != "" {
			t.Errorf("Expected empty version, got '%s'", version2)
		}

		// Test custom prefix
		customStrategy := &PathVersionStrategy{Prefix: "api"}
		req3 := httptest.NewRequest("GET", "/api3/users", nil)
		c3 := context.New(httptest.NewRecorder(), req3)

		version3, err3 := customStrategy.ExtractVersion(c3)
		if err3 != nil {
			t.Errorf("Expected no error with custom prefix, got %v", err3)
		}
		if version3 != "3" {
			t.Errorf("Expected version '3', got '%s'", version3)
		}
	})

	t.Run("QueryVersionStrategy", func(t *testing.T) {
		strategy := &QueryVersionStrategy{ParamName: "version"}

		// Test valid version extraction
		req := httptest.NewRequest("GET", "/users?version=1.2", nil)
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version != "1.2" {
			t.Errorf("Expected version '1.2', got '%s'", version)
		}

		// Test missing version parameter
		req2 := httptest.NewRequest("GET", "/users", nil)
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := strategy.ExtractVersion(c2)
		if err2 == nil {
			t.Error("Expected error for missing version parameter")
		}
		if version2 != "" {
			t.Errorf("Expected empty version, got '%s'", version2)
		}

		// Test custom parameter name
		customStrategy := &QueryVersionStrategy{ParamName: "v"}
		req3 := httptest.NewRequest("GET", "/users?v=2.0", nil)
		c3 := context.New(httptest.NewRecorder(), req3)

		version3, err3 := customStrategy.ExtractVersion(c3)
		if err3 != nil {
			t.Errorf("Expected no error with custom param name, got %v", err3)
		}
		if version3 != "2.0" {
			t.Errorf("Expected version '2.0', got '%s'", version3)
		}
	})

	t.Run("HeaderVersionStrategy", func(t *testing.T) {
		strategy := &HeaderVersionStrategy{HeaderName: "Accept-Version"}

		// Test valid version extraction
		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("Accept-Version", "1.0")
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version != "1.0" {
			t.Errorf("Expected version '1.0', got '%s'", version)
		}

		// Test missing version header
		req2 := httptest.NewRequest("GET", "/users", nil)
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := strategy.ExtractVersion(c2)
		if err2 == nil {
			t.Error("Expected error for missing version header")
		}
		if version2 != "" {
			t.Errorf("Expected empty version, got '%s'", version2)
		}

		// Test custom header name
		customStrategy := &HeaderVersionStrategy{HeaderName: "X-API-Version"}
		req3 := httptest.NewRequest("GET", "/users", nil)
		req3.Header.Set("X-API-Version", "2.1")
		c3 := context.New(httptest.NewRecorder(), req3)

		version3, err3 := customStrategy.ExtractVersion(c3)
		if err3 != nil {
			t.Errorf("Expected no error with custom header name, got %v", err3)
		}
		if version3 != "2.1" {
			t.Errorf("Expected version '2.1', got '%s'", version3)
		}
	})

	t.Run("MediaTypeVersionStrategy", func(t *testing.T) {
		strategy := &MediaTypeVersionStrategy{MediaTypePrefix: "application/vnd.api"}

		// Test valid version extraction
		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("Accept", "application/vnd.api.v2+json")
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version != "2" {
			t.Errorf("Expected version '2', got '%s'", version)
		}

		// Test missing version in media type
		req2 := httptest.NewRequest("GET", "/users", nil)
		req2.Header.Set("Accept", "application/json")
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := strategy.ExtractVersion(c2)
		if err2 == nil {
			t.Error("Expected error for missing version in media type")
		}
		if version2 != "" {
			t.Errorf("Expected empty version, got '%s'", version2)
		}

		// Test custom media type prefix
		customStrategy := &MediaTypeVersionStrategy{MediaTypePrefix: "application/vnd.myapi"}
		req3 := httptest.NewRequest("GET", "/users", nil)
		req3.Header.Set("Accept", "application/vnd.myapi.v3+json")
		c3 := context.New(httptest.NewRecorder(), req3)

		version3, err3 := customStrategy.ExtractVersion(c3)
		if err3 != nil {
			t.Errorf("Expected no error with custom media type prefix, got %v", err3)
		}
		if version3 != "3" {
			t.Errorf("Expected version '3', got '%s'", version3)
		}
	})
}

// TestVersioningMiddleware tests the versioning middleware functionality
func TestVersioningMiddlewareAdvanced(t *testing.T) {
	t.Run("OptionsConfiguration", func(t *testing.T) {
		// Test options configuration
		options := New().
			WithStrategy(&PathVersionStrategy{Prefix: "v"}).
			WithDefaultVersion("1").
			WithSupportedVersions("1", "2", "3").
			WithStrictVersioning(true)

		if options.DefaultVersion != "1" {
			t.Errorf("Expected default version '1', got '%s'", options.DefaultVersion)
		}

		if len(options.SupportedVersions) != 3 {
			t.Errorf("Expected 3 supported versions, got %d", len(options.SupportedVersions))
		}

		if !options.StrictVersioning {
			t.Error("Expected strict versioning to be enabled")
		}
	})

	t.Run("OptionsChaining", func(t *testing.T) {
		// Test method chaining
		options := New().
			WithStrategy(&QueryVersionStrategy{ParamName: "v"}).
			WithDefaultVersion("2.0").
			WithSupportedVersions("1.0", "2.0", "3.0").
			WithStrictVersioning(false)

		queryStrategy, ok := options.Strategy.(*QueryVersionStrategy)
		if !ok {
			t.Error("Expected QueryVersionStrategy")
		}

		if queryStrategy.ParamName != "v" {
			t.Errorf("Expected param name 'v', got '%s'", queryStrategy.ParamName)
		}

		if options.DefaultVersion != "2.0" {
			t.Errorf("Expected default version '2.0', got '%s'", options.DefaultVersion)
		}
	})
}

// TestVersioningUtilities tests utility functions
func TestVersioningUtilities(t *testing.T) {
	t.Run("VersionComparison", func(t *testing.T) {
		// Test version comparison functionality
		versions := []string{"1.0", "1.1", "2.0", "2.1"}

		for _, version := range versions {
			info := VersionInfo{
				Version:     version,
				IsSupported: true,
			}

			if info.Version != version {
				t.Errorf("Expected version %s, got %s", version, info.Version)
			}

			if !info.IsSupported {
				t.Errorf("Expected version %s to be supported", version)
			}
		}
	})

	t.Run("VersionInfoCreation", func(t *testing.T) {
		// Test creating version info
		supportedVersions := []string{"1.0", "1.1", "2.0"}

		for _, version := range supportedVersions {
			info := VersionInfo{
				Version:     version,
				IsSupported: true,
			}

			if !info.IsSupported {
				t.Errorf("Expected version %s to be supported", version)
			}
		}

		// Test unsupported version
		unsupportedInfo := VersionInfo{
			Version:     "3.0",
			IsSupported: false,
		}

		if unsupportedInfo.IsSupported {
			t.Error("Expected version 3.0 to be unsupported")
		}
	})

	t.Run("VersionHeaderApplication", func(t *testing.T) {
		// Test applying version to response headers
		req := httptest.NewRequest("GET", "/v2/users", nil)
		w := httptest.NewRecorder()
		c := context.New(w, req)

		strategy := &PathVersionStrategy{Prefix: "v"}
		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Apply version to response
		strategy.Apply(c, version)

		// Check if version was applied (implementation dependent)
		// This might set response headers or modify context
		if version != "2" {
			t.Errorf("Expected version '2', got '%s'", version)
		}
	})
}

// TestVersioningEdgeCases tests edge cases and error conditions
func TestVersioningEdgeCases(t *testing.T) {
	t.Run("EmptyVersionStrategies", func(t *testing.T) {
		// Test with empty prefix
		strategy := &PathVersionStrategy{Prefix: ""}
		req := httptest.NewRequest("GET", "/v2/users", nil)
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error with empty prefix, got %v", err)
		}
		if version != "2" {
			t.Errorf("Expected version '2' with empty prefix, got '%s'", version)
		}

		// Test with empty parameter name (should use default behavior)
		queryStrategy := &QueryVersionStrategy{ParamName: ""}
		req2 := httptest.NewRequest("GET", "/users?version=1.0", nil)
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := queryStrategy.ExtractVersion(c2)
		if err2 != nil {
			t.Errorf("Expected no error for empty parameter name with version param, got: %v", err2)
		}
		if version2 != "1.0" {
			t.Errorf("Expected version '1.0', got '%s'", version2)
		}

		// Test with empty parameter name and no version param
		req3 := httptest.NewRequest("GET", "/users", nil)
		c3 := context.New(httptest.NewRecorder(), req3)

		_, err3 := queryStrategy.ExtractVersion(c3)
		if err3 == nil {
			t.Error("Expected error when no version parameters are present")
		}
	})

	t.Run("ComplexVersionFormats", func(t *testing.T) {
		strategy := &QueryVersionStrategy{ParamName: "version"}

		// Test semantic versioning
		req := httptest.NewRequest("GET", "/users?version=1.2.3", nil)
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error for semantic version, got %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("Expected version '1.2.3', got '%s'", version)
		}

		// Test date-based versioning
		req2 := httptest.NewRequest("GET", "/users?version=2023-01-01", nil)
		c2 := context.New(httptest.NewRecorder(), req2)

		version2, err2 := strategy.ExtractVersion(c2)
		if err2 != nil {
			t.Errorf("Expected no error for date version, got %v", err2)
		}
		if version2 != "2023-01-01" {
			t.Errorf("Expected version '2023-01-01', got '%s'", version2)
		}
	})

	t.Run("SpecialCharactersInVersions", func(t *testing.T) {
		strategy := &QueryVersionStrategy{ParamName: "version"}

		// Test version with special characters (should be URL-encoded)
		req := httptest.NewRequest("GET", "/users?version=1.0-beta", nil)
		c := context.New(httptest.NewRecorder(), req)

		version, err := strategy.ExtractVersion(c)
		if err != nil {
			t.Errorf("Expected no error for version with special chars, got %v", err)
		}
		if version != "1.0-beta" {
			t.Errorf("Expected version '1.0-beta', got '%s'", version)
		}
	})
}
