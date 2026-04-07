-- Replace basic auth with cookie-based auth
ALTER TABLE nginx_domains ADD COLUMN auth_enabled BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE nginx_domains ADD COLUMN auth_username TEXT NOT NULL DEFAULT '';
ALTER TABLE nginx_domains ADD COLUMN auth_password_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE nginx_domains ADD COLUMN auth_cookie_secret TEXT NOT NULL DEFAULT '';
ALTER TABLE nginx_domains ADD COLUMN auth_cookie_max_age INTEGER NOT NULL DEFAULT 2592000;
ALTER TABLE nginx_domains ADD COLUMN auth_login_css TEXT NOT NULL DEFAULT '';

-- Migrate existing basic auth data
UPDATE nginx_domains SET auth_enabled = 1, auth_username = basic_auth_user, auth_password_hash = basic_auth_password WHERE basic_auth_user != '' AND basic_auth_password != '';

-- Clear old plain-text credentials
UPDATE nginx_domains SET basic_auth_user = '', basic_auth_password = '' WHERE basic_auth_user != '';
