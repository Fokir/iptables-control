package settings

import "log/slog"

const keyForceTLS12 = "force_tls12"

// ChangeListener is notified after a setting changes so dependents (e.g. nginx)
// can rebuild their configuration.
type ChangeListener interface {
	OnSettingsChanged() error
}

type Service struct {
	repo     *Repository
	listener ChangeListener
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetChangeListener(l ChangeListener) {
	s.listener = l
}

func (s *Service) ForceTLS12() bool {
	v, err := s.repo.Get(keyForceTLS12)
	if err != nil {
		slog.Error("failed to read force_tls12 setting", "error", err)
		return false
	}
	return v == "true"
}

func (s *Service) SetForceTLS12(v bool) error {
	value := "false"
	if v {
		value = "true"
	}
	if err := s.repo.Set(keyForceTLS12, value); err != nil {
		return err
	}
	if s.listener != nil {
		if err := s.listener.OnSettingsChanged(); err != nil {
			slog.Error("settings change listener failed", "error", err)
		}
	}
	return nil
}
