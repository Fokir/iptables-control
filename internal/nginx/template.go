package nginx

import (
	"bytes"
	"text/template"
)

const nginxConfTemplate = `# Managed by system-control. Do not edit manually.
server {
    listen 80;
    server_name {{.Domain}};

    location / {
{{- if .HasBasicAuth}}
        auth_basic "Restricted";
        auth_basic_user_file {{.HtpasswdPath}};
{{- end}}
        proxy_pass http://{{.UpstreamIP}}:{{.UpstreamPort}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`

var confTmpl = template.Must(template.New("nginx").Parse(nginxConfTemplate))

type configData struct {
	Domain       string
	UpstreamIP   string
	UpstreamPort int
	HasBasicAuth bool
	HtpasswdPath string
}

func renderConfig(d *Domain, htpasswdPath string) ([]byte, error) {
	data := configData{
		Domain:       d.Domain,
		UpstreamIP:   d.UpstreamIP,
		UpstreamPort: d.UpstreamPort,
		HasBasicAuth: d.BasicAuthUser != "" && d.BasicAuthPassword != "",
		HtpasswdPath: htpasswdPath,
	}
	var buf bytes.Buffer
	if err := confTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
