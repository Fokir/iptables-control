package nginx

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo *Repository
}

func NewAuthHandler(repo *Repository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/verify", h.verify)
	r.Get("/login", h.loginPage)
	r.Post("/login", h.loginSubmit)
	return r
}

// verify is called by nginx auth_request. Returns 200 or 401.
func (h *AuthHandler) verify(w http.ResponseWriter, r *http.Request) {
	host := r.Header.Get("X-Original-Host")
	if host == "" {
		host = r.Host
	}

	domain, err := h.repo.GetByHostname(host)
	if err != nil || !domain.AuthEnabled {
		// Domain not found or auth not enabled — allow
		w.WriteHeader(http.StatusOK)
		return
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := verifyCookie(domain.Domain, domain.AuthCookieSecret, cookie.Value); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// loginPage renders the login form.
func (h *AuthHandler) loginPage(w http.ResponseWriter, r *http.Request) {
	domainName := r.URL.Query().Get("domain")
	redirect := sanitizeRedirect(r.URL.Query().Get("redirect"))

	if domainName == "" {
		http.Error(w, "missing domain parameter", http.StatusBadRequest)
		return
	}

	domain, err := h.repo.GetByHostname(domainName)
	if err != nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	page, err := renderLoginPage(domain.Domain, domain.AuthLoginCSS, redirect, "")
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(page)
}

// loginSubmit processes the login form POST.
func (h *AuthHandler) loginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	domainName := r.FormValue("domain")
	username := r.FormValue("username")
	password := r.FormValue("password")
	redirect := sanitizeRedirect(r.FormValue("redirect"))

	if domainName == "" || username == "" || password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	domain, err := h.repo.GetByHostname(domainName)
	if err != nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	// Always run bcrypt to prevent timing-based username enumeration
	passwordMatch := bcrypt.CompareHashAndPassword([]byte(domain.AuthPasswordHash), []byte(password)) == nil
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(domain.AuthUsername)) == 1

	if !usernameMatch || !passwordMatch {
		page, _ := renderLoginPage(domain.Domain, domain.AuthLoginCSS, redirect, "Invalid username or password")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(page)
		return
	}

	maxAge := domain.AuthCookieMaxAge
	if maxAge <= 0 {
		maxAge = defaultCookieMaxAge
	}
	expiresAt := time.Now().Add(time.Duration(maxAge) * time.Second)
	cookieValue := signCookie(domain.Domain, domain.AuthCookieSecret, expiresAt)

	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    cookieValue,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// sanitizeRedirect prevents open redirect attacks by ensuring the redirect URL
// is a relative path on the same host.
func sanitizeRedirect(redirect string) string {
	if redirect == "" || !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		return "/"
	}
	return redirect
}
