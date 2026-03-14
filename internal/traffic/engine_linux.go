//go:build linux

package traffic

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

const statsTableName = "system_control_stats"

// SetupNodeCounters creates nftables counting rules for each node IP.
// Uses a separate table to avoid interference with NAT rules.
func (e *Engine) SetupNodeCounters(nodes []NodeInfo) error {
	conn, err := nftables.New()
	if err != nil {
		return fmt.Errorf("nftables connect: %w", err)
	}

	// Delete existing stats table (ignore error if not exists)
	conn.DelTable(&nftables.Table{Name: statsTableName, Family: nftables.TableFamilyIPv4})
	_ = conn.Flush()

	if len(nodes) == 0 {
		return nil
	}

	conn, err = nftables.New()
	if err != nil {
		return fmt.Errorf("nftables connect: %w", err)
	}

	table := conn.AddTable(&nftables.Table{
		Name:   statsTableName,
		Family: nftables.TableFamilyIPv4,
	})

	// Forward chain to count traffic passing through this host to/from nodes
	chain := conn.AddChain(&nftables.Chain{
		Name:     "traffic_count",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
	})

	for _, node := range nodes {
		ip := net.ParseIP(node.IP).To4()
		if ip == nil {
			slog.Error("traffic: invalid node IP, skipping", "nodeId", node.ID, "ip", node.IP)
			continue
		}

		nodeIDStr := strconv.FormatInt(node.ID, 10)

		// Rule: match dst IP = node.IP → count incoming traffic to node
		conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: chain,
			Exprs: []expr.Any{
				&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
				&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ip},
				&expr.Counter{},
			},
			UserData: []byte("node:" + nodeIDStr + ":in"),
		})

		// Rule: match src IP = node.IP → count outgoing traffic from node
		conn.AddRule(&nftables.Rule{
			Table: table,
			Chain: chain,
			Exprs: []expr.Any{
				&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 12, Len: 4},
				&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ip},
				&expr.Counter{},
			},
			UserData: []byte("node:" + nodeIDStr + ":out"),
		})
	}

	if err := conn.Flush(); err != nil {
		return fmt.Errorf("nftables flush stats: %w", err)
	}

	slog.Info("traffic: node counters set up", "nodes", len(nodes))
	return nil
}

// CollectNodeCounters reads nftables counter values for all node rules.
func (e *Engine) CollectNodeCounters() ([]NodeTrafficSnapshot, error) {
	conn, err := nftables.New()
	if err != nil {
		return nil, fmt.Errorf("nftables connect: %w", err)
	}

	table := &nftables.Table{Name: statsTableName, Family: nftables.TableFamilyIPv4}
	chain := &nftables.Chain{Name: "traffic_count", Table: table}

	rules, err := conn.GetRules(table, chain)
	if err != nil {
		return nil, fmt.Errorf("get rules: %w", err)
	}

	// Aggregate counters by node ID
	type counterPair struct {
		bytesIn  uint64
		bytesOut uint64
	}
	counters := make(map[int64]*counterPair)

	for _, rule := range rules {
		if len(rule.UserData) == 0 {
			continue
		}
		tag := string(rule.UserData)
		// Format: "node:{id}:{in|out}"
		parts := strings.SplitN(tag, ":", 3)
		if len(parts) != 3 || parts[0] != "node" {
			continue
		}

		nodeID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		direction := parts[2]

		// Find counter expression in rule
		var bytes uint64
		for _, e := range rule.Exprs {
			if c, ok := e.(*expr.Counter); ok {
				bytes = c.Bytes
				break
			}
		}

		pair, ok := counters[nodeID]
		if !ok {
			pair = &counterPair{}
			counters[nodeID] = pair
		}

		switch direction {
		case "in":
			pair.bytesIn = bytes
		case "out":
			pair.bytesOut = bytes
		}
	}

	results := make([]NodeTrafficSnapshot, 0, len(counters))
	for nodeID, pair := range counters {
		results = append(results, NodeTrafficSnapshot{
			NodeID:   nodeID,
			BytesIn:  pair.bytesIn,
			BytesOut: pair.bytesOut,
		})
	}

	return results, nil
}

// CleanupNodeCounters removes the stats table.
func (e *Engine) CleanupNodeCounters() error {
	conn, err := nftables.New()
	if err != nil {
		return fmt.Errorf("nftables connect: %w", err)
	}
	conn.DelTable(&nftables.Table{Name: statsTableName, Family: nftables.TableFamilyIPv4})
	return conn.Flush()
}

// CollectInterfaces reads /proc/net/dev and returns per-interface byte counters.
func (e *Engine) CollectInterfaces(names []string) ([]InterfaceSnapshot, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	var results []InterfaceSnapshot
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue // skip header lines
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		ifName := strings.TrimSpace(parts[0])
		if len(nameSet) > 0 && !nameSet[ifName] {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		var bytesIn, bytesOut uint64
		fmt.Sscanf(fields[0], "%d", &bytesIn)
		fmt.Sscanf(fields[8], "%d", &bytesOut)

		results = append(results, InterfaceSnapshot{
			Name:     ifName,
			BytesIn:  bytesIn,
			BytesOut: bytesOut,
		})
	}

	return results, scanner.Err()
}
