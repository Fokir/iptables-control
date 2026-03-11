package wireguard

import "time"

// Status represents the current state of the WireGuard interface.
type Status struct {
	Installed   bool   `json:"installed"`
	Running     bool   `json:"running"`
	PublicKey   string `json:"publicKey,omitempty"`
	ListenPort  int    `json:"listenPort,omitempty"`
	Address     string `json:"address,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	PeerCount   int    `json:"peerCount"`
}

// Peer represents a WireGuard peer (client).
type Peer struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	PublicKey    string    `json:"publicKey"`
	PresharedKey string    `json:"-"`
	PrivateKey   string    `json:"-"`
	AllowedIPs   string    `json:"allowedIps"`
	Address      string    `json:"address"`
	DNS          string    `json:"dns"`
	Endpoint     string    `json:"endpoint,omitempty"`
	LatestHandshake string `json:"latestHandshake,omitempty"`
	TransferRx   int64     `json:"transferRx,omitempty"`
	TransferTx   int64     `json:"transferTx,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CreatePeerRequest is the request body for creating a new peer.
type CreatePeerRequest struct {
	Name       string `json:"name"`
	AllowedIPs string `json:"allowedIps,omitempty"`
	DNS        string `json:"dns,omitempty"`
}

// UpdatePeerRequest is the request body for updating a peer.
type UpdatePeerRequest struct {
	Name       string `json:"name"`
	AllowedIPs string `json:"allowedIps,omitempty"`
	DNS        string `json:"dns,omitempty"`
}

// PeerConfig is the full client configuration for download.
type PeerConfig struct {
	Config string `json:"config"`
	QRCode string `json:"qrCode,omitempty"`
}
