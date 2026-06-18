package output

import (
	"fmt"
	"io"
	"strings"

	"mdnsmap/internal/model"
)

func WriteText(w io.Writer, result model.ScanResult) error {
	if len(result.Assets) == 0 {
		_, err := fmt.Fprintln(w, "未发现 mDNS 资产")
		return err
	}

	for i, asset := range result.Assets {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeAsset(w, asset); err != nil {
			return err
		}
	}
	return nil
}

func writeAsset(w io.Writer, asset model.Asset) error {
	if _, err := fmt.Fprintln(w, "services:"); err != nil {
		return err
	}
	if err := writeServices(w, asset.Services); err != nil {
		return err
	}
	if err := writeDeviceInfo(w, asset.DeviceInfo); err != nil {
		return err
	}
	return writeAnswers(w, asset.PTRAnswers)
}

func writeServices(w io.Writer, services []model.Service) error {
	for _, service := range services {
		if _, err := fmt.Fprintf(w, "%d/%s %s:\n", service.Port, service.Proto, service.Type); err != nil {
			return err
		}
		if err := writeServiceBody(w, service); err != nil {
			return err
		}
	}
	return nil
}

func writeDeviceInfo(w io.Writer, services []model.Service) error {
	if len(services) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "device-info:"); err != nil {
		return err
	}
	for _, service := range services {
		if err := writeServiceBody(w, service); err != nil {
			return err
		}
	}
	return nil
}

func writeServiceBody(w io.Writer, service model.Service) error {
	for _, line := range serviceLines(service) {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func serviceLines(service model.Service) []string {
	lines := []string{
		"Name=" + service.Name,
		"IPv4=" + strings.Join(service.IPv4, ","),
		"IPv6=" + strings.Join(service.IPv6, ","),
		"Hostname=" + service.Hostname,
		fmt.Sprintf("TTL=%d", service.TTL),
	}
	if len(service.TXT) > 0 {
		lines = append(lines, strings.Join(service.TXT, ","))
	}
	return lines
}

func writeAnswers(w io.Writer, answers []string) error {
	if len(answers) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "answers:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "PTR:"); err != nil {
		return err
	}
	for _, answer := range answers {
		if _, err := fmt.Fprintln(w, strings.TrimSuffix(answer, ".")); err != nil {
			return err
		}
	}
	return nil
}
