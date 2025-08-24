package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

func TestSecureHeaders_HSTSPreloadAndCrossOrigin(t *testing.T) {
	cfg := DefaultSecureHeadersConfig()
	cfg.HSTSMaxAge = 63072000 // 2 years
	cfg.HSTSIncludeSubdomains = true
	cfg.HSTSPreload = true
	cfg.CrossOriginEmbedderPolicy = "require-corp"
	cfg.CrossOriginOpenerPolicy = "same-origin"
	cfg.CrossOriginResourcePolicy = "cross-origin"

	mw := SecureHeadersWithConfig(cfg)
	h := mw(func(c *context.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h(context.New(w, r))

	hdr := w.Result().Header
	if got := hdr.Get("Strict-Transport-Security"); got != "max-age=63072000; includeSubDomains; preload" {
		t.Fatalf("unexpected HSTS header: %q", got)
	}
	if got := hdr.Get("Cross-Origin-Embedder-Policy"); got != "require-corp" {
		t.Fatalf("unexpected COEP: %q", got)
	}
	if got := hdr.Get("Cross-Origin-Opener-Policy"); got != "same-origin" {
		t.Fatalf("unexpected COOP: %q", got)
	}
	if got := hdr.Get("Cross-Origin-Resource-Policy"); got != "cross-origin" {
		t.Fatalf("unexpected CORP: %q", got)
	}
}
