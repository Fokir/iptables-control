package traffic

import (
	"log/slog"
	"sync"
	"time"

	"github.com/sokol/system-control/internal/network_nodes"
)

type Collector struct {
	mu             sync.Mutex
	repo           *Repository
	engine         *Engine
	nodesRepo      *network_nodes.Repository
	interfaces     []string
	collectInterval time.Duration
	stop           chan struct{}

	// Previous counter values for computing deltas
	prevIface map[string]InterfaceSnapshot
}

func NewCollector(repo *Repository, engine *Engine, nodesRepo *network_nodes.Repository, interfaces []string, interval time.Duration) *Collector {
	return &Collector{
		repo:            repo,
		engine:          engine,
		nodesRepo:       nodesRepo,
		interfaces:      interfaces,
		collectInterval: interval,
		stop:            make(chan struct{}),
		prevIface:       make(map[string]InterfaceSnapshot),
	}
}

func (c *Collector) Start() {
	go c.collectLoop()
	go c.aggregateLoop()
}

func (c *Collector) Stop() {
	close(c.stop)
}

func (c *Collector) collectLoop() {
	// Initial collection to establish baselines
	c.collect()

	ticker := time.NewTicker(c.collectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.collect()
		case <-c.stop:
			return
		}
	}
}

func (c *Collector) collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()

	// Collect interface counters
	snapshots, err := c.engine.CollectInterfaces(c.interfaces)
	if err != nil {
		slog.Error("traffic: failed to collect interface stats", "error", err)
	}

	for _, snap := range snapshots {
		prev, exists := c.prevIface[snap.Name]
		if exists && snap.BytesIn >= prev.BytesIn && snap.BytesOut >= prev.BytesOut {
			deltaIn := int64(snap.BytesIn - prev.BytesIn)
			deltaOut := int64(snap.BytesOut - prev.BytesOut)
			if deltaIn > 0 || deltaOut > 0 {
				if err := c.repo.InsertRaw(nil, snap.Name, deltaIn, deltaOut, now); err != nil {
					slog.Error("traffic: failed to insert interface data", "interface", snap.Name, "error", err)
				}
			}
		}
		c.prevIface[snap.Name] = snap
	}
}

func (c *Collector) aggregateLoop() {
	// Run aggregation every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.aggregate()
		case <-c.stop:
			return
		}
	}
}

func (c *Collector) aggregate() {
	now := time.Now().UTC()

	// Raw -> 5min buckets (keep raw for 2 hours)
	cutoff := now.Add(-2 * time.Hour)
	if err := c.repo.AggregateRaw(cutoff, 300); err != nil {
		slog.Error("traffic: failed to aggregate raw->5min", "error", err)
	}
	if err := c.repo.PurgeRaw(cutoff); err != nil {
		slog.Error("traffic: failed to purge raw", "error", err)
	}

	// 5min -> 1hr buckets (keep 5min for 48 hours)
	cutoff = now.Add(-48 * time.Hour)
	if err := c.repo.AggregateAgg(300, 3600, cutoff); err != nil {
		slog.Error("traffic: failed to aggregate 5min->1hr", "error", err)
	}
	if err := c.repo.PurgeAgg(300, cutoff); err != nil {
		slog.Error("traffic: failed to purge 5min", "error", err)
	}

	// 1hr -> 1day buckets (keep 1hr for 30 days)
	cutoff = now.Add(-30 * 24 * time.Hour)
	if err := c.repo.AggregateAgg(3600, 86400, cutoff); err != nil {
		slog.Error("traffic: failed to aggregate 1hr->1day", "error", err)
	}
	if err := c.repo.PurgeAgg(3600, cutoff); err != nil {
		slog.Error("traffic: failed to purge 1hr", "error", err)
	}

	// Purge daily buckets older than 3 months
	cutoff = now.Add(-90 * 24 * time.Hour)
	if err := c.repo.PurgeAgg(86400, cutoff); err != nil {
		slog.Error("traffic: failed to purge daily", "error", err)
	}
}
