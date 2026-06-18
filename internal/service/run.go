package service

import (
	"context"
	"net"
	"sync"

	"mdnsmap/internal/model"
)

func (s *ScanService) Run(ctx context.Context, cfg model.ScanConfig) model.ScanResult {
	result := model.ScanResult{}
	if cfg.Multicast {
		result = s.scanMulticast(ctx, cfg)
	}
	return mergeResults(result, s.runUnicast(ctx, cfg))
}

func (s *ScanService) runUnicast(ctx context.Context, cfg model.ScanConfig) model.ScanResult {
	jobs := make(chan net.IP)
	results := make(chan model.Asset)
	errs := make(chan model.ScanError)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go s.scanWorker(ctx, cfg, jobs, results, errs, &wg)
	}
	go closeWhenDone(jobs, results, errs, &wg, cfg.Targets)
	return collect(results, errs)
}

func (s *ScanService) scanWorker(ctx context.Context, cfg model.ScanConfig, jobs <-chan net.IP, results chan<- model.Asset, errs chan<- model.ScanError, wg *sync.WaitGroup) {
	defer wg.Done()
	for ip := range jobs {
		asset, scanErrs := s.scanTarget(ctx, cfg, ip)
		if hasData(asset) {
			results <- asset
		}
		for _, scanErr := range scanErrs {
			errs <- scanErr
		}
	}
}

func closeWhenDone(jobs chan<- net.IP, results chan<- model.Asset, errs chan<- model.ScanError, wg *sync.WaitGroup, targets []net.IP) {
	for _, target := range targets {
		jobs <- target
	}
	close(jobs)
	wg.Wait()
	close(results)
	close(errs)
}
