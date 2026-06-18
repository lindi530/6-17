package scanner

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

type Client struct {
	Timeout     time.Duration
	Retries     int
	Port        int
	MulticastIP net.IP
}

func NewClient(timeout time.Duration, retries int) *Client {
	return &Client{Timeout: timeout, Retries: retries, Port: MDNSPort, MulticastIP: net.ParseIP(MDNSIPv4Multicast)}
}

func (c *Client) Query(ctx context.Context, target net.IP, name string, qtype uint16) (*dns.Msg, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		msg, err := c.queryOnce(ctx, target, name, qtype)
		if err == nil {
			return msg, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (c *Client) QueryMulticast(ctx context.Context, name string, qtype uint16) ([]MulticastResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		responses, err := c.queryMulticastOnce(ctx, name, qtype)
		if err == nil || len(responses) > 0 {
			return responses, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (c *Client) queryOnce(ctx context.Context, target net.IP, name string, qtype uint16) (*dns.Msg, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", c.address(target))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}
	if err := writeQuery(conn, name, qtype); err != nil {
		return nil, err
	}
	return readResponse(conn)
}

func (c *Client) queryMulticastOnce(ctx context.Context, name string, qtype uint16) ([]MulticastResponse, error) {
	conn, err := (&net.ListenConfig{}).ListenPacket(ctx, "udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}
	if err := writeQueryTo(conn, c.multicastAddress(), name, qtype); err != nil {
		return nil, err
	}
	return readMulticastResponses(conn)
}

func (c *Client) address(target net.IP) string {
	return net.JoinHostPort(target.String(), strconv.Itoa(c.port()))
}

func (c *Client) multicastAddress() *net.UDPAddr {
	ip := c.MulticastIP
	if ip == nil {
		ip = net.ParseIP(MDNSIPv4Multicast)
	}
	return &net.UDPAddr{IP: ip, Port: c.port()}
}

func (c *Client) port() int {
	port := c.Port
	if port == 0 {
		port = MDNSPort
	}
	return port
}

func writeQuery(conn net.Conn, name string, qtype uint16) error {
	packet, err := buildQueryPacket(name, qtype)
	if err != nil {
		return err
	}
	_, err = conn.Write(packet)
	return err
}

func writeQueryTo(conn net.PacketConn, addr net.Addr, name string, qtype uint16) error {
	packet, err := buildQueryPacket(name, qtype)
	if err != nil {
		return err
	}
	_, err = conn.WriteTo(packet, addr)
	return err
}

func readResponse(conn net.Conn) (*dns.Msg, error) {
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := new(dns.Msg)
	if err := resp.Unpack(buf[:n]); err != nil {
		return nil, err
	}
	return resp, nil
}

func readMulticastResponses(conn net.PacketConn) ([]MulticastResponse, error) {
	var responses []MulticastResponse
	for {
		response, err := readMulticastResponse(conn)
		if err != nil {
			if len(responses) > 0 && IsTimeout(err) {
				return responses, nil
			}
			return responses, err
		}
		responses = append(responses, response)
	}
}

func readMulticastResponse(conn net.PacketConn) (MulticastResponse, error) {
	buf := make([]byte, 65535)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return MulticastResponse{}, err
	}
	resp := new(dns.Msg)
	if err := resp.Unpack(buf[:n]); err != nil {
		return MulticastResponse{}, err
	}
	return MulticastResponse{From: responseSource(addr), Message: resp}, nil
}

func responseSource(addr net.Addr) string {
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return addr.String()
	}
	return udpAddr.IP.String()
}

func buildQueryPacket(name string, qtype uint16) ([]byte, error) {
	return buildQuery(name, qtype).Pack()
}

func buildQuery(name string, qtype uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(name), qtype)
	msg.RecursionDesired = false
	msg.Question[0].Qclass = dns.ClassINET | 0x8000
	return msg
}

func IsTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
