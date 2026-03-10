package validate

import (
	"fmt"
	"net"
	"regexp"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func IP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

func Port(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

func Domain(domain string) error {
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain name: %s", domain)
	}
	return nil
}

func Required(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}
