ALTER TABLE nginx_domains ADD COLUMN upstream_scheme TEXT NOT NULL DEFAULT 'http';
ALTER TABLE nginx_domains ADD COLUMN upstream_ssl_verify BOOLEAN NOT NULL DEFAULT 1;
