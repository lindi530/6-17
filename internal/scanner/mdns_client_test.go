package scanner

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestClientQueryReadsUnicastResponse(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	port := conn.LocalAddr().(*net.UDPAddr).Port
	go func() {
		buf := make([]byte, 65535)
		n, addr, readErr := conn.ReadFrom(buf)
		if readErr != nil {
			return
		}
		query := new(dns.Msg)
		if query.Unpack(buf[:n]) != nil {
			return
		}
		resp := new(dns.Msg)
		resp.SetReply(query)
		resp.Answer = []dns.RR{&dns.PTR{
			Hdr: dns.RR_Header{Name: DiscoveryService, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 10},
			Ptr: "_http._tcp.local.",
		}}
		packet, _ := resp.Pack()
		_, _ = conn.WriteTo(packet, addr)
	}()

	client := &Client{Timeout: time.Second, Retries: 0, Port: port}
	msg, err := client.Query(context.Background(), net.ParseIP("127.0.0.1"), DiscoveryService, dns.TypePTR)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.Answer) != 1 {
		t.Fatalf("expected one answer, got %d", len(msg.Answer))
	}
}

func TestClientQueryMulticastCollectsResponses(t *testing.T) {
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	port := conn.LocalAddr().(*net.UDPAddr).Port
	go func() {
		buf := make([]byte, 65535)
		n, addr, readErr := conn.ReadFrom(buf)
		if readErr != nil {
			return
		}
		query := new(dns.Msg)
		if query.Unpack(buf[:n]) != nil {
			return
		}
		resp := new(dns.Msg)
		resp.SetReply(query)
		resp.Answer = []dns.RR{&dns.PTR{
			Hdr: dns.RR_Header{Name: DiscoveryService, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 10},
			Ptr: "_http._tcp.local.",
		}}
		packet, _ := resp.Pack()
		_, _ = conn.WriteTo(packet, addr)
	}()

	client := &Client{
		Timeout:     100 * time.Millisecond,
		Retries:     0,
		Port:        port,
		MulticastIP: net.ParseIP("127.0.0.1"),
	}
	responses, err := client.QueryMulticast(context.Background(), DiscoveryService, dns.TypePTR)
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected one response, got %d", len(responses))
	}
	if responses[0].From != "127.0.0.1" {
		t.Fatalf("unexpected response source: %s", responses[0].From)
	}
	if len(responses[0].Message.Answer) != 1 {
		t.Fatalf("expected one answer, got %d", len(responses[0].Message.Answer))
	}
}
