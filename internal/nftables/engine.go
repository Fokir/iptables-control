package nftables

// Engine applies NAT rules to the system.
// Platform-specific implementations are in engine_linux.go and engine_stub.go.
type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}
