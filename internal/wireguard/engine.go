package wireguard

// Engine manages WireGuard on the system.
// Platform-specific implementations are in engine_linux.go and engine_stub.go.
type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}
