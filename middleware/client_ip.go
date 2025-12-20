package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/lamboktulussimamora/gra/context"
)

// ParseTrustedProxies parses a list of trusted proxy entries (CIDR blocks or plain IPs)
// into a slice of *net.IPNet.
func ParseTrustedProxies(entries []string) ([]*net.IPNet, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	nets := make([]*net.IPNet, 0, len(entries))
	for _, raw := range entries {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}

		if strings.Contains(s, "/") {
			_, ipNet, err := net.ParseCIDR(s)
			if err != nil {
				return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", s, err)
			}
			nets = append(nets, ipNet)
			continue
		}

		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf("invalid trusted proxy IP %q", s)
		}
		nets = append(nets, ipToIPNet(ip))
	}

	return nets, nil
}

func ipToIPNet(ip net.IP) *net.IPNet {
	if ip4 := ip.To4(); ip4 != nil {
		return &net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)}
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
}

func isTrustedProxy(ip net.IP, trusted []*net.IPNet) bool {
	if ip == nil || len(trusted) == 0 {
		return false
	}
	for _, n := range trusted {
		if n != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIPFromContext returns the best-effort client IP address for a request.
//
// Security note: proxy headers are only consulted when the immediate peer (RemoteAddr)
// is in a trusted proxy range.
func ClientIPFromContext(c *context.Context, trustedProxies []*net.IPNet) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return ClientIPFromRequest(c.Request, trustedProxies)
}

// ClientIPFromRequest returns the best-effort client IP address.
//
// If RemoteAddr is not a trusted proxy, it is returned (with any port stripped).
// If RemoteAddr is trusted, X-Forwarded-For is parsed and the last untrusted IP in
// the chain is returned. If unavailable, X-Real-IP is used as a fallback.
func ClientIPFromRequest(r *http.Request, trustedProxies []*net.IPNet) string {
	if r == nil {
		return ""
	}

	remoteHost := stripPort(r.RemoteAddr)
	remoteIP := net.ParseIP(remoteHost)
	if remoteIP == nil {
		// If we can't parse it, keep the original host string (minus port if any).
		return remoteHost
	}

	if !isTrustedProxy(remoteIP, trustedProxies) {
		return remoteIP.String()
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := clientIPFromXFF(xff, trustedProxies); ip != nil {
			return ip.String()
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		if ip := parseIPLoose(xri); ip != nil {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func stripPort(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	// Handle bracketed IPv6 without a port: "[2001:db8::1]".
	if strings.HasPrefix(addr, "[") {
		if end := strings.Index(addr, "]"); end > 0 {
			return addr[1:end]
		}
	}
	return addr
}

func clientIPFromXFF(xff string, trustedProxies []*net.IPNet) net.IP {
	parts := strings.Split(xff, ",")
	ips := make([]net.IP, 0, len(parts))
	for _, p := range parts {
		ip := parseIPLoose(p)
		if ip != nil {
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		return nil
	}

	// Walk right-to-left: drop trusted proxy hops; return the first untrusted IP.
	for i := len(ips) - 1; i >= 0; i-- {
		if !isTrustedProxy(ips[i], trustedProxies) {
			return ips[i]
		}
	}

	// All hops are trusted; best guess is left-most (original client).
	return ips[0]
}

func parseIPLoose(raw string) net.IP {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	// Trim optional quotes.
	s = strings.Trim(s, "\"")

	// Handle bracketed IPv6: "[2001:db8::1]:1234" or "[2001:db8::1]".
	if strings.HasPrefix(s, "[") {
		if end := strings.Index(s, "]"); end > 0 {
			inner := s[1:end]
			if ip := net.ParseIP(inner); ip != nil {
				return ip
			}
		}
	}

	// Try as plain IP first (covers unbracketed IPv6).
	if ip := net.ParseIP(s); ip != nil {
		return ip
	}

	// Try to strip port from IPv4/host:port.
	if host, _, err := net.SplitHostPort(s); err == nil {
		return net.ParseIP(host)
	}

	return nil
}
