package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sokol/system-control/internal/pkg/httputil"
)

type Handler struct {
	svc *Service
	rl  *RateLimiter
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, rl: NewRateLimiter()}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ip := r.RemoteAddr

	if !h.rl.Allow(ip) {
		slog.Warn("login rate limit exceeded", "ip", ip, "username", req.Username)
		httputil.Error(w, http.StatusTooManyRequests, "too many login attempts, try again later")
		return
	}

	session, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		slog.Warn("login failed", "ip", ip, "username", req.Username)
		httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	h.rl.Reset(ip)
	slog.Info("login successful", "ip", ip, "username", req.Username)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
	})

	httputil.JSON(w, http.StatusOK, LoginResponse{User: user})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		_ = h.svc.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	httputil.JSON(w, http.StatusOK, nil)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	httputil.JSON(w, http.StatusOK, user)
}
