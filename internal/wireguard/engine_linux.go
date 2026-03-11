//go:build linux

package wireguard

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	wgConfPath  = "/etc/wireguard/wg0.conf"
	wgInterface = "wg0"
)

// IsInstalled checks if WireGuard tools are available.
func (e *Engine) IsInstalled() bool {
	_, err := exec.LookPath("wg")
	return err == nil
}

// Install installs WireGuard via apt.
func (e *Engine) Install() error {
	cmd := exec.Command("apt-get", "install", "-y", "wireguard", "wireguard-tools", "qrencode")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install wireguard: %w", err)
	}
	slog.Info("wireguard installed")
	return nil
}

// IsRunning checks if the WireGuard interface is up.
func (e *Engine) IsRunning() bool {
	out, err := exec.Command("wg", "show", wgInterface).CombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}

// Start starts the WireGuard interface.
func (e *Engine) Start() error {
	out, err := exec.Command("systemctl", "start", "wg-quick@"+wgInterface).CombinedOutput()
	if err != nil {
		return fmt.Errorf("start wg: %s: %w", string(out), err)
	}
	_ = exec.Command("systemctl", "enable", "wg-quick@"+wgInterface).Run()
	slog.Info("wireguard started")
	return nil
}

// Stop stops the WireGuard interface.
func (e *Engine) Stop() error {
	out, err := exec.Command("systemctl", "stop", "wg-quick@"+wgInterface).CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop wg: %s: %w", string(out), err)
	}
	slog.Info("wireguard stopped")
	return nil
}

// GenerateKeyPair generates a WireGuard private/public key pair.
func (e *Engine) GenerateKeyPair() (privateKey, publicKey string, err error) {
	privOut, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return "", "", fmt.Errorf("genkey: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOut))

	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	pubOut, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("pubkey: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOut))
	return privateKey, publicKey, nil
}

// GeneratePresharedKey generates a WireGuard preshared key.
func (e *Engine) GeneratePresharedKey() (string, error) {
	out, err := exec.Command("wg", "genpsk").Output()
	if err != nil {
		return "", fmt.Errorf("genpsk: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetServerInfo reads the server interface info from config and runtime.
func (e *Engine) GetServerInfo() (publicKey string, listenPort int, address string, endpoint string, err error) {
	// Read config file for address and endpoint
	data, err := os.ReadFile(wgConfPath)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Address") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				address = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "ListenPort") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				listenPort, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
		if strings.HasPrefix(line, "# ENDPOINT") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) == 3 {
				endpoint = strings.TrimSpace(parts[2])
			}
		}
	}

	// Get public key from runtime
	out, err := exec.Command("wg", "show", wgInterface, "public-key").Output()
	if err == nil {
		publicKey = strings.TrimSpace(string(out))
	}

	return publicKey, listenPort, address, endpoint, nil
}

// GetPeerStatus returns runtime status for all peers (keyed by public key).
func (e *Engine) GetPeerStatus() (map[string]PeerRuntime, error) {
	out, err := exec.Command("wg", "show", wgInterface, "dump").Output()
	if err != nil {
		return nil, fmt.Errorf("wg show dump: %w", err)
	}

	result := make(map[string]PeerRuntime)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	first := true
	for scanner.Scan() {
		if first {
			first = false
			continue // skip header (interface line)
		}
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 8 {
			continue
		}
		pubKey := fields[0]
		var handshake string
		if fields[4] != "0" {
			ts, _ := strconv.ParseInt(fields[4], 10, 64)
			if ts > 0 {
				handshake = fmt.Sprintf("%d", ts)
			}
		}
		rx, _ := strconv.ParseInt(fields[5], 10, 64)
		tx, _ := strconv.ParseInt(fields[6], 10, 64)
		endpoint := fields[2]
		if endpoint == "(none)" {
			endpoint = ""
		}

		result[pubKey] = PeerRuntime{
			Endpoint:        endpoint,
			LatestHandshake: handshake,
			TransferRx:      rx,
			TransferTx:      tx,
		}
	}
	return result, nil
}

// PeerRuntime holds runtime info for a peer.
type PeerRuntime struct {
	Endpoint        string
	LatestHandshake string
	TransferRx      int64
	TransferTx      int64
}

// AddPeer adds a peer to the WireGuard config and applies it live.
func (e *Engine) AddPeer(name, publicKey, presharedKey, allowedIPs string) error {
	block := fmt.Sprintf("\n# BEGIN_PEER %s\n[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s\n# END_PEER %s\n",
		name, publicKey, presharedKey, allowedIPs, name)

	f, err := os.OpenFile(wgConfPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("write peer block: %w", err)
	}

	// Apply live without restarting
	if err := e.syncConf(); err != nil {
		return err
	}

	slog.Info("wireguard peer added", "name", name)
	return nil
}

// RemovePeer removes a peer from the WireGuard config and applies it live.
func (e *Engine) RemovePeer(name string) error {
	data, err := os.ReadFile(wgConfPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	re := regexp.MustCompile(`(?ms)\n?# BEGIN_PEER ` + regexp.QuoteMeta(name) + `\n.*?# END_PEER ` + regexp.QuoteMeta(name) + `\n?`)
	newData := re.ReplaceAll(data, []byte("\n"))

	if err := os.WriteFile(wgConfPath, newData, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := e.syncConf(); err != nil {
		return err
	}

	slog.Info("wireguard peer removed", "name", name)
	return nil
}

// UpdatePeerInConfig updates a peer's AllowedIPs in the config.
func (e *Engine) UpdatePeerInConfig(oldName, newName, publicKey, presharedKey, allowedIPs string) error {
	data, err := os.ReadFile(wgConfPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	re := regexp.MustCompile(`(?ms)# BEGIN_PEER ` + regexp.QuoteMeta(oldName) + `\n.*?# END_PEER ` + regexp.QuoteMeta(oldName))
	replacement := fmt.Sprintf("# BEGIN_PEER %s\n[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s\n# END_PEER %s",
		newName, publicKey, presharedKey, allowedIPs, newName)

	newData := re.ReplaceAll(data, []byte(replacement))

	if err := os.WriteFile(wgConfPath, newData, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := e.syncConf(); err != nil {
		return err
	}

	slog.Info("wireguard peer updated", "name", newName)
	return nil
}

// NextAvailableIP finds the next available IP in the WireGuard subnet.
func (e *Engine) NextAvailableIP() (string, error) {
	data, err := os.ReadFile(wgConfPath)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}

	// Find all used IPs
	re := regexp.MustCompile(`AllowedIPs\s*=\s*([\d.]+)/`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	used := make(map[string]bool)
	for _, m := range matches {
		used[m[1]] = true
	}

	// Get base subnet from Address
	addrRe := regexp.MustCompile(`Address\s*=\s*([\d.]+)\.(\d+)/`)
	addrMatch := addrRe.FindStringSubmatch(string(data))
	if addrMatch == nil {
		return "", fmt.Errorf("cannot parse server address from config")
	}

	// Parse the base: e.g., "10.7.0" from "10.7.0.1/24"
	parts := strings.Split(addrMatch[1], ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid address format")
	}
	base := strings.Join(parts[:3], ".")
	serverLast, _ := strconv.Atoi(parts[3])
	used[fmt.Sprintf("%s.%d", base, serverLast)] = true

	for i := 2; i < 255; i++ {
		ip := fmt.Sprintf("%s.%d", base, i)
		if !used[ip] {
			return ip, nil
		}
	}
	return "", fmt.Errorf("no available IPs in subnet")
}

// GenerateQRCode generates a QR code PNG for the given config string.
func (e *Engine) GenerateQRCode(config string) ([]byte, error) {
	cmd := exec.Command("qrencode", "-t", "PNG", "-o", "-")
	cmd.Stdin = strings.NewReader(config)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("qrencode: %w", err)
	}
	return out, nil
}

// InitializeConfig creates the initial WireGuard configuration.
func (e *Engine) InitializeConfig(privateKey string, listenPort int, address string, endpoint string) error {
	content := fmt.Sprintf(`# Do not alter the commented lines
# They are used by wireguard-install
# ENDPOINT %s

[Interface]
Address = %s
PrivateKey = %s
ListenPort = %d
`, endpoint, address, privateKey, listenPort)

	if err := os.MkdirAll("/etc/wireguard", 0700); err != nil {
		return fmt.Errorf("create wireguard dir: %w", err)
	}

	if err := os.WriteFile(wgConfPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Info("wireguard config initialized")
	return nil
}

// syncConf applies the current config without restarting the interface.
func (e *Engine) syncConf() error {
	// Strip the interface-level config and create a temp file for wg syncconf
	data, err := os.ReadFile(wgConfPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// wg syncconf needs only the [Peer] sections
	var peerSections []string
	lines := strings.Split(string(data), "\n")
	inPeer := false
	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[Peer]" {
			inPeer = true
			current = []string{line}
			continue
		}
		if trimmed == "[Interface]" {
			if inPeer && len(current) > 0 {
				peerSections = append(peerSections, strings.Join(current, "\n"))
			}
			inPeer = false
			current = nil
			continue
		}
		if inPeer {
			// Skip comment lines in peer sections for syncconf
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			current = append(current, line)
		}
	}
	if inPeer && len(current) > 0 {
		peerSections = append(peerSections, strings.Join(current, "\n"))
	}

	tmpFile, err := os.CreateTemp("", "wg-syncconf-*.conf")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	content := strings.Join(peerSections, "\n\n")
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	out, err := exec.Command("wg", "syncconf", wgInterface, tmpFile.Name()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg syncconf: %s: %w", string(out), err)
	}

	return nil
}
