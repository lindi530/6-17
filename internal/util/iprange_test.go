package util

import "testing"

func TestParseTargetsSingleIP(t *testing.T) {
	targets, err := ParseTargets("192.168.1.10")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].String() != "192.168.1.10" {
		t.Fatalf("unexpected targets: %v", targets)
	}
}

func TestParseTargetsCIDRSkipsNetworkAndBroadcast(t *testing.T) {
	targets, err := ParseTargets("192.168.1.0/30")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(targets); got != 2 {
		t.Fatalf("expected 2 hosts, got %d", got)
	}
	if targets[0].String() != "192.168.1.1" || targets[1].String() != "192.168.1.2" {
		t.Fatalf("unexpected targets: %v", targets)
	}
}
