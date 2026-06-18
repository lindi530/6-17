package scanner

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

func TestParseMessageKeepsDeepTXTBanner(t *testing.T) {
	msg := new(dns.Msg)
	msg.Answer = []dns.RR{
		ptr("_qdiscover._tcp.local.", "slw-nas._qdiscover._tcp.local.", 10),
		srv("slw-nas._qdiscover._tcp.local.", "slw-nas.local.", 5000, 10),
		txt("slw-nas._qdiscover._tcp.local.", []string{
			"accessType=https", "accessPort=86", "model=TS-X64",
			"displayModel=TS-464C", "fwVer=5.2.9", "fwBuildNum=20260214",
		}, 10),
		a("slw-nas.local.", "192.168.1.20", 10),
		aaaa("slw-nas.local.", "fe80::265e:beff:fe69:a313", 10),
	}

	parsed := ParseMessage(msg, "_qdiscover._tcp.local.")
	if len(parsed.Records) != 1 {
		t.Fatalf("expected one record, got %d", len(parsed.Records))
	}
	record := parsed.Records[0]
	if record.Port != 5000 || record.Name != "slw-nas" || !record.HasSRV {
		t.Fatalf("unexpected service record: %+v", record)
	}
	if len(record.TXT) != 6 || record.TXT[5] != "fwBuildNum=20260214" {
		t.Fatalf("unexpected txt banner: %v", record.TXT)
	}
	if record.IPv4[0] != "192.168.1.20" || record.IPv6[0] != "fe80::265e:beff:fe69:a313" {
		t.Fatalf("unexpected addresses: %+v", record)
	}
}

func TestParseMessageFindsServiceTypes(t *testing.T) {
	msg := new(dns.Msg)
	msg.Answer = []dns.RR{ptr(DiscoveryService, "_http._tcp.local.", 10)}

	parsed := ParseMessage(msg, DiscoveryService)
	if len(parsed.ServiceTypes) != 1 || parsed.ServiceTypes[0] != "_http._tcp.local." {
		t.Fatalf("unexpected service types: %v", parsed.ServiceTypes)
	}
}

func ptr(name, target string, ttl uint32) *dns.PTR {
	return &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: ttl}, Ptr: target}
}

func srv(name, target string, port uint16, ttl uint32) *dns.SRV {
	return &dns.SRV{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: ttl}, Target: target, Port: port}
}

func txt(name string, values []string, ttl uint32) *dns.TXT {
	return &dns.TXT{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: ttl}, Txt: values}
}

func a(name, ip string, ttl uint32) *dns.A {
	return &dns.A{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}, A: net.ParseIP(ip)}
}

func aaaa(name, ip string, ttl uint32) *dns.AAAA {
	return &dns.AAAA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}, AAAA: net.ParseIP(ip)}
}
