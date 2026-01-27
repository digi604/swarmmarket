package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/swarmmarket/swarmmarket/internal/agent"
	"github.com/swarmmarket/swarmmarket/internal/common"
)

// ContextKey is a type for context keys.
type ContextKey string

const (
	// AgentContextKey is the context key for the authenticated agent.
	AgentContextKey ContextKey = "agent"
)

// Auth creates an authentication middleware.
func Auth(agentService *agent.Service, apiKeyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeader)
			if apiKey == "" {
				// Also check Authorization header with Bearer scheme
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("missing api key"))
				return
			}

			ag, err := agentService.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				if err == agent.ErrAgentNotFound {
					common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("invalid api key"))
					return
				}
				common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("authentication failed"))
				return
			}

			// Add agent to context
			ctx := context.WithValue(r.Context(), AgentContextKey, ag)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAgent retrieves the authenticated agent from the request context.
func GetAgent(ctx context.Context) *agent.Agent {
	ag, ok := ctx.Value(AgentContextKey).(*agent.Agent)
	if !ok {
		return nil
	}
	return ag
}

// OptionalAuth creates an optional authentication middleware.
// It attempts to authenticate but doesn't fail if no credentials are provided.
func OptionalAuth(agentService *agent.Service, apiKeyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeader)
			if apiKey == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey != "" {
				ag, err := agentService.ValidateAPIKey(r.Context(), apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), AgentContextKey, ag)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
