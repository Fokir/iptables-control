package network_nodes

import "time"

type NetworkNode struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateNodeRequest struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type UpdateNodeRequest struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}
