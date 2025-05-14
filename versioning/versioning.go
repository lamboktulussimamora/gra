// Package versioning provides API versioning capabilities for the GRA framework.
package versioning

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// VersionStrategy defines the versioning strategy interface
type VersionStrategy interface {
	// ExtractVersion extracts the API version from the request
	ExtractVersion(c *context.Context) (string, error)
	// Apply applies the version to the request/response as needed
	Apply(c *context.Context, version string)
}

// PathVersionStrategy extracts version from URL path (/v1/resource)
type PathVersionStrategy struct {
	Prefix string // Optional prefix before version number (default: "v")
}

// QueryVersionStrategy extracts version from query parameter
type QueryVersionStrategy struct {
	ParamName string // The query parameter name (default: "version" or "v")
}

// HeaderVersionStrategy extracts version from HTTP header
type HeaderVersionStrategy struct {
	HeaderName string // The header name (default: "Accept-Version")
}

// MediaTypeVersionStrategy extracts version from the Accept header media type
type MediaTypeVersionStrategy struct {
	MediaTypePrefix string // The media type prefix (default: "application/vnd.")
}

// VersionInfo represents API version information
type VersionInfo struct {
	Version     string
	IsSupported bool
}

// VersioningOptions contains configuration for API versioning
type VersioningOptions struct {
	Strategy          VersionStrategy    // The versioning strategy to use
	DefaultVersion    string             // The default version to use if none is specified
	SupportedVersions []string           // List of supported versions
	StrictVersioning  bool               // If true, rejects requests that don't specify a version
	ErrorHandler      router.HandlerFunc // Custom handler for version errors
}

// New creates a new versioning middleware with default options
func New() *VersioningOptions {
	return &VersioningOptions{
		Strategy:          &PathVersionStrategy{Prefix: "v"},
		DefaultVersion:    "1",
		SupportedVersions: []string{"1"},
		StrictVersioning:  false,
		ErrorHandler:      nil,
	}
}

// WithStrategy sets the versioning strategy
func (vo *VersioningOptions) WithStrategy(strategy VersionStrategy) *VersioningOptions {
	vo.Strategy = strategy
	return vo
}

// WithDefaultVersion sets the default API version
func (vo *VersioningOptions) WithDefaultVersion(version string) *VersioningOptions {
	vo.DefaultVersion = version
	return vo
}

// WithSupportedVersions sets the supported API versions
func (vo *VersioningOptions) WithSupportedVersions(versions ...string) *VersioningOptions {
	vo.SupportedVersions = versions
	return vo
}

// WithStrictVersioning sets the strict versioning flag
func (vo *VersioningOptions) WithStrictVersioning(strict bool) *VersioningOptions {
	vo.StrictVersioning = strict
	return vo
}

// WithErrorHandler sets a custom error handler for version errors
func (vo *VersioningOptions) WithErrorHandler(handler router.HandlerFunc) *VersioningOptions {
	vo.ErrorHandler = handler
	return vo
}

// Middleware returns a middleware that applies API versioning
func (vo *VersioningOptions) Middleware() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Extract version
			version, err := vo.Strategy.ExtractVersion(c)

			// Handle missing version
			if err != nil {
				if vo.StrictVersioning {
					if vo.ErrorHandler != nil {
						vo.ErrorHandler(c)
					} else {
						c.Error(400, "API version required")
					}
					return
				}
				// Use default version if not strict
				version = vo.DefaultVersion
			}

			// Check if version is supported
			isSupported := false
			for _, v := range vo.SupportedVersions {
				if v == version {
					isSupported = true
					break
				}
			}

			if !isSupported {
				if vo.ErrorHandler != nil {
					vo.ErrorHandler(c)
				} else {
					c.Error(400, fmt.Sprintf("API version %s is not supported", version))
				}
				return
			}

			// Apply version to the request
			vo.Strategy.Apply(c, version)

			// Store version info in context
			versionInfo := VersionInfo{
				Version:     version,
				IsSupported: true,
			}
			c.WithValue("API-Version", versionInfo)

			// Continue with next handler
			next(c)
		}
	}
}

// ExtractVersion extracts version from URL path
func (s *PathVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	path := c.Request.URL.Path

	prefix := s.Prefix
	if prefix == "" {
		prefix = "v"
	}

	// Check if path contains version segment
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) == 0 {
		return "", fmt.Errorf("no version in path")
	}

	// Check if first segment matches our version format
	if strings.HasPrefix(segments[0], prefix) {
		versionStr := strings.TrimPrefix(segments[0], prefix)
		return versionStr, nil
	}

	return "", fmt.Errorf("no version in path")
}

// Apply doesn't need to do anything for path versioning
func (s *PathVersionStrategy) Apply(c *context.Context, version string) {
	// Path versioning is handled by the router, so we don't need to do anything here
}

// ExtractVersion extracts version from query parameter
func (s *QueryVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	paramName := s.ParamName
	if paramName == "" {
		// Try common parameter names
		v := c.GetQuery("version")
		if v != "" {
			return v, nil
		}
		v = c.GetQuery("v")
		if v != "" {
			return v, nil
		}
		return "", fmt.Errorf("no version in query parameters")
	}

	v := c.GetQuery(paramName)
	if v == "" {
		return "", fmt.Errorf("no version in query parameters")
	}

	return v, nil
}

// Apply doesn't need to do anything for query versioning
func (s *QueryVersionStrategy) Apply(c *context.Context, version string) {
	// Query versioning is extracted from the request, so we don't need to do anything here
}

// ExtractVersion extracts version from HTTP header
func (s *HeaderVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	headerName := s.HeaderName
	if headerName == "" {
		headerName = "Accept-Version"
	}

	v := c.GetHeader(headerName)
	if v == "" {
		return "", fmt.Errorf("no version in headers")
	}

	return v, nil
}

// Apply sets the header with the current version
func (s *HeaderVersionStrategy) Apply(c *context.Context, version string) {
	headerName := s.HeaderName
	if headerName == "" {
		headerName = "Accept-Version"
	}

	// Set the version in response header
	c.SetHeader(headerName, version)
}

// ExtractVersion extracts version from Accept header media type
func (s *MediaTypeVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	mediaTypePrefix := s.MediaTypePrefix
	if mediaTypePrefix == "" {
		mediaTypePrefix = "application/vnd."
	}

	accept := c.GetHeader("Accept")
	if accept == "" {
		return "", fmt.Errorf("no Accept header")
	}

	// Parse Accept header and look for vendor media type
	mediaTypes := strings.Split(accept, ",")
	for _, mediaType := range mediaTypes {
		mediaType = strings.TrimSpace(mediaType)
		if strings.HasPrefix(mediaType, mediaTypePrefix) {
			// Format is typically: application/vnd.company.resource.v1+json
			parts := strings.Split(mediaType, ".")
			for _, part := range parts {
				if strings.HasPrefix(part, "v") {
					// Extract version number
					version := strings.TrimPrefix(part, "v")
					// Handle +json or similar suffix
					if idx := strings.Index(version, "+"); idx > 0 {
						version = version[:idx]
					}
					_, err := strconv.Atoi(version) // Validate it's a number
					if err == nil {
						return version, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no version in Accept header")
}

// Apply sets the content type with the current version
func (s *MediaTypeVersionStrategy) Apply(c *context.Context, version string) {
	mediaTypePrefix := s.MediaTypePrefix
	if mediaTypePrefix == "" {
		mediaTypePrefix = "application/vnd."
	}

	// Set the content type with version
	contentType := fmt.Sprintf("%sapi.v%s+json", mediaTypePrefix, version)
	c.SetHeader("Content-Type", contentType)
}

// GetAPIVersion retrieves the API version from the context
func GetAPIVersion(c *context.Context) (VersionInfo, bool) {
	if v := c.Value("API-Version"); v != nil {
		if versionInfo, ok := v.(VersionInfo); ok {
			return versionInfo, true
		}
	}
	return VersionInfo{}, false
}
