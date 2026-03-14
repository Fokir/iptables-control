//go:build !linux

package traffic

import "log/slog"

// CollectInterfaces is a no-op on non-Linux platforms.
func (e *Engine) CollectInterfaces(names []string) ([]InterfaceSnapshot, error) {
	slog.Debug("traffic engine: interface collection is only supported on Linux")
	return nil, nil
}

// SetupNodeCounters is a no-op on non-Linux platforms.
func (e *Engine) SetupNodeCounters(nodes []NodeInfo) error {
	slog.Debug("traffic engine: node counters are only supported on Linux")
	return nil
}

// CollectNodeCounters is a no-op on non-Linux platforms.
func (e *Engine) CollectNodeCounters() ([]NodeTrafficSnapshot, error) {
	slog.Debug("traffic engine: node counters are only supported on Linux")
	return nil, nil
}

// CleanupNodeCounters is a no-op on non-Linux platforms.
func (e *Engine) CleanupNodeCounters() error {
	return nil
}
