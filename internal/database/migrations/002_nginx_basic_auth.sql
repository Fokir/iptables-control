ALTER TABLE nginx_domains ADD COLUMN basic_auth_user TEXT NOT NULL DEFAULT '';
ALTER TABLE nginx_domains ADD COLUMN basic_auth_password TEXT NOT NULL DEFAULT '';
