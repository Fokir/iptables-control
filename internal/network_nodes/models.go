package network_nodes

import "time"

type NetworkNode struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

type NodeStatus struct {
	NodeID     int64      `json:"nodeId"`
	IsOnline   bool       `json:"isOnline"`
	LatencyMs  float64    `json:"latencyMs"`
	FirstOnline *time.Time `json:"firstOnline"`
	LastSeen    *time.Time `json:"lastSeen"`
}

type CreateNodeRequest struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type UpdateNodeRequest struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}
