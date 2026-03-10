//go:build !linux

package nftables

import "log/slog"

// Apply is a no-op on non-Linux platforms.
func (e *Engine) Apply(groups []NatGroup) error {
	slog.Warn("nftables engine is only supported on Linux, skipping rule application")
	return nil
}
