package traffic

import (
	"fmt"
	"time"

	"github.com/sokol/system-control/internal/network_nodes"
)

type Service struct {
	repo      *Repository
	nodesRepo *network_nodes.Repository
}

func NewService(repo *Repository, nodesRepo *network_nodes.Repository) *Service {
	return &Service{repo: repo, nodesRepo: nodesRepo}
}

// intervalConfig maps interval query parameter to data source and time range.
type intervalConfig struct {
	duration      time.Duration
	useRaw        bool
	bucketSeconds int
}

var intervals = map[string]intervalConfig{
	"1m":  {duration: 1 * time.Minute, useRaw: true},
	"5m":  {duration: 5 * time.Minute, useRaw: true},
	"30m": {duration: 30 * time.Minute, bucketSeconds: 300},
	"1h":  {duration: 1 * time.Hour, bucketSeconds: 300},
	"1d":  {duration: 24 * time.Hour, bucketSeconds: 3600},
	"1w":  {duration: 7 * 24 * time.Hour, bucketSeconds: 86400},
}

func (s *Service) GetNodeTraffic(nodeID int64, interval string) (*TrafficSeries, error) {
	cfg, ok := intervals[interval]
	if !ok {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	now := time.Now().UTC()
	from := now.Add(-cfg.duration)

	var points []TrafficPoint
	var err error

	if cfg.useRaw {
		points, err = s.repo.QueryRaw(&nodeID, "", from, now)
	} else {
		points, err = s.repo.QueryAggregated(&nodeID, "", from, now, cfg.bucketSeconds)
	}
	if err != nil {
		return nil, fmt.Errorf("query traffic: %w", err)
	}

	return &TrafficSeries{
		NodeID: &nodeID,
		Points: points,
	}, nil
}

func (s *Service) GetInterfaceTraffic(name, interval string) (*TrafficSeries, error) {
	cfg, ok := intervals[interval]
	if !ok {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	now := time.Now().UTC()
	from := now.Add(-cfg.duration)

	var points []TrafficPoint
	var err error

	if cfg.useRaw {
		points, err = s.repo.QueryRaw(nil, name, from, now)
	} else {
		points, err = s.repo.QueryAggregated(nil, name, from, now, cfg.bucketSeconds)
	}
	if err != nil {
		return nil, fmt.Errorf("query traffic: %w", err)
	}

	return &TrafficSeries{
		InterfaceName: name,
		Points:        points,
	}, nil
}

func (s *Service) GetAllNodeTraffic(interval string) ([]TrafficSeries, error) {
	nodes, err := s.nodesRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", err)
	}

	var result []TrafficSeries
	for _, node := range nodes {
		series, err := s.GetNodeTraffic(node.ID, interval)
		if err != nil {
			return nil, err
		}
		series.NodeName = node.Name
		result = append(result, *series)
	}
	if result == nil {
		result = []TrafficSeries{}
	}
	return result, nil
}

func (s *Service) GetAllInterfaceTraffic(interval string) ([]TrafficSeries, error) {
	names, err := s.repo.GetInterfaces()
	if err != nil {
		return nil, fmt.Errorf("get interfaces: %w", err)
	}

	var result []TrafficSeries
	for _, name := range names {
		series, err := s.GetInterfaceTraffic(name, interval)
		if err != nil {
			return nil, err
		}
		result = append(result, *series)
	}
	if result == nil {
		result = []TrafficSeries{}
	}
	return result, nil
}

// ValidIntervals returns the list of valid interval strings.
func ValidIntervals() []string {
	return []string{"1m", "5m", "30m", "1h", "1d", "1w"}
}
