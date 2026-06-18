package util

import (
	"fmt"
	"strconv"
	"strings"

	"mdnsmap/internal/model"
)

func ParsePortSet(input string) (model.PortSet, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		value = "1-65535"
	}

	ports := model.PortSet{}
	for _, part := range strings.Split(value, ",") {
		if err := addPortPart(ports, part); err != nil {
			return nil, err
		}
	}
	return ports, nil
}

func addPortPart(ports model.PortSet, part string) error {
	start, end, err := parsePortPart(part)
	if err != nil {
		return err
	}
	for port := start; port <= end; port++ {
		ports[port] = struct{}{}
	}
	return nil
}

func parsePortPart(part string) (int, int, error) {
	part = strings.TrimSpace(part)
	if part == "" {
		return 0, 0, fmt.Errorf("empty port range item")
	}
	if !strings.Contains(part, "-") {
		port, err := parsePort(part)
		return port, port, err
	}
	return parsePortRange(part)
}

func parsePortRange(part string) (int, int, error) {
	limits := strings.Split(part, "-")
	if len(limits) != 2 {
		return 0, 0, fmt.Errorf("invalid port range: %s", part)
	}
	start, err := parsePort(limits[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := parsePort(limits[1])
	if err != nil {
		return 0, 0, err
	}
	if start > end {
		return 0, 0, fmt.Errorf("invalid port range %s: start is greater than end", part)
	}
	return start, end, nil
}

func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", value)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d", port)
	}
	return port, nil
}
