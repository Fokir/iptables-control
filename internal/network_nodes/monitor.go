package network_nodes

import (
	"log/slog"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Monitor struct {
	mu       sync.RWMutex
	statuses map[int64]*NodeStatus
	repo     *Repository
	interval time.Duration
	stop     chan struct{}
}

func NewMonitor(repo *Repository, interval time.Duration) *Monitor {
	return &Monitor{
		statuses: make(map[int64]*NodeStatus),
		repo:     repo,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	go func() {
		m.pingAll()
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.pingAll()
			case <-m.stop:
				return
			}
		}
	}()
}

func (m *Monitor) Stop() {
	close(m.stop)
}

func (m *Monitor) GetStatuses() []NodeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]NodeStatus, 0, len(m.statuses))
	for _, s := range m.statuses {
		result = append(result, *s)
	}
	return result
}

func (m *Monitor) pingAll() {
	nodes, err := m.repo.GetAll()
	if err != nil {
		slog.Error("monitor: failed to get nodes", "error", err)
		return
	}

	activeIDs := make(map[int64]bool, len(nodes))

	type pingResult struct {
		nodeID  int64
		latency time.Duration
		ok      bool
	}

	results := make(chan pingResult, len(nodes))
	var wg sync.WaitGroup

	for _, node := range nodes {
		activeIDs[node.ID] = true
		wg.Add(1)
		go func(n NetworkNode) {
			defer wg.Done()
			latency, ok := pingNode(n.IP)
			results <- pingResult{nodeID: n.ID, latency: latency, ok: ok}
		}(node)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for r := range results {
		s, exists := m.statuses[r.nodeID]
		if !exists {
			s = &NodeStatus{NodeID: r.nodeID}
			m.statuses[r.nodeID] = s
		}

		if r.ok {
			s.LatencyMs = float64(r.latency.Microseconds()) / 1000.0
			nowCopy := now
			s.LastSeen = &nowCopy
			if !s.IsOnline {
				s.IsOnline = true
				s.FirstOnline = &nowCopy
			}
		} else {
			s.IsOnline = false
			s.LatencyMs = 0
			s.FirstOnline = nil
		}
	}

	// Remove statuses for deleted nodes
	for id := range m.statuses {
		if !activeIDs[id] {
			delete(m.statuses, id)
		}
	}
}

func pingNode(ip string) (time.Duration, bool) {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return 0, false
	}
	pinger.Count = 1
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(true)

	if err := pinger.Run(); err != nil {
		return 0, false
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return 0, false
	}
	return stats.AvgRtt, true
}
