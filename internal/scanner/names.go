package scanner

import (
	"strings"

	"github.com/miekg/dns"
)

func cleanName(name string) string {
	return strings.TrimSuffix(dns.Fqdn(name), ".")
}

func instanceLabel(instance, serviceType string) string {
	name := cleanName(instance)
	suffix := "." + cleanName(serviceType)
	return strings.TrimSuffix(name, suffix)
}

func ServiceLabel(serviceType string) string {
	parts := strings.Split(cleanName(serviceType), ".")
	if len(parts) == 0 {
		return cleanName(serviceType)
	}
	return strings.TrimPrefix(parts[0], "_")
}

func inferType(instance, fallback string) string {
	if fallback != "" && !strings.EqualFold(dns.Fqdn(fallback), DiscoveryService) {
		return dns.Fqdn(fallback)
	}
	parts := strings.SplitN(cleanName(instance), ".", 2)
	if len(parts) == 2 {
		return dns.Fqdn(parts[1])
	}
	return dns.Fqdn(fallback)
}

func protoOf(serviceType string) string {
	if strings.Contains(serviceType, "._udp.") {
		return "udp"
	}
	return "tcp"
}

func isDeviceInfo(serviceType string) bool {
	return strings.HasPrefix(strings.ToLower(cleanName(serviceType)), "_device-info.")
}
