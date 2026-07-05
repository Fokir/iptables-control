package nginx

import (
	"strings"
	"testing"
)

func sslDomain() *Domain {
	return &Domain{
		Domain:         "example.com",
		UpstreamIP:     "10.0.0.1",
		UpstreamPort:   8080,
		UpstreamScheme: "http",
		SSLEnabled:     true,
	}
}

func TestRenderConfig_HTTP2AlwaysEnabled(t *testing.T) {
	out, err := renderConfig(sslDomain(), 8080, false)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}
	if !strings.Contains(string(out), "listen 443 ssl http2;") {
		t.Errorf("expected 'listen 443 ssl http2;' in SSL config, got:\n%s", out)
	}
}

func TestRenderConfig_ForceTLS12(t *testing.T) {
	out, err := renderConfig(sslDomain(), 8080, true)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "ssl_protocols TLSv1.2;") {
		t.Errorf("expected 'ssl_protocols TLSv1.2;' when ForceTLS12, got:\n%s", s)
	}
	if strings.Contains(s, "options-ssl-nginx.conf") {
		t.Errorf("expected certbot include ABSENT when ForceTLS12, got:\n%s", s)
	}
}

func TestRenderConfig_DefaultUsesCertbotInclude(t *testing.T) {
	out, err := renderConfig(sslDomain(), 8080, false)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}
	if !strings.Contains(string(out), "options-ssl-nginx.conf") {
		t.Errorf("expected certbot include when not ForceTLS12, got:\n%s", out)
	}
}
