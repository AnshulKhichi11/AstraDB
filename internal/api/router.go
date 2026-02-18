package api

import (
	"net/http"
	"fmt"
	"os"

	"testDB/internal/api/handlers"
	"testDB/internal/auth"
	"testDB/internal/engine"
)

func NewRouter(e *engine.Engine) http.Handler {
	mux := http.NewServeMux()

	// Initialize auth system
	authSystem := auth.New()
	rateLimiter := auth.NewRateLimiter()

	h := handlers.New(e)
	authH := handlers.NewAuthHandlers(authSystem)

	// Public endpoints (no auth)
	mux.HandleFunc("/", h.Health)
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/auth/login", authH.Login)

	// Protected endpoints (with auth)
	protected := http.NewServeMux()
	protected.HandleFunc("/api/insert", h.Insert)
	protected.HandleFunc("/api/query", h.Query)
	protected.HandleFunc("/api/update", h.Update)
	protected.HandleFunc("/api/delete", h.Delete)
	protected.HandleFunc("/api/list", h.List)
	protected.HandleFunc("/api/databases", h.Databases)
	protected.HandleFunc("/api/collections", h.Collections)
	protected.HandleFunc("/api/stats", h.Stats)
	protected.HandleFunc("/api/createIndex", h.CreateIndex)
	protected.HandleFunc("/api/compact", h.Compact)
	protected.HandleFunc("/api/segment-stats", h.SegmentStats)

	// Admin-only endpoints
	protected.HandleFunc("/api/auth/users", authSystem.RequireAdmin(authH.ListUsers))
	protected.HandleFunc("/api/auth/create-user", authSystem.RequireAdmin(authH.CreateUser))
	protected.HandleFunc("/api/auth/delete-user", authSystem.RequireAdmin(authH.DeleteUser))

	// User endpoints
	protected.HandleFunc("/api/auth/rotate-key", authH.RotateAPIKey)
	protected.HandleFunc("/api/auth/change-password", authH.ChangePassword)

	// Apply middleware: rate limiting -> auth -> handlers
	//chain := rateLimiter.Middleware(100)(authSystem.Middleware(protected))

	
	// Add to protected endpoints
    protected.HandleFunc("/api/wal-stats", h.WALStats)



	// ─────────────────────────────────────────
	// AUTH MODE CHECK
	// ─────────────────────────────────────────
	// ASTRA_AUTH=true  → Auth required (production)
	// ASTRA_AUTH=false → No auth needed (development) ← DEFAULT
	authEnabled := os.Getenv("ASTRA_AUTH") == "true"

	var chain http.Handler
	if authEnabled {
		// Production: full auth + rate limiting
		chain = rateLimiter.Middleware(100)(authSystem.Middleware(protected))
		fmt.Println("🔐 Auth mode: ENABLED (set ASTRA_AUTH=false to disable)")
	} else {
		// Development: no auth, just rate limiting
		chain = rateLimiter.Middleware(1000)(protected)
		fmt.Println("🟢 Auth mode: DISABLED (set ASTRA_AUTH=true for production)")
	}

	mux.Handle("/api/", chain)

	return mux
}
