package network_nodes

import (
	"fmt"

	"github.com/sokol/system-control/internal/pkg/validate"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
	return node, nil
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
