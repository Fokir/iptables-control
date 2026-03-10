package nftables

import (
	"fmt"
	"log/slog"

	"github.com/sokol/system-control/internal/pkg/validate"
)

type Service struct {
	repo   *Repository
	engine *Engine
}

func NewService(repo *Repository, engine *Engine) *Service {
	return &Service{repo: repo, engine: engine}
}

func (s *Service) GetAll() ([]NatGroup, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*NatGroup, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(req CreateGroupRequest) (*NatGroup, error) {
	if err := validateGroup(req.Name, req.TargetIP, req.TargetReverseIP, req.DestinationIP); err != nil {
		return nil, err
	}
	if err := validatePorts(req.Ports); err != nil {
		return nil, err
	}

	group := &NatGroup{
		Name:            req.Name,
		Enabled:         req.Enabled,
		TargetIP:        req.TargetIP,
		TargetReverseIP: req.TargetReverseIP,
		DestinationIP:   req.DestinationIP,
	}

	if err := s.repo.Create(group); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}

	ports, err := s.repo.ReplaceGroupPorts(group.ID, req.Ports)
	if err != nil {
		return nil, fmt.Errorf("create ports: %w", err)
	}
	group.Ports = ports

	if group.Enabled {
		if err := s.applyAllRules(); err != nil {
			slog.Error("failed to apply rules after create", "error", err)
		}
	}

	return group, nil
}

func (s *Service) Update(id int64, req UpdateGroupRequest) (*NatGroup, error) {
	if err := validateGroup(req.Name, req.TargetIP, req.TargetReverseIP, req.DestinationIP); err != nil {
		return nil, err
	}
	if err := validatePorts(req.Ports); err != nil {
		return nil, err
	}

	group, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	group.Name = req.Name
	group.Enabled = req.Enabled
	group.TargetIP = req.TargetIP
	group.TargetReverseIP = req.TargetReverseIP
	group.DestinationIP = req.DestinationIP

	if err := s.repo.Update(group); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}

	ports, err := s.repo.ReplaceGroupPorts(group.ID, req.Ports)
	if err != nil {
		return nil, fmt.Errorf("replace ports: %w", err)
	}
	group.Ports = ports

	if err := s.applyAllRules(); err != nil {
		slog.Error("failed to apply rules after update", "error", err)
	}

	return group, nil
}

func (s *Service) Delete(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	if err := s.applyAllRules(); err != nil {
		slog.Error("failed to apply rules after delete", "error", err)
	}

	return nil
}

func (s *Service) SetEnabled(id int64, enabled bool) (*NatGroup, error) {
	group, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	group.Enabled = enabled
	if err := s.repo.Update(group); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}

	if err := s.applyAllRules(); err != nil {
		slog.Error("failed to apply rules after enable/disable", "error", err)
	}

	return group, nil
}

func (s *Service) SyncRules() error {
	return s.applyAllRules()
}

func (s *Service) applyAllRules() error {
	groups, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("get all groups: %w", err)
	}
	return s.engine.Apply(groups)
}

func validateGroup(name, targetIP, targetReverseIP, destinationIP string) error {
	if err := validate.Required("name", name); err != nil {
		return err
	}
	if err := validate.IP(targetIP); err != nil {
		return fmt.Errorf("targetIp: %w", err)
	}
	if targetReverseIP != "" {
		if err := validate.IP(targetReverseIP); err != nil {
			return fmt.Errorf("targetReverseIp: %w", err)
		}
	}
	if destinationIP != "" {
		if err := validate.IP(destinationIP); err != nil {
			return fmt.Errorf("destinationIp: %w", err)
		}
	}
	return nil
}

func validatePorts(ports []CreatePortRequest) error {
	for i, p := range ports {
		if err := validate.Port(p.ExternalPort); err != nil {
			return fmt.Errorf("port[%d].externalPort: %w", i, err)
		}
		if err := validate.Port(p.InternalPort); err != nil {
			return fmt.Errorf("port[%d].internalPort: %w", i, err)
		}
		if !p.ProtocolTCP && !p.ProtocolUDP {
			return fmt.Errorf("port[%d]: at least one protocol must be selected", i)
		}
	}
	return nil
}
