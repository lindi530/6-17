package service

import (
	"sort"

	"mdnsmap/internal/model"
)

func sortAsset(asset *model.Asset) {
	sort.Slice(asset.Services, func(i, j int) bool {
		if asset.Services[i].Port == asset.Services[j].Port {
			return asset.Services[i].Type < asset.Services[j].Type
		}
		return asset.Services[i].Port < asset.Services[j].Port
	})
	sort.Slice(asset.DeviceInfo, func(i, j int) bool {
		return asset.DeviceInfo[i].Name < asset.DeviceInfo[j].Name
	})
	sort.Strings(asset.PTRAnswers)
}

func hasData(asset model.Asset) bool {
	return len(asset.Services) > 0 || len(asset.DeviceInfo) > 0 || len(asset.PTRAnswers) > 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func maxTTL(left, right uint32) uint32 {
	if right > left {
		return right
	}
	return left
}

func appendUnique(values []string, next ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range next {
		if _, ok := seen[value]; value != "" && !ok {
			values = append(values, value)
			seen[value] = struct{}{}
		}
	}
	return values
}
