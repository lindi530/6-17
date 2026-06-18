package scanner

import (
	"strings"

	"github.com/miekg/dns"
)

func ParseMessage(msg *dns.Msg, requested string) ParsedResponse {
	records := map[string]*ServiceRecord{}
	addresses := map[string]*hostAddrs{}
	result := ParsedResponse{}

	for _, rr := range allRecords(msg) {
		applyRecord(rr, requested, records, addresses, &result)
	}
	for _, record := range records {
		applyRecordAddresses(record, addresses)
		result.Records = append(result.Records, *record)
	}
	return result
}

func applyRecord(rr dns.RR, requested string, records map[string]*ServiceRecord, addresses map[string]*hostAddrs, result *ParsedResponse) {
	switch rec := rr.(type) {
	case *dns.PTR:
		parsePTR(rec, records, result)
	case *dns.SRV:
		entry := ensureRecord(records, rec.Hdr.Name, requested)
		entry.Hostname = cleanName(rec.Target)
		entry.Port = rec.Port
		entry.TTL = maxTTL(entry.TTL, rec.Hdr.Ttl)
		entry.HasSRV = true
	case *dns.TXT:
		entry := ensureRecord(records, rec.Hdr.Name, requested)
		entry.TXT = appendUnique(entry.TXT, rec.Txt...)
		entry.TTL = maxTTL(entry.TTL, rec.Hdr.Ttl)
	case *dns.A:
		addrs := ensureAddrs(addresses, cleanName(rec.Hdr.Name))
		addrs.ipv4 = appendUnique(addrs.ipv4, rec.A.String())
	case *dns.AAAA:
		addrs := ensureAddrs(addresses, cleanName(rec.Hdr.Name))
		addrs.ipv6 = appendUnique(addrs.ipv6, rec.AAAA.String())
	}
}

func parsePTR(ptr *dns.PTR, records map[string]*ServiceRecord, result *ParsedResponse) {
	name := dns.Fqdn(ptr.Hdr.Name)
	target := dns.Fqdn(ptr.Ptr)
	if strings.EqualFold(name, DiscoveryService) {
		result.ServiceTypes = appendUnique(result.ServiceTypes, target)
		result.PTRAnswers = appendUnique(result.PTRAnswers, target)
		return
	}
	entry := ensureRecord(records, target, name)
	entry.Type = name
	entry.InstanceName = cleanName(target)
	entry.Name = instanceLabel(target, name)
	entry.TTL = maxTTL(entry.TTL, ptr.Hdr.Ttl)
	entry.DeviceInfo = isDeviceInfo(name)
	result.PTRAnswers = appendUnique(result.PTRAnswers, name)
}

func ensureRecord(records map[string]*ServiceRecord, instance, fallbackType string) *ServiceRecord {
	key := dns.Fqdn(instance)
	if record, ok := records[key]; ok {
		return record
	}
	records[key] = newServiceRecord(key, fallbackType)
	return records[key]
}

func newServiceRecord(instance, fallbackType string) *ServiceRecord {
	serviceType := inferType(instance, fallbackType)
	return &ServiceRecord{
		Type:         serviceType,
		InstanceName: cleanName(instance),
		Name:         instanceLabel(instance, serviceType),
		Proto:        protoOf(serviceType),
		DeviceInfo:   isDeviceInfo(serviceType),
	}
}

func allRecords(msg *dns.Msg) []dns.RR {
	records := make([]dns.RR, 0, len(msg.Answer)+len(msg.Ns)+len(msg.Extra))
	records = append(records, msg.Answer...)
	records = append(records, msg.Ns...)
	records = append(records, msg.Extra...)
	return records
}

func ensureAddrs(addresses map[string]*hostAddrs, host string) *hostAddrs {
	if addrs, ok := addresses[host]; ok {
		return addrs
	}
	addresses[host] = &hostAddrs{}
	return addresses[host]
}

func applyRecordAddresses(record *ServiceRecord, addresses map[string]*hostAddrs) {
	addrs, ok := addresses[record.Hostname]
	if !ok {
		return
	}
	record.IPv4 = appendUnique(record.IPv4, addrs.ipv4...)
	record.IPv6 = appendUnique(record.IPv6, addrs.ipv6...)
}
