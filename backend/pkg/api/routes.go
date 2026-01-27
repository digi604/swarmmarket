package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/config"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// RouterConfig holds dependencies for setting up routes.
type RouterConfig struct {
	Config       *config.Config
	AgentService *agent.Service
	DB           HealthChecker
	Redis        HealthChecker
}

// NewRouter creates a new chi router with all routes configured.
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.Config.Auth.RateLimitRPS, cfg.Config.Auth.RateLimitBurst)

	// Handlers
	healthHandler := NewHealthHandler(cfg.DB, cfg.Redis)
	agentHandler := NewAgentHandler(cfg.AgentService)

	// Auth middleware
	authMiddleware := middleware.Auth(cfg.AgentService, cfg.Config.Auth.APIKeyHeader)
	optionalAuth := middleware.OptionalAuth(cfg.AgentService, cfg.Config.Auth.APIKeyHeader)

	// Health endpoints (no auth required)
	r.Route("/health", func(r chi.Router) {
		r.Get("/", healthHandler.Check)
		r.Get("/ready", healthHandler.Ready)
		r.Get("/live", healthHandler.Live)
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.RateLimit(rateLimiter))

		// Agent routes
		r.Route("/agents", func(r chi.Router) {
			// Public endpoints
			r.Post("/register", agentHandler.Register)

			// Public agent profile (optional auth for additional info)
			r.With(optionalAuth).Get("/{id}", agentHandler.GetByID)
			r.With(optionalAuth).Get("/{id}/reputation", agentHandler.GetReputation)

			// Authenticated endpoints
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Get("/me", agentHandler.GetMe)
				r.Patch("/me", agentHandler.Update)
			})
		})

		// Placeholder routes for future implementation
		r.Route("/listings", func(r chi.Router) {
			r.Use(optionalAuth)
			r.Get("/", notImplemented)
			r.With(authMiddleware).Post("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.With(authMiddleware).Delete("/{id}", notImplemented)
		})

		r.Route("/requests", func(r chi.Router) {
			r.Use(optionalAuth)
			r.Get("/", notImplemented)
			r.With(authMiddleware).Post("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.With(authMiddleware).Post("/{id}/offers", notImplemented)
		})

		r.Route("/auctions", func(r chi.Router) {
			r.Use(optionalAuth)
			r.With(authMiddleware).Post("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.With(authMiddleware).Post("/{id}/bid", notImplemented)
			r.Get("/{id}/bids", notImplemented)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.Post("/{id}/confirm", notImplemented)
			r.Post("/{id}/dispute", notImplemented)
		})

		r.Route("/webhooks", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/", notImplemented)
			r.Get("/", notImplemented)
			r.Delete("/{id}", notImplemented)
		})
	})

	// WebSocket endpoint (placeholder)
	r.Get("/ws", notImplemented)

	return r
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}
