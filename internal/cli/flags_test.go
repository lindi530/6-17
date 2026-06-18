package cli

import "testing"

func TestParseEnablesMulticastByDefault(t *testing.T) {
	cfg, err := Parse([]string{"--cidr", "127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Multicast {
		t.Fatal("expected multicast to be enabled by default")
	}
}

func TestParseCanDisableMulticast(t *testing.T) {
	cfg, err := Parse([]string{"--cidr", "127.0.0.1", "--multicast=false"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Multicast {
		t.Fatal("expected multicast to be disabled")
	}
}
