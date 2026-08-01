package speedtester

import (
	"net"
	"strings"
)

// HasPublicIPv6 reports whether the system has a globally routable IPv6
// address on an up, non-loopback interface. It excludes loopback, link-local,
// ULA/private, multicast, and unspecified addresses so virtual adapters
// (mihomo fake-ip, Tailscale, etc.) don't trigger false positives.
func HasPublicIPv6() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if isPublicIPv6(parseAddrIP(addr.String())) {
				return true
			}
		}
	}
	return false
}

// isPublicIPv6 reports whether ip is a globally routable IPv6 address.
func isPublicIPv6(ip net.IP) bool {
	if ip == nil {
		return false
	}
	v6 := ip.To16()
	if v6 == nil || v6.To4() != nil {
		return false // not IPv6, or IPv4 mapped in IPv6
	}
	if v6.IsLoopback() || v6.IsLinkLocalUnicast() || v6.IsLinkLocalMulticast() ||
		v6.IsInterfaceLocalMulticast() || v6.IsPrivate() || v6.IsUnspecified() {
		return false
	}
	return true
}

// parseAddrIP strips the "/prefix" mask from an iface.Addrs() string and
// returns the parsed IP, or nil if it cannot be parsed.
func parseAddrIP(s string) net.IP {
	host := s
	if idx := strings.LastIndex(s, "/"); idx >= 0 {
		host = s[:idx]
	}
	return net.ParseIP(host)
}
