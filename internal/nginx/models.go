package nginx

import "time"

type Domain struct {
	ID                int64     `json:"id"`
	Domain            string    `json:"domain"`
	UpstreamIP        string    `json:"upstreamIp"`
	UpstreamPort      int       `json:"upstreamPort"`
	UpstreamScheme    string    `json:"upstreamScheme"`
	UpstreamSSLVerify bool      `json:"upstreamSslVerify"`
	SSLEnabled        bool      `json:"sslEnabled"`
	Enabled           bool      `json:"enabled"`
	AuthEnabled       bool      `json:"authEnabled"`
	AuthUsername      string    `json:"authUsername"`
	AuthPasswordHash  string    `json:"-"`
	AuthCookieSecret  string    `json:"-"`
	AuthCookieMaxAge  int       `json:"authCookieMaxAge"`
	AuthLoginCSS      string    `json:"authLoginCss"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateDomainRequest struct {
	Domain            string `json:"domain"`
	UpstreamIP        string `json:"upstreamIp"`
	UpstreamPort      int    `json:"upstreamPort"`
	UpstreamScheme    string `json:"upstreamScheme"`
	UpstreamSSLVerify *bool  `json:"upstreamSslVerify"`
	AuthEnabled       bool   `json:"authEnabled"`
	AuthUsername      string `json:"authUsername"`
	AuthPassword      string `json:"authPassword"`
	AuthCookieMaxAge  int    `json:"authCookieMaxAge"`
	AuthLoginCSS      string `json:"authLoginCss"`
}

type UpdateDomainRequest struct {
	Domain            string `json:"domain"`
	UpstreamIP        string `json:"upstreamIp"`
	UpstreamPort      int    `json:"upstreamPort"`
	UpstreamScheme    string `json:"upstreamScheme"`
	UpstreamSSLVerify *bool  `json:"upstreamSslVerify"`
	AuthEnabled       bool   `json:"authEnabled"`
	AuthUsername      string `json:"authUsername"`
	AuthPassword      string `json:"authPassword"`
	AuthCookieMaxAge  int    `json:"authCookieMaxAge"`
	AuthLoginCSS      string `json:"authLoginCss"`
}
