package util

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

func ParseTargets(input string) ([]net.IP, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return nil, fmt.Errorf("--cidr is required")
	}
	if !strings.Contains(value, "/") {
		return parseSingleTarget(value)
	}
	return parseCIDRTargets(value)
}

func parseSingleTarget(value string) ([]net.IP, error) {
	ip := net.ParseIP(value)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip: %s", value)
	}
	if ip.To4() == nil {
		return nil, fmt.Errorf("only IPv4 targets are supported")
	}
	return []net.IP{ip.To4()}, nil
}

func parseCIDRTargets(value string) ([]net.IP, error) {
	ip, network, err := net.ParseCIDR(value)
	if err != nil {
		return nil, fmt.Errorf("invalid cidr: %w", err)
	}
	if ip.To4() == nil {
		return nil, fmt.Errorf("only IPv4 CIDR targets are supported")
	}
	return cidrHosts(network), nil
}

func cidrHosts(network *net.IPNet) []net.IP {
	first := ipToInt(network.IP.To4())
	ones, bits := network.Mask.Size()
	count := hostCount(bits, ones)
	start, end := hostBounds(count)

	result := make([]net.IP, 0, end-start)
	for i := start; i < end; i++ {
		next := big.NewInt(0).Add(first, big.NewInt(int64(i)))
		result = append(result, intToIP(next))
	}
	return result
}

func hostCount(bits, ones int) int {
	total := big.NewInt(1)
	total.Lsh(total, uint(bits-ones))
	return int(total.Int64())
}

func hostBounds(count int) (int, int) {
	if count > 2 {
		return 1, count - 1
	}
	return 0, count
}

func ipToInt(ip net.IP) *big.Int {
	return big.NewInt(0).SetBytes(ip.To4())
}

func intToIP(value *big.Int) net.IP {
	bytes := value.Bytes()
	ip := make([]byte, 4)
	copy(ip[4-len(bytes):], bytes)
	return net.IP(ip)
}
