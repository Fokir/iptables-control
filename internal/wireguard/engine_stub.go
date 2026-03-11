//go:build !linux

package wireguard

import "log/slog"

func (e *Engine) IsInstalled() bool {
	slog.Warn("wireguard engine is only supported on Linux")
	return false
}

func (e *Engine) Install() error {
	slog.Warn("wireguard install is only supported on Linux")
	return nil
}

func (e *Engine) IsRunning() bool {
	return false
}

func (e *Engine) Start() error {
	slog.Warn("wireguard start is only supported on Linux")
	return nil
}

func (e *Engine) Stop() error {
	slog.Warn("wireguard stop is only supported on Linux")
	return nil
}

func (e *Engine) GenerateKeyPair() (string, string, error) {
	return "stub-private-key", "stub-public-key", nil
}

func (e *Engine) GeneratePresharedKey() (string, error) {
	return "stub-preshared-key", nil
}

func (e *Engine) GetServerInfo() (publicKey string, listenPort int, address string, endpoint string, err error) {
	return "stub-pubkey", 51820, "10.7.0.1/24", "127.0.0.1", nil
}

// PeerRuntime holds runtime info for a peer.
type PeerRuntime struct {
	Endpoint        string
	LatestHandshake string
	TransferRx      int64
	TransferTx      int64
}

func (e *Engine) GetPeerStatus() (map[string]PeerRuntime, error) {
	return map[string]PeerRuntime{}, nil
}

func (e *Engine) AddPeer(name, publicKey, presharedKey, allowedIPs string) error {
	slog.Warn("wireguard add peer is only supported on Linux")
	return nil
}

func (e *Engine) RemovePeer(name string) error {
	slog.Warn("wireguard remove peer is only supported on Linux")
	return nil
}

func (e *Engine) UpdatePeerInConfig(oldName, newName, publicKey, presharedKey, allowedIPs string) error {
	slog.Warn("wireguard update peer is only supported on Linux")
	return nil
}

func (e *Engine) NextAvailableIP() (string, error) {
	return "10.7.0.2", nil
}

func (e *Engine) GenerateQRCode(config string) ([]byte, error) {
	return []byte("stub-qr"), nil
}

func (e *Engine) InitializeConfig(privateKey string, listenPort int, address string, endpoint string) error {
	slog.Warn("wireguard init config is only supported on Linux")
	return nil
}
