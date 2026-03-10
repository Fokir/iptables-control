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
	BasicAuthUser     string    `json:"basicAuthUser"`
	BasicAuthPassword string    `json:"basicAuthPassword"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateDomainRequest struct {
	Domain            string `json:"domain"`
	UpstreamIP        string `json:"upstreamIp"`
	UpstreamPort      int    `json:"upstreamPort"`
	UpstreamScheme    string `json:"upstreamScheme"`
	UpstreamSSLVerify *bool  `json:"upstreamSslVerify"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
}

type UpdateDomainRequest struct {
	Domain            string `json:"domain"`
	UpstreamIP        string `json:"upstreamIp"`
	UpstreamPort      int    `json:"upstreamPort"`
	UpstreamScheme    string `json:"upstreamScheme"`
	UpstreamSSLVerify *bool  `json:"upstreamSslVerify"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
}
