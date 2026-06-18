package service

import (
	"sort"

	"mdnsmap/internal/model"
)

func collect(results <-chan model.Asset, errs <-chan model.ScanError) model.ScanResult {
	result := model.ScanResult{}
	for results != nil || errs != nil {
		select {
		case asset, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			result.Assets = append(result.Assets, asset)
		case scanErr, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			result.Errors = append(result.Errors, scanErr)
		}
	}
	sort.Slice(result.Assets, func(i, j int) bool {
		return result.Assets[i].TargetIP < result.Assets[j].TargetIP
	})
	return result
}
