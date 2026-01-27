package main

import (
	"context"
	"log"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/config"
	"github.com/digi604/swarmmarket/backend/internal/database"
	"github.com/digi604/swarmmarket/backend/pkg/api"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()
	log.Println("Configuration loaded")

	// Create context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	redis, err := database.NewRedisDB(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("Connected to Redis")

	// Initialize services
	agentRepo := agent.NewRepository(db.Pool)
	agentService := agent.NewService(agentRepo, cfg.Auth.APIKeyLength)

	// Create router
	router := api.NewRouter(api.RouterConfig{
		Config:       cfg,
		AgentService: agentService,
		DB:           db,
		Redis:        redis,
	})

	// Create and run server
	server := api.NewServer(cfg.Server, router)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
