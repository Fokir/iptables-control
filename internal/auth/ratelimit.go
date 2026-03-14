package auth

import (
	"sync"
	"time"
)

const (
	maxLoginAttempts = 5
	loginWindow      = 15 * time.Minute
)

type loginAttempt struct {
	count    int
	windowAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{attempts: make(map[string]*loginAttempt)}
	go rl.cleanup()
	return rl
}

// Allow checks if the IP is allowed to attempt login.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	a, ok := rl.attempts[ip]
	if !ok {
		rl.attempts[ip] = &loginAttempt{count: 1, windowAt: time.Now().Add(loginWindow)}
		return true
	}

	if time.Now().After(a.windowAt) {
		a.count = 1
		a.windowAt = time.Now().Add(loginWindow)
		return true
	}

	a.count++
	return a.count <= maxLoginAttempts
}

// Reset clears the counter for an IP after successful login.
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, ip)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, a := range rl.attempts {
			if now.After(a.windowAt) {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}
