package dns

import (
	"context"
	"net"
	"slices"
	"strings"
	"sync"
	"time"
)

// HostType represents host type.
type HostType int

// HostType representation.
const (
	HostUnknown HostType = iota
	HostPublic
	HostPrivate
)

// ClassifyHost returns host type.
func ClassifyHost(ctx context.Context, host string) HostType {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return HostUnknown
	}

	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return HostPrivate
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPublicIP(ip) {
			return HostPublic
		}
		return HostPrivate
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}

	var localIPs []net.IP
	wg.Go(func() {
		localIPs = resolve(ctx, net.DefaultResolver, host)
	})

	var cloudflareIPs []net.IP
	wg.Go(func() {
		cloudflareIPs = resolve(ctx, cloudflareResolver, host)
	})

	var googleIPs []net.IP
	wg.Go(func() {
		googleIPs = resolve(ctx, googleResolver, host)
	})

	wg.Wait()

	if len(localIPs) == 0 {
		return HostUnknown
	}

	publicIPs := slices.Concat(cloudflareIPs, googleIPs)
	if len(publicIPs) == 0 {
		return HostPrivate
	}

	if slices.ContainsFunc(localIPs, isPublicIP) && slices.ContainsFunc(publicIPs, isPublicIP) {
		return HostPublic
	}

	return HostPrivate
}

func resolve(ctx context.Context, resolver *net.Resolver, host string) []net.IP {
	ips, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil
	}
	return ips
}

//nolint:gochecknoglobals
var (
	cloudflareResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}
	googleResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
)

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		return !isPrivateIPv4(ip4)
	}
	return !isPrivateIPv6(ip)
}

// Reference: https://en.wikipedia.org/wiki/Private_network
//
//nolint:gochecknoglobals
var (
	privateIPv4CIDRs = []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"100.64.0.0/10",
	}
	privateIPv6CIDRs = []string{
		"fc00::/7",
		"fe80::/10",
	}
)

func isPrivateIPv4(ip net.IP) bool {
	for _, cidr := range privateIPv4CIDRs {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateIPv6(ip net.IP) bool {
	for _, cidr := range privateIPv6CIDRs {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
