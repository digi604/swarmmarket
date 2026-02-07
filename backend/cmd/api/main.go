package main

import (
	"context"
	"log"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/capability"
	"github.com/digi604/swarmmarket/backend/internal/config"
	"github.com/digi604/swarmmarket/backend/internal/database"
	"github.com/digi604/swarmmarket/backend/internal/email"
	"github.com/digi604/swarmmarket/backend/internal/marketplace"
	"github.com/digi604/swarmmarket/backend/internal/matching"
	"github.com/digi604/swarmmarket/backend/internal/messaging"
	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/storage"
	"github.com/digi604/swarmmarket/backend/internal/task"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/internal/wallet"
	"github.com/digi604/swarmmarket/backend/internal/worker"
	"github.com/digi604/swarmmarket/backend/pkg/api"
	"github.com/digi604/swarmmarket/backend/pkg/logger"
	"github.com/digi604/swarmmarket/backend/pkg/websocket"
	"github.com/google/uuid"
)

func main() {
	// Start Axiom logger
	logger.StartPeriodicFlush()
	defer logger.FlushLogs()

	logger.Info("SwarmMarket starting", map[string]interface{}{
		"version": "1.0.0",
	})

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

	// Run migrations
	if err := database.RunMigrations(ctx, db.Pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

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

	// Initialize notification repository and service (implements EventPublisher for marketplace)
	// The repository handles event persistence to the database for activity logging
	notificationRepo := notification.NewRepository(db.Pool)
	notificationService := notification.NewServiceWithRepo(redis.Client, notificationRepo)

	// Initialize marketplace service
	marketplaceRepo := marketplace.NewRepository(db.Pool)
	marketplaceService := marketplace.NewService(marketplaceRepo, notificationService)

	// Initialize capability service
	capabilityRepo := capability.NewRepository(db.Pool)
	capabilityService := capability.NewService(capabilityRepo)

	// Initialize task service (capability-linked task execution)
	taskRepo := task.NewRepository(db.Pool)
	capabilityAdapter := task.NewCapabilityAdapter(capabilityService)
	taskService := task.NewService(taskRepo, capabilityAdapter, notificationService)
	taskService.SetSchemaValidator(task.NewJSONSchemaValidator())
	taskService.SetCallbackDeliverer(task.NewAsyncCallbackDeliverer())
	taskService.SetCapabilityStatsUpdater(task.NewCapabilityStatsAdapter(capabilityService))
	log.Println("Task service initialized")

	// Initialize transaction service
	transactionRepo := transaction.NewRepository(db.Pool)
	transactionService := transaction.NewService(transactionRepo, notificationService)

	// Wire transaction creator to marketplace and task services (avoids circular dependency)
	marketplaceService.SetTransactionCreator(transactionService)
	marketplaceService.SetListingTransactionCreator(transactionService)
	taskService.SetTransactionCreator(transactionService)

	// Initialize auction service
	auctionRepo := auction.NewRepository(db.Pool)
	auctionService := auction.NewService(auctionRepo, notificationService)

	// webhookRepo uses the same notification repository for webhook management
	webhookRepo := notificationRepo

	// Initialize WebSocket hub for real-time notifications
	wsHub := websocket.NewHub()
	go wsHub.Run(context.Background())
	go websocket.BridgeRedisToHub(context.Background(), redis.Client, wsHub)
	log.Println("WebSocket hub started")


	// Initialize matching engine (NYSE-style order book)
	matchingEngine := matching.NewEngine(func(ctx context.Context, trade matching.Trade) {
		// Publish trade events
		notificationService.Publish(ctx, "match.found", map[string]any{
			"trade_id":   trade.ID,
			"product_id": trade.ProductID,
			"buyer_id":   trade.BuyerID,
			"seller_id":  trade.SellerID,
			"price":      trade.Price,
			"quantity":   trade.Quantity,
		})
	})
	log.Println("Matching engine initialized")

	// Initialize payment service (Stripe)
	var paymentService *payment.Service
	if cfg.Stripe.SecretKey != "" {
		paymentService = payment.NewService(payment.Config{
			SecretKey:          cfg.Stripe.SecretKey,
			WebhookSecret:      cfg.Stripe.WebhookSecret,
			PlatformFeePercent: cfg.Stripe.PlatformFeePercent,
			DefaultReturnURL:   cfg.Stripe.DefaultReturnURL,
		})
		// Wire payment adapter to transaction service for escrow
		paymentAdapter := payment.NewAdapter(paymentService)
		transactionService.SetPaymentService(paymentAdapter)
		marketplaceService.SetPaymentCreator(paymentAdapter)
		log.Println("Stripe payment service initialized and wired to transactions and marketplace")
	} else {
		log.Println("Stripe not configured - payment endpoints disabled")
	}

	// Initialize user service (for human dashboard)
	var userService *user.Service
	if cfg.Clerk.SecretKey != "" {
		userRepo := user.NewRepository(db.Pool)
		userService = user.NewService(userRepo)
		log.Println("User service initialized (Clerk authentication enabled)")
	} else {
		log.Println("Clerk not configured - dashboard endpoints disabled")
	}

	// Initialize wallet service (for deposits)
	var walletService *wallet.Service
	if cfg.Stripe.SecretKey != "" {
		walletRepo := wallet.NewRepository(db.Pool)
		walletService = wallet.NewService(walletRepo, wallet.StripeConfig{
			SecretKey: cfg.Stripe.SecretKey,
		})
		log.Println("Wallet service initialized")

		// Wire wallet balance checker to marketplace and auction services
		// This enforces that agents have sufficient funds before accepting offers or placing bids
		balanceChecker := wallet.NewBalanceChecker(walletService)
		marketplaceService.SetWalletChecker(balanceChecker)
		auctionService.SetWalletChecker(balanceChecker)
		log.Println("Wallet balance checker wired to marketplace and auction services")
	}

	// Initialize storage service (Cloudflare R2 for images)
	var storageService *storage.Service
	var imageRepo *storage.Repository
	if cfg.Storage.R2AccountID != "" && cfg.Storage.R2AccessKeyID != "" {
		var err error
		storageService, err = storage.NewService(storage.Config{
			AccountID:       cfg.Storage.R2AccountID,
			AccessKeyID:     cfg.Storage.R2AccessKeyID,
			SecretAccessKey: cfg.Storage.R2SecretAccessKey,
			BucketName:      cfg.Storage.R2BucketName,
			PublicURL:       cfg.Storage.R2PublicURL,
			MaxFileSizeMB:   cfg.Storage.MaxFileSizeMB,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize storage service: %v", err)
		} else {
			imageRepo = storage.NewRepository(db.Pool)
			log.Println("Storage service initialized (Cloudflare R2)")
		}
	} else {
		log.Println("R2 storage not configured - image upload endpoints disabled")
	}

	// Initialize email service (SendGrid)
	var emailService *email.Service
	if cfg.Email.SendGridAPIKey != "" {
		emailRepo := email.NewRepository(db.Pool)
		emailService = email.NewService(
			cfg.Email.SendGridAPIKey,
			cfg.Email.FromEmail,
			cfg.Email.FromName,
			cfg.Server.PublicURL,
			cfg.Email.CooldownMinutes,
			emailRepo,
		)
		log.Println("Email service initialized (SendGrid)")
	} else {
		log.Println("SendGrid not configured - email notifications disabled")
	}

	// Initialize messaging service with email and agent adapters
	messagingRepo := messaging.NewRepository(db.Pool)
	emailAdapter := messaging.NewEmailAdapter(emailService)
	agentAdapter := messaging.NewAgentAdapter(func(ctx context.Context, id uuid.UUID) (messaging.AgentInfo, error) {
		agent, err := agentService.GetByID(ctx, id)
		if err != nil {
			return messaging.AgentInfo{}, err
		}
		return messaging.AgentInfo{
			ID:    agent.ID,
			Name:  agent.Name,
			Email: agent.OwnerEmail,
		}, nil
	})
	messagingService := messaging.NewService(messagingRepo, notificationService, emailAdapter, agentAdapter)
	log.Println("Messaging service initialized")

	// Initialize and start background worker for event processing and webhook delivery
	bgWorker := worker.New(worker.Config{
		NotificationService: notificationService,
		WebhookRepo:         webhookRepo,
		AuctionService:      auctionService,
		AuctionRepo:         auctionRepo,
		EmailService:        emailService,
		RedisClient:         redis.Client,
	})
	go bgWorker.Run(context.Background())
	log.Println("Background worker started (webhook delivery, auction scheduler)")

	// Create router
	router := api.NewRouter(api.RouterConfig{
		Config:              cfg,
		AgentService:        agentService,
		MarketplaceService:  marketplaceService,
		CapabilityService:   capabilityService,
		TransactionService:  transactionService,
		AuctionService:      auctionService,
		MatchingEngine:      matchingEngine,
		PaymentService:      paymentService,
		WalletService:       walletService,
		TaskService:         taskService,
		MessagingService:    messagingService,
		WebhookRepo:         webhookRepo,
		NotificationService: notificationService,
		WebSocketHub:        wsHub,
		UserService:         userService,
		StorageService:      storageService,
		ImageRepo:           imageRepo,
		DB:                  db,
		Redis:               redis,
	})

	// Create and run server
	server := api.NewServer(cfg.Server, router)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
