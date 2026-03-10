package nginx

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

	domain := &Domain{
		Domain:       req.Domain,
		UpstreamIP:   req.UpstreamIP,
		UpstreamPort: req.UpstreamPort,
		Enabled:      true,
	}

	if err := s.repo.Create(domain); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}

	if err := s.writeAndReload(domain); err != nil {
		slog.Error("failed to write nginx config", "error", err, "domain", domain.Domain)
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

	oldDomain := domain.Domain
	domain.Domain = req.Domain
	domain.UpstreamIP = req.UpstreamIP
	domain.UpstreamPort = req.UpstreamPort

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

	content, err := renderConfig(domain)
	if err != nil {
		return fmt.Errorf("render config: %w", err)
	}

	path := s.configPath(domain.Domain)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	// Test config before reload
	if err := exec.Command("nginx", "-t").Run(); err != nil {
		// Rollback: remove bad config
		os.Remove(path)
		return fmt.Errorf("nginx config test failed: %w", err)
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
