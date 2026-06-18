package service

import (
	"net"
	"sort"
	"strings"

	"mdnsmap/internal/model"
	"mdnsmap/internal/scanner"
)

func mergeResults(results ...model.ScanResult) model.ScanResult {
	assets := map[string]model.Asset{}
	merged := model.ScanResult{}
	for _, result := range results {
		for _, asset := range result.Assets {
			current := assets[asset.TargetIP]
			mergeAsset(&current, asset)
			assets[asset.TargetIP] = current
		}
		merged.Errors = append(merged.Errors, result.Errors...)
	}
	for _, asset := range assets {
		sortAsset(&asset)
		merged.Assets = append(merged.Assets, asset)
	}
	sort.Slice(merged.Assets, func(i, j int) bool {
		return merged.Assets[i].TargetIP < merged.Assets[j].TargetIP
	})
	return merged
}

func mergeAsset(asset *model.Asset, next model.Asset) {
	asset.TargetIP = firstNonEmpty(asset.TargetIP, next.TargetIP)
	asset.Hostname = firstNonEmpty(asset.Hostname, next.Hostname)
	asset.IPv4 = appendUnique(asset.IPv4, next.IPv4...)
	asset.IPv6 = appendUnique(asset.IPv6, next.IPv6...)
	asset.TTL = maxTTL(asset.TTL, next.TTL)
	asset.PTRAnswers = appendUnique(asset.PTRAnswers, next.PTRAnswers...)
	for _, service := range next.Services {
		asset.Services = upsertService(asset.Services, service)
	}
	for _, service := range next.DeviceInfo {
		asset.DeviceInfo = upsertService(asset.DeviceInfo, service)
	}
}

func mergeRecords(asset *model.Asset, cfg model.ScanConfig, records []scanner.ServiceRecord) {
	for _, record := range records {
		mergeRecord(asset, cfg, record)
	}
}

func mergeRecord(asset *model.Asset, cfg model.ScanConfig, record scanner.ServiceRecord) {
	service := toService(record, asset.TargetIP)
	if record.DeviceInfo && !record.HasSRV {
		asset.DeviceInfo = upsertService(asset.DeviceInfo, service)
		return
	}
	if record.HasSRV && cfg.Ports.Contains(int(record.Port)) {
		asset.Services = upsertService(asset.Services, service)
	}
	updateAssetIdentity(asset, service)
}

func updateAssetIdentity(asset *model.Asset, service model.Service) {
	asset.Hostname = firstNonEmpty(asset.Hostname, service.Hostname)
	asset.IPv4 = appendUnique(asset.IPv4, service.IPv4...)
	asset.IPv6 = appendUnique(asset.IPv6, service.IPv6...)
	asset.TTL = maxTTL(asset.TTL, service.TTL)
}

func toService(record scanner.ServiceRecord, targetIP string) model.Service {
	return model.Service{
		Port:         record.Port,
		Proto:        record.Proto,
		Type:         scanner.ServiceLabel(record.Type),
		InstanceName: record.InstanceName,
		Name:         record.Name,
		Hostname:     serviceHostname(record),
		IPv4:         serviceIPv4(record, targetIP),
		IPv6:         record.IPv6,
		TTL:          record.TTL,
		TXT:          record.TXT,
		DeviceInfo:   record.DeviceInfo,
	}
}

func serviceIPv4(record scanner.ServiceRecord, targetIP string) []string {
	if len(record.IPv4) > 0 || net.ParseIP(targetIP).To4() == nil {
		return record.IPv4
	}
	return []string{targetIP}
}

func serviceHostname(record scanner.ServiceRecord) string {
	if record.Hostname != "" || !record.DeviceInfo {
		return record.Hostname
	}
	return deviceInfoHostname(record.Name)
}

func deviceInfoHostname(name string) string {
	host := name
	if idx := strings.Index(host, "("); idx > 0 {
		host = strings.TrimSpace(host[:idx])
	}
	if host == "" || strings.Contains(host, ".") {
		return host
	}
	return host + ".local"
}
