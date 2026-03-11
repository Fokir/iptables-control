package traffic

import "time"

// TrafficPoint is a single data point for charts.
type TrafficPoint struct {
	Time     time.Time `json:"t"`
	BytesIn  int64     `json:"in"`
	BytesOut int64     `json:"out"`
}

// TrafficSeries is a time series for a specific source.
type TrafficSeries struct {
	NodeID        *int64  `json:"nodeId,omitempty"`
	NodeName      string  `json:"nodeName,omitempty"`
	InterfaceName string  `json:"interfaceName,omitempty"`
	Points        []TrafficPoint `json:"points"`
}

// InterfaceSnapshot holds raw counter values from /proc/net/dev.
type InterfaceSnapshot struct {
	Name     string
	BytesIn  uint64
	BytesOut uint64
}

// NodeTrafficSnapshot holds per-node nftables counter values.
type NodeTrafficSnapshot struct {
	NodeID   int64
	BytesIn  uint64
	BytesOut uint64
}
