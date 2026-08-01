package speedtester

import (
	"net"
	"testing"
)

func TestIsPublicIPv6(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{"public china mobile", "240e:3bb:3034:2f00::1", true},
		{"public cloudflare", "2606:4700:4700::1111", true},
		{"public google", "2001:4860:4860::8888", true},
		{"ula mihomo fake-ip", "fdfe:dcba:9876::1", false},
		{"ula tailscale", "fd7a:115c:a1e0:ab12::1", false},
		{"link-local", "fe80::1", false},
		{"loopback", "::1", false},
		{"unspecified", "::", false},
		{"ipv4", "1.1.1.1", false},
		{"ipv4-in-v6 mapped", "::ffff:1.1.1.1", false},
		{"invalid", "not-an-ip", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isPublicIPv6(net.ParseIP(c.ip))
			if got != c.want {
				t.Fatalf("isPublicIPv6(%s) = %v, want %v", c.ip, got, c.want)
			}
		})
	}
}

func TestHasPublicIPv6DoesNotPanic(t *testing.T) {
	// Result depends on the host's interfaces; only assert it runs without panic.
	_ = HasPublicIPv6()
}
