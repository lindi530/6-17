package scanner

import "github.com/miekg/dns"

type ParsedResponse struct {
	ServiceTypes []string
	Records      []ServiceRecord
	PTRAnswers   []string
}

type MulticastResponse struct {
	From    string
	Message *dns.Msg
}

type ServiceRecord struct {
	Type         string
	InstanceName string
	Name         string
	Hostname     string
	Port         uint16
	Proto        string
	TTL          uint32
	TXT          []string
	IPv4         []string
	IPv6         []string
	HasSRV       bool
	DeviceInfo   bool
}

type hostAddrs struct {
	ipv4 []string
	ipv6 []string
}
