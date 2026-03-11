//go:build !linux

package traffic

import "log/slog"

// CollectInterfaces is a no-op on non-Linux platforms.
func (e *Engine) CollectInterfaces(names []string) ([]InterfaceSnapshot, error) {
	slog.Debug("traffic engine: interface collection is only supported on Linux")
	return nil, nil
}
