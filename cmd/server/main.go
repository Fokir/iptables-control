package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/sokol/system-control/internal/audit"
	"github.com/sokol/system-control/internal/auth"
	"github.com/sokol/system-control/internal/config"
	"github.com/sokol/system-control/internal/database"
	"github.com/sokol/system-control/internal/network_nodes"
	"github.com/sokol/system-control/internal/nftables"
	"github.com/sokol/system-control/internal/nginx"
	"github.com/sokol/system-control/internal/traffic"
	"github.com/sokol/system-control/internal/wireguard"
)

var version = "dev"

//go:embed all:frontend
var frontendFS embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		os.Exit(0)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("starting system-control", "version", version)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	authRepo := auth.NewRepository(db)
	natRepo := nftables.NewRepository(db)
	nodesRepo := network_nodes.NewRepository(db)
	nginxRepo := nginx.NewRepository(db)
	auditRepo := audit.NewRepository(db)
	trafficRepo := traffic.NewRepository(db)
	wgRepo := wireguard.NewRepository(db)

	// Services
	authSvc := auth.NewService(authRepo, cfg.SessionMaxAge)
	natEngine := nftables.NewEngine()
	natSvc := nftables.NewService(natRepo, natEngine)
	nodesMonitor := network_nodes.NewMonitor(nodesRepo, 30*time.Second)
	nodesSvc := network_nodes.NewService(nodesRepo, nodesMonitor)
	nginxSvc := nginx.NewService(nginxRepo, cfg.NginxSitesDir)
	auditSvc := audit.NewService(auditRepo)
	trafficSvc := traffic.NewService(trafficRepo, nodesRepo)
	wgEngine := wireguard.NewEngine()
	wgSvc := wireguard.NewService(wgRepo, wgEngine)

	// Traffic collector
	var trafficInterfaces []string
	if cfg.TrafficInterfaces != "" {
		for _, s := range strings.Split(cfg.TrafficInterfaces, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				trafficInterfaces = append(trafficInterfaces, s)
			}
		}
	}
	trafficEngine := traffic.NewEngine()
	trafficCollector := traffic.NewCollector(trafficRepo, trafficEngine, nodesRepo, trafficInterfaces, time.Duration(cfg.TrafficInterval)*time.Second)

	// Ensure admin user exists
	if err := authSvc.EnsureAdminUser(cfg.AdminUser, cfg.AdminPassword); err != nil {
		slog.Error("failed to ensure admin user", "error", err)
		os.Exit(1)
	}

	// Sync nftables rules on startup
	if err := natSvc.SyncRules(); err != nil {
		slog.Error("failed to sync nftables rules on startup", "error", err)
	}

	// Sync nginx configs on startup
	if err := nginxSvc.SyncConfigs(); err != nil {
		slog.Error("failed to sync nginx configs on startup", "error", err)
	}

	// Sync wireguard peers from config on startup
	if imported, err := wgSvc.SyncPeers(); err != nil {
		slog.Error("failed to sync wireguard peers on startup", "error", err)
	} else if imported > 0 {
		slog.Info("imported wireguard peers from config", "count", imported)
	}

	// Handlers
	authHandler := auth.NewHandler(authSvc)
	natHandler := nftables.NewHandler(natSvc)
	nodesHandler := network_nodes.NewHandler(nodesSvc)
	nginxHandler := nginx.NewHandler(nginxSvc)
	auditHandler := audit.NewHandler(auditSvc)
	trafficHandler := traffic.NewHandler(trafficSvc)
	wgHandler := wireguard.NewHandler(wgSvc)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Compress(5))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public
		r.Post("/auth/login", authHandler.Login)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(authSvc))
			r.Use(audit.Middleware(auditSvc))
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/auth/me", authHandler.Me)
			r.Mount("/nat-groups", natHandler.Routes())
			r.Mount("/network-nodes", nodesHandler.Routes())
			r.Mount("/nginx-domains", nginxHandler.Routes())
			r.Mount("/audit-logs", auditHandler.Routes())
			r.Mount("/traffic", trafficHandler.Routes())
			r.Mount("/wireguard", wgHandler.Routes())
		})
	})

	// Serve frontend SPA
	r.Handle("/*", spaHandler())

	// Start node monitoring
	nodesMonitor.Start()

	// Start traffic collection
	trafficCollector.Start()

	// Background: cleanup expired sessions and old audit logs
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			_ = authSvc.CleanupExpiredSessions()
			_ = auditSvc.Cleanup(90)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func spaHandler() http.Handler {
	subFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = strings.TrimPrefix(path, "/")
		}

		// Try to open the file
		f, err := subFS.Open(path)
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
