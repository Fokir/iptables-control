package audit

import (
	"encoding/json"
	"log/slog"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(userID *int64, username, action, entityType string, entityID *int64, details any, ipAddress string) {
	var detailsStr string
	if details != nil {
		b, _ := json.Marshal(details)
		detailsStr = string(b)
	}

	entry := &LogEntry{
		UserID:     userID,
		Username:   username,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    detailsStr,
		IPAddress:  ipAddress,
	}

	if err := s.repo.Create(entry); err != nil {
		slog.Error("failed to write audit log", "error", err)
	}
}

func (s *Service) List(params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 50
	}
	return s.repo.List(params)
}

func (s *Service) Cleanup(days int) error {
	return s.repo.Cleanup(days)
}
