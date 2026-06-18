package service

import "mdnsmap/internal/model"

func upsertService(services []model.Service, next model.Service) []model.Service {
	for i, service := range services {
		if sameService(service, next) {
			services[i] = mergeService(service, next)
			return services
		}
	}
	return append(services, next)
}

func sameService(left, right model.Service) bool {
	return left.InstanceName == right.InstanceName && left.Port == right.Port && left.Type == right.Type
}

func mergeService(left, right model.Service) model.Service {
	left.Hostname = firstNonEmpty(left.Hostname, right.Hostname)
	left.IPv4 = appendUnique(left.IPv4, right.IPv4...)
	left.IPv6 = appendUnique(left.IPv6, right.IPv6...)
	left.TXT = appendUnique(left.TXT, right.TXT...)
	left.TTL = maxTTL(left.TTL, right.TTL)
	return left
}
