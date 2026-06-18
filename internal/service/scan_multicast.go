package service

import (
	"context"
	"net"
	"sort"

	"github.com/miekg/dns"

	"mdnsmap/internal/model"
	"mdnsmap/internal/scanner"
)

func (s *ScanService) scanMulticast(ctx context.Context, cfg model.ScanConfig) model.ScanResult {
	client := scanner.NewClient(cfg.Timeout, cfg.Retries)
	targets := targetSet(cfg.Targets)
	assets := map[string]model.Asset{}
	services := serviceQueries(cfg.Services)
	var scanErrs []model.ScanError

	scanErrs = append(scanErrs, queryMulticastDiscovery(ctx, client, cfg, targets, assets, &services)...)
	for _, name := range services {
		scanErrs = append(scanErrs, queryMulticastService(ctx, client, name, cfg, targets, assets)...)
	}
	return multicastScanResult(assets, scanErrs)
}

func queryMulticastDiscovery(ctx context.Context, client *scanner.Client, cfg model.ScanConfig, targets map[string]struct{}, assets map[string]model.Asset, services *[]string) []model.ScanError {
	responses, err := client.QueryMulticast(ctx, scanner.DiscoveryService, dns.TypePTR)
	if err != nil {
		return multicastScanError(err)
	}
	for _, response := range responses {
		applyMulticastResponse(response, scanner.DiscoveryService, cfg, targets, assets, services)
	}
	return nil
}

func queryMulticastService(ctx context.Context, client *scanner.Client, name string, cfg model.ScanConfig, targets map[string]struct{}, assets map[string]model.Asset) []model.ScanError {
	responses, err := client.QueryMulticast(ctx, name, dns.TypePTR)
	if err != nil {
		return multicastScanError(err)
	}
	for _, response := range responses {
		applyMulticastResponse(response, name, cfg, targets, assets, nil)
	}
	return nil
}

func applyMulticastResponse(response scanner.MulticastResponse, requested string, cfg model.ScanConfig, targets map[string]struct{}, assets map[string]model.Asset, services *[]string) {
	if !targetAllowed(response.From, targets) {
		return
	}
	ip := net.ParseIP(response.From)
	if ip == nil {
		return
	}
	asset := assets[response.From]
	if asset.TargetIP == "" {
		asset = newAsset(ip)
	}
	parsed := scanner.ParseMessage(response.Message, requested)
	asset.PTRAnswers = appendUnique(asset.PTRAnswers, parsed.PTRAnswers...)
	if services != nil {
		*services = appendUnique(*services, parsed.ServiceTypes...)
	}
	mergeRecords(&asset, cfg, parsed.Records)
	assets[response.From] = asset
}

func multicastScanError(err error) []model.ScanError {
	if scanner.IsTimeout(err) {
		return nil
	}
	return []model.ScanError{{TargetIP: scanner.MDNSIPv4Multicast, Err: err.Error()}}
}

func targetSet(targets []net.IP) map[string]struct{} {
	result := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		result[target.String()] = struct{}{}
	}
	return result
}

func targetAllowed(source string, targets map[string]struct{}) bool {
	if len(targets) == 0 {
		return true
	}
	_, ok := targets[source]
	return ok
}

func multicastScanResult(assets map[string]model.Asset, errs []model.ScanError) model.ScanResult {
	result := model.ScanResult{Errors: errs}
	for _, asset := range assets {
		if !hasData(asset) {
			continue
		}
		sortAsset(&asset)
		result.Assets = append(result.Assets, asset)
	}
	sort.Slice(result.Assets, func(i, j int) bool {
		return result.Assets[i].TargetIP < result.Assets[j].TargetIP
	})
	return result
}
