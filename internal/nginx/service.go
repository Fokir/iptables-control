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
	repo     *Repository
	sitesDir string
}

func NewService(repo *Repository, sitesDir string) *Service {
	return &Service{repo: repo, sitesDir: sitesDir}
}

func (s *Service) GetAll() ([]Domain, error) {
	return s.repo.GetAll()
}

// SyncConfigs writes nginx configs for all enabled domains that are missing on disk.
func (s *Service) SyncConfigs() error {
	if runtime.GOOS != "linux" {
		return nil
	}

	// Migrate: remove stale htpasswd files from sites-enabled (they belong in htpasswd dir)
	s.cleanupStaleHtpasswd()

	domains, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("get domains: %w", err)
	}

	for _, d := range domains {
		if !d.Enabled {
			continue
		}
		path := s.configPath(d.Domain)
		if _, err := os.Stat(path); err == nil {
			continue // config already exists
		}
		slog.Info("syncing missing nginx config", "domain", d.Domain)
		if err := s.writeAndReload(&d); err != nil {
			slog.Error("failed to sync nginx config", "error", err, "domain", d.Domain)
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
	if err := validateBasicAuth(req.BasicAuthUser, req.BasicAuthPassword); err != nil {
		return nil, err
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
		BasicAuthUser:     req.BasicAuthUser,
		BasicAuthPassword: req.BasicAuthPassword,
		Enabled:           true,
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

	// If user is set but password is empty, keep existing password
	if req.BasicAuthUser != "" && req.BasicAuthPassword == "" && domain.BasicAuthUser == req.BasicAuthUser {
		req.BasicAuthPassword = domain.BasicAuthPassword
	}

	if err := validateBasicAuth(req.BasicAuthUser, req.BasicAuthPassword); err != nil {
		return nil, err
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
	domain.BasicAuthUser = req.BasicAuthUser
	domain.BasicAuthPassword = req.BasicAuthPassword

	if err := s.repo.Update(domain); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}

	// Remove old config if domain name changed
	if oldDomain != domain.Domain {
		s.removeConfig(oldDomain)
		s.removeHtpasswd(oldDomain)
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
	s.removeHtpasswd(domain.Domain)
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
		s.removeHtpasswd(domain.Domain)
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

func (s *Service) htpasswdDir() string {
	return filepath.Join(filepath.Dir(s.sitesDir), "htpasswd")
}

func (s *Service) htpasswdPath(domain string) string {
	return filepath.Join(s.htpasswdDir(), fmt.Sprintf("sc_%s.htpasswd", domain))
}

func (s *Service) writeAndReload(domain *Domain) error {
	if runtime.GOOS != "linux" {
		slog.Warn("nginx config write skipped on non-Linux", "domain", domain.Domain)
		return nil
	}

	hasAuth := domain.BasicAuthUser != "" && domain.BasicAuthPassword != ""

	htpasswdFile := s.htpasswdPath(domain.Domain)
	if hasAuth {
		if err := s.writeHtpasswd(domain); err != nil {
			return fmt.Errorf("write htpasswd: %w", err)
		}
	} else {
		s.removeHtpasswd(domain.Domain)
		htpasswdFile = ""
	}

	content, err := renderConfig(domain, htpasswdFile)
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
		if hasAuth {
			s.removeHtpasswd(domain.Domain)
		}
		return fmt.Errorf("nginx config test failed: %s: %w", string(output), err)
	}

	s.reloadNginx()
	return nil
}

func (s *Service) writeHtpasswd(domain *Domain) error {
	if err := os.MkdirAll(s.htpasswdDir(), 0755); err != nil {
		return fmt.Errorf("create htpasswd dir: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(domain.BasicAuthPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	line := fmt.Sprintf("%s:%s\n", domain.BasicAuthUser, string(hash))
	return os.WriteFile(s.htpasswdPath(domain.Domain), []byte(line), 0644)
}

func (s *Service) removeConfig(domain string) {
	os.Remove(s.configPath(domain))
}

// cleanupStaleHtpasswd removes htpasswd files that were incorrectly placed in sites-enabled.
func (s *Service) cleanupStaleHtpasswd() {
	entries, err := os.ReadDir(s.sitesDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".htpasswd") {
			old := filepath.Join(s.sitesDir, e.Name())
			slog.Info("removing stale htpasswd from sites-enabled", "file", old)
			os.Remove(old)
		}
	}
}

func (s *Service) removeHtpasswd(domain string) {
	os.Remove(s.htpasswdPath(domain))
}

func (s *Service) reloadNginx() {
	if runtime.GOOS != "linux" {
		return
	}
	if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
		slog.Error("failed to reload nginx", "error", err)
	}
}

func validateBasicAuth(user, password string) error {
	if (user == "") != (password == "") {
		return fmt.Errorf("basic auth requires both user and password, or neither")
	}
	return nil
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
