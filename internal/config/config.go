package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port              int
	AdminUser         string
	AdminPassword     string
	DBPath            string
	NginxSitesDir     string
	SessionMaxAge     int // seconds
	TrafficInterfaces string
	TrafficInterval   int // seconds
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnvInt("PORT", 8080),
		AdminUser:     getEnv("ADMIN_USER", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		DBPath:        getEnv("DB_PATH", "system-control.db"),
		NginxSitesDir: getEnv("NGINX_SITES_DIR", "/etc/nginx/sites-enabled"),
		SessionMaxAge:     getEnvInt("SESSION_MAX_AGE", 604800), // 7 days
		TrafficInterfaces: getEnv("TRAFFIC_INTERFACES", ""),
		TrafficInterval:   getEnvInt("TRAFFIC_INTERVAL", 10),
	}

	if cfg.AdminPassword == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
