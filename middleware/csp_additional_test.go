package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

// TestCSPBuilder_AllDirectives ensures all directive helper methods add content
func TestCSPBuilder_AllDirectives(t *testing.T) {
	b := NewCSPBuilder()
	b.DefaultSrc("'self'")
	b.ScriptSrc("'self'", "cdn.example.com")
	b.StyleSrc("'self'")
	b.ImgSrc("data:", "https:")
	b.ConnectSrc("https://api.example.com")
	b.FontSrc("https://fonts.example.com")
	b.ObjectSrc("'none'")
	b.MediaSrc("https://media.example.com")
	b.FrameSrc("https://frame.example.com")
	b.WorkerSrc("'self'")
	b.FrameAncestors("'none'")
	b.FormAction("'self'")
	b.ReportTo("csp-endpoint")
	b.ReportURI("/csp-report")
	b.UpgradeInsecureRequests()

	csp := b.Build()

	// Validate presence of all directives
	want := []string{
		"default-src 'self'",
		"script-src 'self' cdn.example.com",
		"style-src 'self'",
		"img-src data: https:",
		"connect-src https://api.example.com",
		"font-src https://fonts.example.com",
		"object-src 'none'",
		"media-src https://media.example.com",
		"frame-src https://frame.example.com",
		"worker-src 'self'",
		"frame-ancestors 'none'",
		"form-action 'self'",
		"report-to csp-endpoint",
		"report-uri /csp-report",
		"upgrade-insecure-requests",
	}
	for _, w := range want {
		if !strings.Contains(csp, w) {
			t.Fatalf("expected CSP to contain %q, got %q", w, csp)
		}
	}

	// Also verify the middleware wires it into the header
	mw := CSP(b)
	handler := mw(func(c *context.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(context.New(w, r))
	if got := w.Header().Get("Content-Security-Policy"); got != csp {
		t.Fatalf("expected CSP header to equal built policy, got %q", got)
	}
}

// TestRateLimit_Default exercises the thin wrapper around RateLimitWithConfig
func TestRateLimit_Default(t *testing.T) {
	mw := RateLimit(1, 1) // allow a single request per second
	called := 0
	handler := mw(func(c *context.Context) { called++; c.Status(http.StatusOK) })

	// First request should pass
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(context.New(w1, r1))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on first request, got %d", w1.Code)
	}

	// Second request should be rate-limited
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(context.New(w2, r2))
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on second request, got %d", w2.Code)
	}

	if called != 1 {
		t.Fatalf("expected handler called once, got %d", called)
	}
}
