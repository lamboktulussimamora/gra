// Package versioning provides API versioning capabilities for the GRA framework.
package versioning

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

const (
	// DefaultVersionHeader is the default HTTP header for version information
	DefaultVersionHeader = "Accept-Version"
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

// Options contains configuration for API versioning.
type Options struct {
	Strategy          VersionStrategy    // The versioning strategy to use
	DefaultVersion    string             // The default version to use if none is specified
	SupportedVersions []string           // List of supported versions
	StrictVersioning  bool               // If true, rejects requests that don't specify a version
	ErrorHandler      router.HandlerFunc // Custom handler for version errors
}

// New creates a new versioning middleware with default options
func New() *Options {
	return &Options{
		Strategy:          &PathVersionStrategy{Prefix: "v"},
		DefaultVersion:    "1",
		SupportedVersions: []string{"1"},
		StrictVersioning:  false,
		ErrorHandler:      nil,
	}
}

// WithStrategy sets the versioning strategy
func (vo *Options) WithStrategy(strategy VersionStrategy) *Options {
	vo.Strategy = strategy
	return vo
}

// WithDefaultVersion sets the default API version
func (vo *Options) WithDefaultVersion(version string) *Options {
	vo.DefaultVersion = version
	return vo
}

// WithSupportedVersions sets the supported API versions
func (vo *Options) WithSupportedVersions(versions ...string) *Options {
	vo.SupportedVersions = versions
	return vo
}

// WithStrictVersioning sets the strict versioning flag
func (vo *Options) WithStrictVersioning(strict bool) *Options {
	vo.StrictVersioning = strict
	return vo
}

// WithErrorHandler sets a custom error handler for version errors
func (vo *Options) WithErrorHandler(handler router.HandlerFunc) *Options {
	vo.ErrorHandler = handler
	return vo
}

// handleVersionError handles versioning errors with custom or default error responses
func (vo *Options) handleVersionError(c *context.Context, message string) {
	if vo.ErrorHandler != nil {
		vo.ErrorHandler(c)
	} else {
		c.Error(400, message)
	}
}

// isVersionSupported checks if the given version is in the list of supported versions
func (vo *Options) isVersionSupported(version string) bool {
	for _, v := range vo.SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// applyVersionToContext adds version information to the request context
func (vo *Options) applyVersionToContext(c *context.Context, version string) {
	// Apply version to the request
	vo.Strategy.Apply(c, version)

	// Store version info in context
	versionInfo := VersionInfo{
		Version:     version,
		IsSupported: true,
	}
	c.WithValue("API-Version", versionInfo)
}

// Middleware returns a middleware that applies API versioning
func (vo *Options) Middleware() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Extract version
			version, err := vo.Strategy.ExtractVersion(c)

			// Handle missing version
			if err != nil {
				if vo.StrictVersioning {
					vo.handleVersionError(c, "API version required")
					return
				}
				version = vo.DefaultVersion
			}

			// Check if version is supported
			if !vo.isVersionSupported(version) {
				vo.handleVersionError(c, fmt.Sprintf("API version %s is not supported", version))
				return
			}

			// Apply version and continue
			vo.applyVersionToContext(c, version)
			next(c)
		}
	}
}

// getDefaultPrefix returns the default prefix if none is provided
func getDefaultPrefix(prefix string) string {
	if prefix == "" {
		return "v"
	}
	return prefix
}

// extractPathSegments gets URL path segments without the leading slash
func extractPathSegments(path string) []string {
	return strings.Split(strings.TrimPrefix(path, "/"), "/")
}

// ExtractVersion extracts version from URL path
func (s *PathVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	path := c.Request.URL.Path
	prefix := getDefaultPrefix(s.Prefix)

	// Check if path contains version segment
	segments := extractPathSegments(path)
	if len(segments) == 0 {
		return "", fmt.Errorf("no version in path")
	}

	// Check if first segment matches our version format
	if strings.HasPrefix(segments[0], prefix) {
		return strings.TrimPrefix(segments[0], prefix), nil
	}

	return "", fmt.Errorf("no version in path")
}

// Apply doesn't need to do anything for path versioning
func (s *PathVersionStrategy) Apply(_ *context.Context, _ string) {
	// Path versioning is handled by the router, so we don't need to do anything here
}

// getVersionFromQuery attempts to get a version from a specific query param
func getVersionFromQuery(c *context.Context, paramName string) string {
	return c.GetQuery(paramName)
}

// ExtractVersion extracts version from query parameter
func (s *QueryVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	// If param name is specified, check only that param
	if s.ParamName != "" {
		v := getVersionFromQuery(c, s.ParamName)
		if v != "" {
			return v, nil
		}
		return "", fmt.Errorf("no version in query parameter %s", s.ParamName)
	}

	// Try common parameter names
	commonParams := []string{"version", "v"}
	for _, param := range commonParams {
		v := getVersionFromQuery(c, param)
		if v != "" {
			return v, nil
		}
	}

	return "", fmt.Errorf("no version in query parameters")
}

// Apply doesn't need to do anything for query versioning
func (s *QueryVersionStrategy) Apply(_ *context.Context, _ string) {
	// Query versioning is extracted from the request, so we don't need to do anything here
}

// getHeaderName returns the configured header name or the default
func (s *HeaderVersionStrategy) getHeaderName() string {
	if s.HeaderName == "" {
		return DefaultVersionHeader
	}
	return s.HeaderName
}

// ExtractVersion extracts version from HTTP header
func (s *HeaderVersionStrategy) ExtractVersion(c *context.Context) (string, error) {
	headerName := s.getHeaderName()
	v := c.GetHeader(headerName)

	if v == "" {
		return "", fmt.Errorf("no version in headers")
	}

	return v, nil
}

// Apply sets the header with the current version
func (s *HeaderVersionStrategy) Apply(c *context.Context, version string) {
	// Set the version in response header
	c.SetHeader(s.getHeaderName(), version)
}

// parseVersionFromMediaType attempts to extract a version from a media type string
func parseVersionFromMediaType(mediaType string, prefix string) (string, bool) {
	mediaType = strings.TrimSpace(mediaType)
	if !strings.HasPrefix(mediaType, prefix) {
		return "", false
	}

	// Format is typically: application/vnd.company.resource.v1+json
	parts := strings.Split(mediaType, ".")
	for _, part := range parts {
		if !strings.HasPrefix(part, "v") {
			continue
		}

		// Extract version number
		version := strings.TrimPrefix(part, "v")
		// Handle +json or similar suffix
		if idx := strings.Index(version, "+"); idx > 0 {
			version = version[:idx]
		}

		// Ensure it's a valid numeric version
		_, err := strconv.Atoi(version)
		if err == nil {
			return version, true
		}
	}

	return "", false
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
		version, found := parseVersionFromMediaType(mediaType, mediaTypePrefix)
		if found {
			return version, nil
		}
	}

	return "", fmt.Errorf("no version in Accept header")
}

// getMediaTypePrefix returns the configured media type prefix or the default
func getMediaTypePrefix(prefix string) string {
	if prefix == "" {
		return "application/vnd."
	}
	return prefix
}

// Apply sets the content type with the current version
func (s *MediaTypeVersionStrategy) Apply(c *context.Context, version string) {
	prefix := getMediaTypePrefix(s.MediaTypePrefix)

	// Set the content type with version
	contentType := fmt.Sprintf("%sAPI.v%s+json", prefix, version)
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
