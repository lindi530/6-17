package service

import (
	"testing"

	"mdnsmap/internal/model"
)

func TestMergeResultsMergesAssetsByTargetIP(t *testing.T) {
	multicast := model.ScanResult{Assets: []model.Asset{{
		TargetIP:   "192.168.1.20",
		IPv4:       []string{"192.168.1.20"},
		PTRAnswers: []string{"_http._tcp.local."},
		Services: []model.Service{{
			Port: 80, Proto: "tcp", Type: "http", Name: "web",
			Hostname: "web.local", TXT: []string{"path=/"},
		}},
	}}}
	unicast := model.ScanResult{Assets: []model.Asset{{
		TargetIP:   "192.168.1.20",
		Hostname:   "web.local",
		PTRAnswers: []string{"_device-info._tcp.local."},
		Services: []model.Service{{
			Port: 80, Proto: "tcp", Type: "http", Name: "web",
			IPv4: []string{"192.168.1.20"}, TTL: 120,
		}},
		DeviceInfo: []model.Service{{Type: "device-info", Name: "web"}},
	}}}

	result := mergeResults(multicast, unicast)
	if len(result.Assets) != 1 {
		t.Fatalf("expected one merged asset, got %d", len(result.Assets))
	}
	asset := result.Assets[0]
	if asset.Hostname != "web.local" || len(asset.PTRAnswers) != 2 {
		t.Fatalf("unexpected merged asset: %+v", asset)
	}
	if len(asset.Services) != 1 || asset.Services[0].TTL != 120 || len(asset.Services[0].TXT) != 1 {
		t.Fatalf("unexpected merged service: %+v", asset.Services)
	}
	if len(asset.DeviceInfo) != 1 {
		t.Fatalf("expected device-info to be merged, got %+v", asset.DeviceInfo)
	}
}
