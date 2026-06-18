package service

import (
	"context"
	"net"

	"github.com/miekg/dns"

	"mdnsmap/internal/model"
	"mdnsmap/internal/scanner"
)

func (s *ScanService) scanTarget(ctx context.Context, cfg model.ScanConfig, ip net.IP) (model.Asset, []model.ScanError) {
	client := scanner.NewClient(cfg.Timeout, cfg.Retries)
	asset := newAsset(ip)
	services := serviceQueries(cfg.Services)
	scanErrs := queryDiscovery(ctx, client, ip, cfg, &asset, &services)

	for _, name := range services {
		scanErrs = append(scanErrs, queryService(ctx, client, ip, name, cfg, &asset)...)
	}
	sortAsset(&asset)
	return asset, scanErrs
}

func newAsset(ip net.IP) model.Asset {
	asset := model.Asset{TargetIP: ip.String()}
	if ip.To4() != nil {
		asset.IPv4 = []string{ip.String()}
	}
	return asset
}

func queryDiscovery(ctx context.Context, client *scanner.Client, ip net.IP, cfg model.ScanConfig, asset *model.Asset, services *[]string) []model.ScanError {
	msg, err := client.Query(ctx, ip, scanner.DiscoveryService, dns.TypePTR)
	if err != nil {
		return scanError(ip, err)
	}
	parsed := scanner.ParseMessage(msg, scanner.DiscoveryService)
	asset.PTRAnswers = appendUnique(asset.PTRAnswers, parsed.PTRAnswers...)
	*services = appendUnique(*services, parsed.ServiceTypes...)
	mergeRecords(asset, cfg, parsed.Records)
	return nil
}

func queryService(ctx context.Context, client *scanner.Client, ip net.IP, name string, cfg model.ScanConfig, asset *model.Asset) []model.ScanError {
	msg, err := client.Query(ctx, ip, name, dns.TypePTR)
	if err != nil {
		return scanError(ip, err)
	}
	parsed := scanner.ParseMessage(msg, name)
	asset.PTRAnswers = appendUnique(asset.PTRAnswers, parsed.PTRAnswers...)
	mergeRecords(asset, cfg, parsed.Records)
	return nil
}

func scanError(ip net.IP, err error) []model.ScanError {
	if scanner.IsTimeout(err) {
		return nil
	}
	return []model.ScanError{{TargetIP: ip.String(), Err: err.Error()}}
}
