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

// rawRetention must match the cutoff in collector.aggregate().
const rawRetention = 2 * time.Hour

var intervals = map[string]intervalConfig{
	"1m":  {duration: 1 * time.Minute, useRaw: true},
	"5m":  {duration: 5 * time.Minute, useRaw: true},
	"30m": {duration: 30 * time.Minute, useRaw: true},
	"1h":  {duration: 1 * time.Hour, useRaw: true},
	"1d":  {duration: 24 * time.Hour, bucketSeconds: 300},
	"1w":  {duration: 7 * 24 * time.Hour, bucketSeconds: 3600},
}

// queryPoints fetches traffic data, combining raw and aggregated sources
// to avoid gaps caused by aggregation lag.
func (s *Service) queryPoints(nodeID *int64, ifaceName string, cfg intervalConfig) ([]TrafficPoint, error) {
	now := time.Now().UTC()
	from := now.Add(-cfg.duration)

	if cfg.useRaw {
		return s.repo.QueryRaw(nodeID, ifaceName, from, now)
	}

	// For longer intervals: raw covers the recent window that hasn't been
	// aggregated yet, aggregated covers the older portion.
	rawBoundary := now.Add(-rawRetention)
	var allPoints []TrafficPoint

	// Aggregated data for the older portion
	if from.Before(rawBoundary) {
		aggPoints, err := s.repo.QueryAggregated(nodeID, ifaceName, from, rawBoundary, cfg.bucketSeconds)
		if err != nil {
			return nil, err
		}
		allPoints = append(allPoints, aggPoints...)
	}

	// Raw data for the recent portion
	rawFrom := rawBoundary
	if from.After(rawBoundary) {
		rawFrom = from
	}
	rawPoints, err := s.repo.QueryRaw(nodeID, ifaceName, rawFrom, now)
	if err != nil {
		return nil, err
	}
	allPoints = append(allPoints, rawPoints...)

	if allPoints == nil {
		allPoints = []TrafficPoint{}
	}
	return allPoints, nil
}

func (s *Service) GetNodeTraffic(nodeID int64, interval string) (*TrafficSeries, error) {
	cfg, ok := intervals[interval]
	if !ok {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	points, err := s.queryPoints(&nodeID, "", cfg)
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

	points, err := s.queryPoints(nil, name, cfg)
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
