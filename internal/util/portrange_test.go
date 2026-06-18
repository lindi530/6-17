package util

import "testing"

func TestParsePortSet(t *testing.T) {
	ports, err := ParsePortSet("80,443,5000-5002")
	if err != nil {
		t.Fatal(err)
	}
	for _, port := range []int{80, 443, 5000, 5001, 5002} {
		if !ports.Contains(port) {
			t.Fatalf("expected port %d", port)
		}
	}
	if ports.Contains(22) {
		t.Fatal("did not expect port 22")
	}
}

func TestParsePortSetRejectsInvalidRange(t *testing.T) {
	if _, err := ParsePortSet("90-80"); err == nil {
		t.Fatal("expected invalid range error")
	}
}
