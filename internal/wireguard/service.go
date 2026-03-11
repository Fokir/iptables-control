package wireguard

import (
	"fmt"
	"strings"
)

type Service struct {
	repo   *Repository
	engine *Engine
}

func NewService(repo *Repository, engine *Engine) *Service {
	return &Service{repo: repo, engine: engine}
}

// GetStatus returns the current WireGuard status.
func (s *Service) GetStatus() (*Status, error) {
	st := &Status{
		Installed: s.engine.IsInstalled(),
	}

	if !st.Installed {
		return st, nil
	}

	st.Running = s.engine.IsRunning()
	if !st.Running {
		return st, nil
	}

	pubKey, port, addr, endpoint, err := s.engine.GetServerInfo()
	if err != nil {
		return st, nil
	}
	st.PublicKey = pubKey
	st.ListenPort = port
	st.Address = addr
	st.Endpoint = endpoint

	peers, err := s.repo.GetAll()
	if err == nil {
		st.PeerCount = len(peers)
	}

	return st, nil
}

// Enable installs (if needed) and starts WireGuard.
func (s *Service) Enable(endpoint string, listenPort int, address string) error {
	if !s.engine.IsInstalled() {
		if err := s.engine.Install(); err != nil {
			return fmt.Errorf("install: %w", err)
		}
	}

	// Check if config exists, create if not
	if !s.engine.IsRunning() {
		privKey, _, err := s.engine.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("generate keys: %w", err)
		}

		if endpoint == "" {
			return fmt.Errorf("endpoint is required for initial setup")
		}
		if address == "" {
			address = "10.7.0.1/24"
		}
		if listenPort == 0 {
			listenPort = 51820
		}

		if err := s.engine.InitializeConfig(privKey, listenPort, address, endpoint); err != nil {
			return fmt.Errorf("init config: %w", err)
		}

		if err := s.engine.Start(); err != nil {
			return fmt.Errorf("start: %w", err)
		}
	}

	return nil
}

// Disable stops WireGuard.
func (s *Service) Disable() error {
	if !s.engine.IsRunning() {
		return nil
	}
	return s.engine.Stop()
}

// GetPeers returns all peers with runtime status merged.
func (s *Service) GetPeers() ([]Peer, error) {
	peers, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("get peers: %w", err)
	}

	// Merge runtime status
	runtimeMap, err := s.engine.GetPeerStatus()
	if err == nil {
		for i, p := range peers {
			if rt, ok := runtimeMap[p.PublicKey]; ok {
				peers[i].Endpoint = rt.Endpoint
				peers[i].LatestHandshake = rt.LatestHandshake
				peers[i].TransferRx = rt.TransferRx
				peers[i].TransferTx = rt.TransferTx
			}
		}
	}

	return peers, nil
}

// GetPeer returns a single peer by ID with runtime status.
func (s *Service) GetPeer(id int64) (*Peer, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}

	runtimeMap, err := s.engine.GetPeerStatus()
	if err == nil {
		if rt, ok := runtimeMap[p.PublicKey]; ok {
			p.Endpoint = rt.Endpoint
			p.LatestHandshake = rt.LatestHandshake
			p.TransferRx = rt.TransferRx
			p.TransferTx = rt.TransferTx
		}
	}

	return p, nil
}

// CreatePeer creates a new WireGuard peer.
func (s *Service) CreatePeer(req CreatePeerRequest) (*Peer, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Check for duplicate name
	existing, _ := s.repo.GetByName(req.Name)
	if existing != nil {
		return nil, fmt.Errorf("peer with name %q already exists", req.Name)
	}

	// Generate keys
	privKey, pubKey, err := s.engine.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate keys: %w", err)
	}

	psk, err := s.engine.GeneratePresharedKey()
	if err != nil {
		return nil, fmt.Errorf("generate preshared key: %w", err)
	}

	// Get next available IP
	nextIP, err := s.engine.NextAvailableIP()
	if err != nil {
		return nil, fmt.Errorf("next available IP: %w", err)
	}

	allowedIPs := req.AllowedIPs
	if allowedIPs == "" {
		allowedIPs = nextIP + "/32"
	}

	dns := req.DNS
	if dns == "" {
		dns = "1.1.1.1, 1.0.0.1"
	}

	peer := &Peer{
		Name:         req.Name,
		PublicKey:    pubKey,
		PresharedKey: psk,
		PrivateKey:   privKey,
		AllowedIPs:   allowedIPs,
		Address:      nextIP + "/32",
		DNS:          dns,
	}

	if err := s.repo.Create(peer); err != nil {
		return nil, fmt.Errorf("save peer: %w", err)
	}

	// Add to WireGuard config
	if err := s.engine.AddPeer(peer.Name, peer.PublicKey, peer.PresharedKey, allowedIPs); err != nil {
		// Rollback DB
		_ = s.repo.Delete(peer.ID)
		return nil, fmt.Errorf("add to wireguard: %w", err)
	}

	return peer, nil
}

// UpdatePeer updates an existing peer.
func (s *Service) UpdatePeer(id int64, req UpdatePeerRequest) (*Peer, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}

	peer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}

	// Check for name conflicts
	if req.Name != peer.Name {
		existing, _ := s.repo.GetByName(req.Name)
		if existing != nil {
			return nil, fmt.Errorf("peer with name %q already exists", req.Name)
		}
	}

	oldName := peer.Name
	peer.Name = req.Name
	if req.AllowedIPs != "" {
		peer.AllowedIPs = req.AllowedIPs
	}
	if req.DNS != "" {
		peer.DNS = req.DNS
	}

	if err := s.repo.Update(peer); err != nil {
		return nil, fmt.Errorf("update peer: %w", err)
	}

	if err := s.engine.UpdatePeerInConfig(oldName, peer.Name, peer.PublicKey, peer.PresharedKey, peer.AllowedIPs); err != nil {
		return nil, fmt.Errorf("update wireguard config: %w", err)
	}

	return peer, nil
}

// DeletePeer removes a peer.
func (s *Service) DeletePeer(id int64) error {
	peer, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("peer not found: %w", err)
	}

	if err := s.engine.RemovePeer(peer.Name); err != nil {
		return fmt.Errorf("remove from wireguard: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete from db: %w", err)
	}

	return nil
}

// GetPeerConfig generates the client configuration for a peer.
func (s *Service) GetPeerConfig(id int64) (*PeerConfig, error) {
	peer, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}

	pubKey, _, _, endpoint, err := s.engine.GetServerInfo()
	if err != nil {
		return nil, fmt.Errorf("get server info: %w", err)
	}

	_, listenPort, _, _, _ := s.engine.GetServerInfo()

	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
PresharedKey = %s
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = %s:%d
PersistentKeepalive = 25
`, peer.PrivateKey, peer.Address, peer.DNS, pubKey, peer.PresharedKey, endpoint, listenPort)

	return &PeerConfig{Config: config}, nil
}

// GetPeerQRCode generates a QR code for the peer configuration.
func (s *Service) GetPeerQRCode(id int64) ([]byte, error) {
	pc, err := s.GetPeerConfig(id)
	if err != nil {
		return nil, err
	}

	return s.engine.GenerateQRCode(pc.Config)
}
