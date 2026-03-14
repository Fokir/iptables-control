package network_nodes

import (
	"fmt"

	"github.com/sokol/system-control/internal/pkg/validate"
)

// NodeChangeListener is notified when nodes are created, updated, or deleted.
type NodeChangeListener interface {
	RefreshNodes()
}

type Service struct {
	repo     *Repository
	monitor  *Monitor
	listener NodeChangeListener
}

func NewService(repo *Repository, monitor *Monitor) *Service {
	return &Service{repo: repo, monitor: monitor}
}

// SetNodeChangeListener registers a listener for node changes.
func (s *Service) SetNodeChangeListener(l NodeChangeListener) {
	s.listener = l
}

func (s *Service) notifyChange() {
	if s.listener != nil {
		go s.listener.RefreshNodes()
	}
}

func (s *Service) GetStatuses() []NodeStatus {
	return s.monitor.GetStatuses()
}

func (s *Service) GetAll() ([]NetworkNode, error) {
	return s.repo.GetAll()
}

func (s *Service) Create(req CreateNodeRequest) (*NetworkNode, error) {
	if err := validate.Required("name", req.Name); err != nil {
		return nil, err
	}
	if err := validate.IP(req.IP); err != nil {
		return nil, err
	}

	node := &NetworkNode{Name: req.Name, IP: req.IP}
	if err := s.repo.Create(node); err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}
	s.notifyChange()
	return node, nil
}

func (s *Service) Update(id int64, req UpdateNodeRequest) (*NetworkNode, error) {
	if err := validate.Required("name", req.Name); err != nil {
		return nil, err
	}
	if err := validate.IP(req.IP); err != nil {
		return nil, err
	}

	node := &NetworkNode{ID: id, Name: req.Name, IP: req.IP}
	if err := s.repo.Update(node); err != nil {
		return nil, fmt.Errorf("update node: %w", err)
	}
	s.notifyChange()
	return node, nil
}

func (s *Service) Delete(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.notifyChange()
	return nil
}
