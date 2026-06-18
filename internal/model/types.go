package model

import (
	"net"
	"time"
)

type ScanConfig struct {
	Targets   []net.IP
	Ports     PortSet
	Timeout   time.Duration
	Workers   int
	Retries   int
	Services  []string
	Multicast bool
}

type PortSet map[int]struct{}

func (p PortSet) Contains(port int) bool {
	_, ok := p[port]
	return ok
}

type Service struct {
	Port         uint16
	Proto        string
	Type         string
	InstanceName string
	Name         string
	Hostname     string
	IPv4         []string
	IPv6         []string
	TTL          uint32
	TXT          []string
	DeviceInfo   bool
}

type Asset struct {
	TargetIP   string
	Hostname   string
	IPv4       []string
	IPv6       []string
	TTL        uint32
	Services   []Service
	DeviceInfo []Service
	PTRAnswers []string
}

type ScanError struct {
	TargetIP string
	Err      string
}

type ScanResult struct {
	Assets []Asset
	Errors []ScanError
}
