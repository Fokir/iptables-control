package nginx

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]Domain, error) {
	rows, err := r.db.Query(`SELECT id, domain, upstream_ip, upstream_port, upstream_scheme, upstream_ssl_verify,
		ssl_enabled, enabled, auth_enabled, auth_username, auth_password_hash, auth_cookie_secret, auth_cookie_max_age, auth_login_css,
		created_at, updated_at FROM nginx_domains ORDER BY domain`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		if err := rows.Scan(&d.ID, &d.Domain, &d.UpstreamIP, &d.UpstreamPort, &d.UpstreamScheme, &d.UpstreamSSLVerify,
			&d.SSLEnabled, &d.Enabled, &d.AuthEnabled, &d.AuthUsername, &d.AuthPasswordHash, &d.AuthCookieSecret, &d.AuthCookieMaxAge, &d.AuthLoginCSS,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	if domains == nil {
		domains = []Domain{}
	}
	return domains, nil
}

func (r *Repository) GetByID(id int64) (*Domain, error) {
	var d Domain
	err := r.db.QueryRow(
		`SELECT id, domain, upstream_ip, upstream_port, upstream_scheme, upstream_ssl_verify,
			ssl_enabled, enabled, auth_enabled, auth_username, auth_password_hash, auth_cookie_secret, auth_cookie_max_age, auth_login_css,
			created_at, updated_at FROM nginx_domains WHERE id = ?`,
		id,
	).Scan(&d.ID, &d.Domain, &d.UpstreamIP, &d.UpstreamPort, &d.UpstreamScheme, &d.UpstreamSSLVerify,
		&d.SSLEnabled, &d.Enabled, &d.AuthEnabled, &d.AuthUsername, &d.AuthPasswordHash, &d.AuthCookieSecret, &d.AuthCookieMaxAge, &d.AuthLoginCSS,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) GetByHostname(hostname string) (*Domain, error) {
	var d Domain
	err := r.db.QueryRow(
		`SELECT id, domain, upstream_ip, upstream_port, upstream_scheme, upstream_ssl_verify,
			ssl_enabled, enabled, auth_enabled, auth_username, auth_password_hash, auth_cookie_secret, auth_cookie_max_age, auth_login_css,
			created_at, updated_at FROM nginx_domains WHERE domain = ?`,
		hostname,
	).Scan(&d.ID, &d.Domain, &d.UpstreamIP, &d.UpstreamPort, &d.UpstreamScheme, &d.UpstreamSSLVerify,
		&d.SSLEnabled, &d.Enabled, &d.AuthEnabled, &d.AuthUsername, &d.AuthPasswordHash, &d.AuthCookieSecret, &d.AuthCookieMaxAge, &d.AuthLoginCSS,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) Create(d *Domain) error {
	res, err := r.db.Exec(
		`INSERT INTO nginx_domains (domain, upstream_ip, upstream_port, upstream_scheme, upstream_ssl_verify,
			auth_enabled, auth_username, auth_password_hash, auth_cookie_secret, auth_cookie_max_age, auth_login_css)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.Domain, d.UpstreamIP, d.UpstreamPort, d.UpstreamScheme, d.UpstreamSSLVerify,
		d.AuthEnabled, d.AuthUsername, d.AuthPasswordHash, d.AuthCookieSecret, d.AuthCookieMaxAge, d.AuthLoginCSS,
	)
	if err != nil {
		return err
	}
	d.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) Update(d *Domain) error {
	_, err := r.db.Exec(
		`UPDATE nginx_domains SET domain = ?, upstream_ip = ?, upstream_port = ?, upstream_scheme = ?, upstream_ssl_verify = ?,
			ssl_enabled = ?, enabled = ?, auth_enabled = ?, auth_username = ?, auth_password_hash = ?, auth_cookie_secret = ?,
			auth_cookie_max_age = ?, auth_login_css = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		d.Domain, d.UpstreamIP, d.UpstreamPort, d.UpstreamScheme, d.UpstreamSSLVerify,
		d.SSLEnabled, d.Enabled, d.AuthEnabled, d.AuthUsername, d.AuthPasswordHash, d.AuthCookieSecret,
		d.AuthCookieMaxAge, d.AuthLoginCSS, d.ID,
	)
	return err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM nginx_domains WHERE id = ?", id)
	return err
}
