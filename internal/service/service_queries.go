package service

import "github.com/miekg/dns"

func serviceQueries(extra []string) []string {
	queries := append([]string{}, defaultServices...)
	for _, service := range extra {
		queries = appendUnique(queries, dns.Fqdn(service))
	}
	return queries
}
