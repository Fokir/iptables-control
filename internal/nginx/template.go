package nginx

import (
	"bytes"
	"text/template"
)

// locationBlock is a reusable sub-template for proxy location directives.
const locationBlock = `
    # Determine the real client-facing scheme. When behind a proxy that
    # terminates TLS (e.g. Cloudflare), the origin connection may be plain
    # HTTP, so $scheme is unreliable. Trust X-Forwarded-Proto when present;
    # for direct connections the header is absent and we fall back to $scheme.
    set $real_scheme $scheme;
    if ($http_x_forwarded_proto = "https") {
        set $real_scheme "https";
    }
{{- if .HasAuth}}
    # Cookie-based auth via system-control
    location = /_sc_auth_verify {
        internal;
        proxy_pass http://127.0.0.1:{{.BackendPort}}/api/domain-auth/verify;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
        proxy_set_header X-Original-Host $host;
        proxy_set_header X-Original-URI $request_uri;
        proxy_set_header Cookie $http_cookie;
    }

    location /_sc_auth/ {
        proxy_pass http://127.0.0.1:{{.BackendPort}}/api/domain-auth/;
        proxy_set_header Host $host;
        proxy_set_header X-Original-Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
{{- end}}

    location / {
{{- if .HasAuth}}
        auth_request /_sc_auth_verify;
        error_page 401 = @login_redirect;
{{- end}}
        proxy_pass {{.UpstreamScheme}}://{{.UpstreamIP}}:{{.UpstreamPort}};
{{- if not .UpstreamSSLVerify}}
        proxy_ssl_verify off;
{{- end}}
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $real_scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # WebSocket / WebRTC: long-lived connections
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
        proxy_connect_timeout 60s;
        proxy_buffering off;
    }

{{- if .HasAuth}}
    location @login_redirect {
        return 302 $real_scheme://$host/_sc_auth/login?domain=$host&redirect=$real_scheme://$host$request_uri;
    }
{{- end}}`

const nginxConfTemplate = `# Managed by system-control. Do not edit manually.
{{- if .SSLEnabled}}
server {
    listen 80;
    server_name {{.Domain}};
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name {{.Domain}};

    ssl_certificate /etc/letsencrypt/live/{{.Domain}}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/{{.Domain}}/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
{{template "locations" .}}
}
{{- else}}
server {
    listen 80;
    server_name {{.Domain}};
{{template "locations" .}}
}
{{- end}}
`

var confTmpl = template.Must(
	template.Must(template.New("locations").Parse(locationBlock)).
		New("nginx").Parse(nginxConfTemplate),
)

type configData struct {
	Domain            string
	UpstreamIP        string
	UpstreamPort      int
	UpstreamScheme    string
	UpstreamSSLVerify bool
	HasAuth           bool
	BackendPort       int
	SSLEnabled        bool
}

func renderConfig(d *Domain, backendPort int) ([]byte, error) {
	scheme := d.UpstreamScheme
	if scheme == "" {
		scheme = "http"
	}
	data := configData{
		Domain:            d.Domain,
		UpstreamIP:        d.UpstreamIP,
		UpstreamPort:      d.UpstreamPort,
		UpstreamScheme:    scheme,
		UpstreamSSLVerify: d.UpstreamSSLVerify,
		HasAuth:           d.AuthEnabled,
		BackendPort:       backendPort,
		SSLEnabled:        d.SSLEnabled,
	}
	var buf bytes.Buffer
	if err := confTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
