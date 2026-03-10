package nginx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ExternalDomain represents a nginx domain parsed from an external config file.
type ExternalDomain struct {
	Filename     string `json:"filename"`
	Domain       string `json:"domain"`
	UpstreamIP   string `json:"upstreamIp"`
	UpstreamPort int    `json:"upstreamPort"`
	SSLEnabled   bool   `json:"sslEnabled"`
	HasBasicAuth bool   `json:"hasBasicAuth"`
}

var (
	reServerName = regexp.MustCompile(`(?m)^\s*server_name\s+([^\s;]+)`)
	reProxyPass  = regexp.MustCompile(`(?m)^\s*proxy_pass\s+https?://([^:/\s]+):(\d+)`)
	reProxyHost  = regexp.MustCompile(`(?m)^\s*proxy_pass\s+https?://([^:/;\s]+)\s*;`)
	reSSL        = regexp.MustCompile(`(?m)^\s*listen\s+.*\bssl\b`)
	reBasicAuth  = regexp.MustCompile(`(?m)^\s*auth_basic\s+"[^"]*"`)
)

func parseExternalConfig(content string) (ExternalDomain, error) {
	var d ExternalDomain

	if m := reServerName.FindStringSubmatch(content); len(m) > 1 {
		d.Domain = m[1]
	} else {
		return d, fmt.Errorf("server_name not found")
	}

	if m := reProxyPass.FindStringSubmatch(content); len(m) > 2 {
		d.UpstreamIP = m[1]
		port, err := strconv.Atoi(m[2])
		if err != nil {
			return d, fmt.Errorf("invalid proxy_pass port: %w", err)
		}
		d.UpstreamPort = port
	} else if m := reProxyHost.FindStringSubmatch(content); len(m) > 1 {
		d.UpstreamIP = m[1]
		if strings.Contains(content, "proxy_pass https://") {
			d.UpstreamPort = 443
		} else {
			d.UpstreamPort = 80
		}
	}

	d.SSLEnabled = reSSL.MatchString(content)
	d.HasBasicAuth = reBasicAuth.MatchString(content)

	return d, nil
}
