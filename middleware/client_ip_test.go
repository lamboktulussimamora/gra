package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPFromRequest_IgnoresHeadersWhenPeerNotTrusted(t *testing.T) {
	trusted, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("ParseTrustedProxies error: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.10:1234"
	r.Header.Set("X-Forwarded-For", "198.51.100.77")
	r.Header.Set("X-Real-IP", "198.51.100.88")

	if got := ClientIPFromRequest(r, trusted); got != "203.0.113.10" {
		t.Fatalf("expected remote IP when peer is untrusted, got %q", got)
	}
}

func TestClientIPFromRequest_UsesXFFWhenPeerTrusted(t *testing.T) {
	trusted, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("ParseTrustedProxies error: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:2222"
	r.Header.Set("X-Forwarded-For", "198.51.100.77, 10.0.0.1")

	if got := ClientIPFromRequest(r, trusted); got != "198.51.100.77" {
		t.Fatalf("expected client IP from XFF, got %q", got)
	}
}

func TestClientIPFromRequest_XFFStopsAtFirstUntrustedFromRight(t *testing.T) {
	trusted, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("ParseTrustedProxies error: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:2222"
	r.Header.Set("X-Forwarded-For", "198.51.100.77, 192.0.2.33, 10.0.0.1")

	if got := ClientIPFromRequest(r, trusted); got != "192.0.2.33" {
		t.Fatalf("expected first untrusted hop from right, got %q", got)
	}
}

func TestClientIPFromRequest_UsesXRealIPFallbackWhenPeerTrusted(t *testing.T) {
	trusted, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("ParseTrustedProxies error: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:2222"
	r.Header.Set("X-Real-IP", "198.51.100.88")

	if got := ClientIPFromRequest(r, trusted); got != "198.51.100.88" {
		t.Fatalf("expected client IP from X-Real-IP, got %q", got)
	}
}

func TestClientIPFromRequest_IPv6XFF(t *testing.T) {
	trusted, err := ParseTrustedProxies([]string{"2001:db8::/32"})
	if err != nil {
		t.Fatalf("ParseTrustedProxies error: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "[2001:db8::1]:2222"
	r.Header.Set("X-Forwarded-For", "2001:db8::2")

	if got := ClientIPFromRequest(r, trusted); got != "2001:db8::2" {
		t.Fatalf("expected IPv6 client IP from XFF, got %q", got)
	}
}
