package nginx

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/sokol/system-control/internal/pkg/validate"
)

type Service struct {
	repo        *Repository
	sitesDir    string
	backendPort int
}

func NewService(repo *Repository, sitesDir string, backendPort int) *Service {
	return &Service{repo: repo, sitesDir: sitesDir, backendPort: backendPort}
}

func (s *Service) GetAll() ([]Domain, error) {
	return s.repo.GetAll()
}

// SyncConfigs writes nginx configs for all enabled domains that are missing on disk.
// It also migrates legacy basic auth domains to cookie auth.
func (s *Service) SyncConfigs() error {
	domains, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("get domains: %w", err)
	}

	for i := range domains {
		d := &domains[i]

		// Migrate: generate missing cookie secret for auth-enabled domains
		if d.AuthEnabled && d.AuthCookieSecret == "" {
			secret, err := generateCookieSecret()
			if err != nil {
				slog.Error("failed to generate cookie secret", "error", err, "domain", d.Domain)
				continue
			}
			d.AuthCookieSecret = secret

			// If password hash doesn't look like bcrypt, it's a plain text leftover — re-hash it
			if d.AuthPasswordHash != "" && !strings.HasPrefix(d.AuthPasswordHash, "$2") {
				hash, err := bcrypt.GenerateFromPassword([]byte(d.AuthPasswordHash), bcrypt.DefaultCost)
				if err != nil {
					slog.Error("failed to hash legacy password", "error", err, "domain", d.Domain)
					continue
				}
				d.AuthPasswordHash = string(hash)
			}

			if d.AuthCookieMaxAge <= 0 {
				d.AuthCookieMaxAge = defaultCookieMaxAge
			}

			if err := s.repo.Update(d); err != nil {
				slog.Error("failed to update domain after migration", "error", err, "domain", d.Domain)
				continue
			}
			slog.Info("migrated domain to cookie auth", "domain", d.Domain)

			// Force regenerate nginx config to replace auth_basic with auth_request
			if runtime.GOOS == "linux" && d.Enabled {
				if err := s.writeAndReload(d); err != nil {
					slog.Error("failed to regenerate nginx config after migration", "error", err, "domain", d.Domain)
				}
			}
		}

		if runtime.GOOS != "linux" || !d.Enabled {
			continue
		}

		path := s.configPath(d.Domain)
		if _, err := os.Stat(path); err == nil {
			continue // config already exists
		}
		slog.Info("syncing missing nginx config", "domain", d.Domain)
		if err := s.writeAndReload(d); err != nil {
			slog.Error("failed to sync nginx config", "error", err, "domain", d.Domain)
		}
	}

	// Cleanup: remove old htpasswd directory
	if runtime.GOOS == "linux" {
		htpasswdDir := filepath.Join(filepath.Dir(s.sitesDir), "htpasswd")
		if entries, err := os.ReadDir(htpasswdDir); err == nil && len(entries) > 0 {
			slog.Info("cleaning up legacy htpasswd files", "dir", htpasswdDir)
			for _, e := range entries {
				os.Remove(filepath.Join(htpasswdDir, e.Name()))
			}
		}
	}

	return nil
}

func (s *Service) Create(req CreateDomainRequest) (*Domain, error) {
	if err := validate.Domain(req.Domain); err != nil {
		return nil, err
	}
	if err := validate.IP(req.UpstreamIP); err != nil {
		return nil, fmt.Errorf("upstreamIp: %w", err)
	}
	if req.UpstreamPort == 0 {
		req.UpstreamPort = 80
	}
	if err := validate.Port(req.UpstreamPort); err != nil {
		return nil, fmt.Errorf("upstreamPort: %w", err)
	}

	if req.AuthEnabled {
		if req.AuthUsername == "" {
			return nil, fmt.Errorf("auth username is required when auth is enabled")
		}
		if req.AuthPassword == "" {
			return nil, fmt.Errorf("auth password is required when auth is enabled")
		}
	}

	scheme := req.UpstreamScheme
	if scheme == "" {
		scheme = "http"
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("upstreamScheme: must be http or https")
	}

	sslVerify := true
	if req.UpstreamSSLVerify != nil {
		sslVerify = *req.UpstreamSSLVerify
	}

	domain := &Domain{
		Domain:            req.Domain,
		UpstreamIP:        req.UpstreamIP,
		UpstreamPort:      req.UpstreamPort,
		UpstreamScheme:    scheme,
		UpstreamSSLVerify: sslVerify,
		AuthEnabled:       req.AuthEnabled,
		AuthUsername:      req.AuthUsername,
		AuthCookieMaxAge:  req.AuthCookieMaxAge,
		AuthLoginCSS:      req.AuthLoginCSS,
		Enabled:           true,
	}

	if req.AuthEnabled {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.AuthPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		domain.AuthPasswordHash = string(hash)

		secret, err := generateCookieSecret()
		if err != nil {
			return nil, fmt.Errorf("generate cookie secret: %w", err)
		}
		domain.AuthCookieSecret = secret
	}

	if domain.AuthCookieMaxAge <= 0 {
		domain.AuthCookieMaxAge = defaultCookieMaxAge // 30 days
	}

	if err := s.repo.Create(domain); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}

	if err := s.writeAndReload(domain); err != nil {
		return domain, fmt.Errorf("domain created but nginx config failed: %w", err)
	}

	return domain, nil
}

func (s *Service) Update(id int64, req UpdateDomainRequest) (*Domain, error) {
	if err := validate.Domain(req.Domain); err != nil {
		return nil, err
	}
	if err := validate.IP(req.UpstreamIP); err != nil {
		return nil, fmt.Errorf("upstreamIp: %w", err)
	}
	if req.UpstreamPort == 0 {
		req.UpstreamPort = 80
	}

	domain, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	scheme := req.UpstreamScheme
	if scheme == "" {
		scheme = "http"
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("upstreamScheme: must be http or https")
	}

	sslVerify := domain.UpstreamSSLVerify
	if req.UpstreamSSLVerify != nil {
		sslVerify = *req.UpstreamSSLVerify
	}

	oldDomain := domain.Domain
	domain.Domain = req.Domain
	domain.UpstreamIP = req.UpstreamIP
	domain.UpstreamPort = req.UpstreamPort
	domain.UpstreamScheme = scheme
	domain.UpstreamSSLVerify = sslVerify
	domain.AuthEnabled = req.AuthEnabled
	domain.AuthLoginCSS = req.AuthLoginCSS

	if req.AuthUsername != "" {
		domain.AuthUsername = req.AuthUsername
	}

	if req.AuthCookieMaxAge > 0 {
		domain.AuthCookieMaxAge = req.AuthCookieMaxAge
	}

	if req.AuthEnabled {
		if domain.AuthUsername == "" {
			return nil, fmt.Errorf("auth username is required when enabling auth")
		}

		// If password provided, re-hash and regenerate cookie secret (invalidates old cookies)
		if req.AuthPassword != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.AuthPassword), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("hash password: %w", err)
			}
			domain.AuthPasswordHash = string(hash)

			secret, err := generateCookieSecret()
			if err != nil {
				return nil, fmt.Errorf("generate cookie secret: %w", err)
			}
			domain.AuthCookieSecret = secret
		}

		// Generate secret if not set (e.g. enabling auth on existing domain)
		if domain.AuthCookieSecret == "" {
			secret, err := generateCookieSecret()
			if err != nil {
				return nil, fmt.Errorf("generate cookie secret: %w", err)
			}
			domain.AuthCookieSecret = secret
		}

		if domain.AuthPasswordHash == "" {
			return nil, fmt.Errorf("auth password is required when enabling auth")
		}
	}

	if err := s.repo.Update(domain); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}

	// Remove old config if domain name changed
	if oldDomain != domain.Domain {
		s.removeConfig(oldDomain)
	}

	if domain.Enabled {
		if err := s.writeAndReload(domain); err != nil {
			slog.Error("failed to write nginx config", "error", err, "domain", domain.Domain)
		}
	}

	return domain, nil
}

func (s *Service) Delete(id int64) error {
	domain, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	s.removeConfig(domain.Domain)
	s.reloadNginx()

	return s.repo.Delete(id)
}

func (s *Service) SetEnabled(id int64, enabled bool) (*Domain, error) {
	domain, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	domain.Enabled = enabled
	if err := s.repo.Update(domain); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}

	if enabled {
		if err := s.writeAndReload(domain); err != nil {
			slog.Error("failed to enable nginx config", "error", err, "domain", domain.Domain)
		}
	} else {
		s.removeConfig(domain.Domain)
		s.reloadNginx()
	}

	return domain, nil
}

func (s *Service) RequestSSL(id int64, email string) error {
	if err := validate.Email(email); err != nil {
		return fmt.Errorf("email: %w", err)
	}

	domain, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	if runtime.GOOS != "linux" {
		slog.Warn("certbot is only supported on Linux, skipping SSL")
		return nil
	}

	cmd := exec.Command("certbot", "--nginx", "-d", domain.Domain, "--non-interactive", "--agree-tos", "-m", email)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot failed: %s: %w", string(output), err)
	}

	domain.SSLEnabled = true
	return s.repo.Update(domain)
}

func (s *Service) configPath(domain string) string {
	return filepath.Join(s.sitesDir, fmt.Sprintf("sc_%s.conf", domain))
}

func (s *Service) writeAndReload(domain *Domain) error {
	if runtime.GOOS != "linux" {
		slog.Warn("nginx config write skipped on non-Linux", "domain", domain.Domain)
		return nil
	}

	content, err := renderConfig(domain, s.backendPort)
	if err != nil {
		return fmt.Errorf("render config: %w", err)
	}

	path := s.configPath(domain.Domain)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	// Test config before reload
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Rollback: remove bad config
		os.Remove(path)
		return fmt.Errorf("nginx config test failed: %s: %w", string(output), err)
	}

	s.reloadNginx()
	return nil
}

func (s *Service) removeConfig(domain string) {
	os.Remove(s.configPath(domain))
}

func (s *Service) reloadNginx() {
	if runtime.GOOS != "linux" {
		return
	}
	if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
		slog.Error("failed to reload nginx", "error", err)
	}
}

func (s *Service) ListExternal() ([]ExternalDomain, error) {
	if runtime.GOOS != "linux" {
		return []ExternalDomain{}, nil
	}

	entries, err := os.ReadDir(s.sitesDir)
	if err != nil {
		return nil, fmt.Errorf("read sites dir: %w", err)
	}

	var result []ExternalDomain
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Skip our managed configs and non-conf files
		if strings.HasPrefix(name, "sc_") || !strings.HasSuffix(name, ".conf") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(s.sitesDir, name))
		if err != nil {
			slog.Warn("failed to read external config", "file", name, "error", err)
			continue
		}

		ext, err := parseExternalConfig(string(content))
		if err != nil {
			slog.Debug("failed to parse external config", "file", name, "error", err)
			continue
		}
		ext.Filename = name
		result = append(result, ext)
	}

	if result == nil {
		result = []ExternalDomain{}
	}
	return result, nil
}

func (s *Service) ImportExternal(filename string) (*Domain, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("import is only supported on Linux")
	}

	if err := validate.Filename(filename); err != nil {
		return nil, err
	}

	srcPath := filepath.Join(s.sitesDir, filename)
	info, err := os.Stat(srcPath)
	if err != nil || info.IsDir() {
		return nil, fmt.Errorf("config file not found: %s", filename)
	}

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	ext, err := parseExternalConfig(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Check domain doesn't already exist in DB
	existing, _ := s.repo.GetAll()
	for _, d := range existing {
		if d.Domain == ext.Domain {
			return nil, fmt.Errorf("domain %s already exists in database", ext.Domain)
		}
	}

	// Create backup
	backupPath := srcPath + ".bak"
	if err := os.WriteFile(backupPath, content, fs.FileMode(info.Mode())); err != nil {
		return nil, fmt.Errorf("create backup: %w", err)
	}
	slog.Info("created backup of external config", "backup", backupPath)

	// Create domain in DB
	domain := &Domain{
		Domain:            ext.Domain,
		UpstreamIP:        ext.UpstreamIP,
		UpstreamPort:      ext.UpstreamPort,
		UpstreamScheme:    "http",
		UpstreamSSLVerify: true,
		SSLEnabled:        ext.SSLEnabled,
		AuthCookieMaxAge:  defaultCookieMaxAge,
		Enabled:           true,
	}

	if ext.UpstreamPort == 0 {
		domain.UpstreamPort = 80
	}

	if err := s.repo.Create(domain); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}

	// Write our managed config
	if err := s.writeAndReload(domain); err != nil {
		slog.Error("failed to write nginx config after import", "error", err, "domain", domain.Domain)
		return domain, nil
	}

	// Remove original file (backup exists)
	os.Remove(srcPath)

	return domain, nil
}
