package output

import (
	"bytes"
	"strings"
	"testing"

	"mdnsmap/internal/model"
)

func TestWriteTextExampleShape(t *testing.T) {
	result := model.ScanResult{Assets: []model.Asset{{
		PTRAnswers: []string{"_qdiscover._tcp.local."},
		Services: []model.Service{{
			Port: 5000, Proto: "tcp", Type: "qdiscover", Name: "slw-nas",
			IPv4: []string{"192.168.1.20"}, IPv6: []string{"fe80::1"},
			Hostname: "slw-nas.local", TTL: 10,
			TXT: []string{"accessType=https", "accessPort=86", "model=TS-X64"},
		}},
	}}}
	var buf bytes.Buffer
	if err := WriteText(&buf, result); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"services:", "5000/tcp qdiscover:", "Name=slw-nas",
		"IPv4=192.168.1.20", "Hostname=slw-nas.local",
		"accessType=https,accessPort=86,model=TS-X64", "answers:", "PTR:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestWriteTextEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, model.ScanResult{}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "未发现 mDNS 资产" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
