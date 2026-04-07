package nginx

import (
	"bytes"
	"html/template"
)

const defaultLoginCSS = `
:root {
  --bg: #0f172a;
  --card-bg: #1e293b;
  --border: #334155;
  --text: #e2e8f0;
  --text-muted: #94a3b8;
  --primary: #3b82f6;
  --primary-hover: #2563eb;
  --error: #ef4444;
  --error-bg: rgba(239, 68, 68, 0.1);
  --radius: 12px;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: var(--bg);
  color: var(--text);
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.login-card {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 2rem;
  width: 100%;
  max-width: 400px;
}

.login-card h1 {
  font-size: 1.25rem;
  margin-bottom: 0.5rem;
}

.login-card .domain {
  color: var(--text-muted);
  font-size: 0.875rem;
  font-family: monospace;
  margin-bottom: 1.5rem;
}

.login-card label {
  display: block;
  font-size: 0.875rem;
  color: var(--text-muted);
  margin-bottom: 0.375rem;
}

.login-card input[type="text"],
.login-card input[type="password"] {
  width: 100%;
  padding: 0.625rem 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  font-size: 0.9375rem;
  outline: none;
  transition: border-color 0.2s;
  margin-bottom: 0.75rem;
}

.login-card input[type="text"]:focus,
.login-card input[type="password"]:focus {
  border-color: var(--primary);
}

.login-card button {
  width: 100%;
  margin-top: 1rem;
  padding: 0.625rem;
  background: var(--primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.login-card button:hover {
  background: var(--primary-hover);
}

.error {
  background: var(--error-bg);
  color: var(--error);
  padding: 0.625rem 0.75rem;
  border-radius: 8px;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}
`

const loginPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Login — {{.Domain}}</title>
<style>{{.CSS}}</style>
</head>
<body>
<div class="login-card">
  <h1>Authorization Required</h1>
  <p class="domain">{{.Domain}}</p>
  {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
  <form method="POST" action="/_sc_auth/login">
    <input type="hidden" name="domain" value="{{.Domain}}">
    <input type="hidden" name="redirect" value="{{.Redirect}}">
    <label for="username">Username</label>
    <input type="text" id="username" name="username" required autofocus autocomplete="username">
    <label for="password">Password</label>
    <input type="password" id="password" name="password" required autocomplete="current-password">
    <button type="submit">Sign In</button>
  </form>
</div>
</body>
</html>`

var loginTmpl = template.Must(template.New("login").Parse(loginPageHTML))

type loginPageData struct {
	Domain   string
	CSS      template.CSS
	Redirect string
	Error    string
}

func renderLoginPage(domain, customCSS, redirect, errMsg string) ([]byte, error) {
	css := defaultLoginCSS
	if customCSS != "" {
		css = customCSS
	}
	data := loginPageData{
		Domain:   domain,
		CSS:      template.CSS(css),
		Redirect: redirect,
		Error:    errMsg,
	}
	var buf bytes.Buffer
	if err := loginTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
