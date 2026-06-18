package cli

import (
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"mdnsmap/internal/model"
	"mdnsmap/internal/util"
)

type serviceFlags []string

type flagValues struct {
	cidr      string
	ports     string
	timeout   time.Duration
	workers   int
	retries   int
	services  serviceFlags
	multicast bool
}

func Parse(args []string) (model.ScanConfig, error) {
	values := flagValues{}
	if err := newFlagSet(&values).Parse(args); err != nil {
		return model.ScanConfig{}, err
	}
	targets, err := util.ParseTargets(values.cidr)
	if err != nil {
		return model.ScanConfig{}, err
	}
	ports, err := util.ParsePortSet(values.ports)
	if err != nil {
		return model.ScanConfig{}, err
	}
	if err := validate(values); err != nil {
		return model.ScanConfig{}, err
	}
	return buildConfig(values, targets, ports), nil
}

func newFlagSet(values *flagValues) *flag.FlagSet {
	fs := flag.NewFlagSet("mdnsmap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&values.cidr, "cidr", "", "IPv4 CIDR or single IP to scan")
	fs.StringVar(&values.ports, "ports", "1-65535", "port filter, e.g. 80,443,5000-6000")
	fs.DurationVar(&values.timeout, "timeout", 2*time.Second, "per query timeout")
	fs.IntVar(&values.workers, "workers", 128, "concurrent target workers")
	fs.IntVar(&values.retries, "retries", 1, "retry count per mDNS query")
	fs.Var(&values.services, "service", "extra mDNS service type, repeatable")
	fs.BoolVar(&values.multicast, "multicast", true, "send multicast mDNS queries before unicast scan")
	return fs
}

func validate(values flagValues) error {
	if values.timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than zero")
	}
	if values.workers <= 0 {
		return fmt.Errorf("--workers must be greater than zero")
	}
	if values.retries < 0 {
		return fmt.Errorf("--retries must be zero or greater")
	}
	return nil
}

func buildConfig(values flagValues, targets []net.IP, ports model.PortSet) model.ScanConfig {
	return model.ScanConfig{
		Targets:   targets,
		Ports:     ports,
		Timeout:   values.timeout,
		Workers:   values.workers,
		Retries:   values.retries,
		Services:  []string(values.services),
		Multicast: values.multicast,
	}
}

func (s *serviceFlags) String() string {
	return fmt.Sprint([]string(*s))
}

func (s *serviceFlags) Set(value string) error {
	*s = append(*s, value)
	return nil
}
