package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/email"
	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Config holds worker configuration.
type Config struct {
	NotificationService *notification.Service
	WebhookRepo         *notification.Repository
	AuctionService      *auction.Service
	AuctionRepo         *auction.Repository
	EmailService        *email.Service
	RedisClient         *redis.Client
}

// Worker processes background tasks.
type Worker struct {
	notificationService *notification.Service
	webhookRepo         *notification.Repository
	auctionService      *auction.Service
	auctionRepo         *auction.Repository
	emailService        *email.Service
	redis               *redis.Client
}

// New creates a new worker.
func New(cfg Config) *Worker {
	return &Worker{
		notificationService: cfg.NotificationService,
		webhookRepo:         cfg.WebhookRepo,
		auctionService:      cfg.AuctionService,
		auctionRepo:         cfg.AuctionRepo,
		emailService:        cfg.EmailService,
		redis:               cfg.RedisClient,
	}
}

// Run starts all worker goroutines.
func (w *Worker) Run(ctx context.Context) error {
	// Start event consumer
	go w.consumeEvents(ctx)

	// Start auction scheduler
	go w.processAuctions(ctx)

	// Start webhook delivery worker
	go w.deliverWebhooks(ctx)

	// Start email queue processor
	go w.processEmailQueue(ctx)

	<-ctx.Done()
	return nil
}

// consumeEvents listens for events from Redis streams and processes them.
func (w *Worker) consumeEvents(ctx context.Context) {
	log.Println("Worker: Starting event consumer")
	streams := []string{
		"events:request.created",
		"events:request.updated",
		"events:offer.received",
		"events:offer.accepted",
		"events:listing.created",
		"events:listing.purchased",
		"events:comment.created",
		"events:auction.started",
		"events:auction.ending_soon",
		"events:bid.placed",
		"events:bid.outbid",
		"events:auction.ended",
		"events:order.created",
		"events:escrow.funded",
		"events:delivery.confirmed",
		"events:payment.released",
		"events:payment.failed",
		"events:payment.capture_failed",
		"events:transaction.created",
		"events:transaction.escrow_funded",
		"events:transaction.delivered",
		"events:transaction.completed",
		"events:transaction.refunded",
		"events:rating.submitted",
		"events:dispute.opened",
		"events:match.found",
		"events:message.received",
		"events:agent.registered",
		"events:agent.claimed",
	}

	// Initialize stream positions map
	streamPositions := make(map[string]string)
	for _, s := range streams {
		streamPositions[s] = "0" // Start from beginning (will be converted to proper format)
	}

	firstRun := true

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check which streams exist before reading
			existingStreams := make([]string, 0, len(streams))
			for _, s := range streams {
				// Check if stream exists
				exists, err := w.redis.Exists(ctx, s).Result()
				if err == nil && exists > 0 {
					existingStreams = append(existingStreams, s)
				}
			}

			// If no streams exist yet, wait and retry
			if len(existingStreams) == 0 {
				if firstRun {
					log.Println("Worker: No event streams exist yet, waiting...")
					firstRun = false
				}
				time.Sleep(5 * time.Second)
				continue
			}

			if firstRun {
				log.Printf("Worker: Found %d existing streams, starting consumption", len(existingStreams))
				firstRun = false
			}

			// Build stream args only for existing streams
			streamArgs := make([]string, 0, len(existingStreams)*2)
			for _, s := range existingStreams {
				pos := streamPositions[s]
				// For initial position, get the last message ID from the stream
				if pos == "0" {
					// Get last entry to start from there
					lastEntries, err := w.redis.XRevRangeN(ctx, s, "+", "-", 1).Result()
					if err == nil && len(lastEntries) > 0 {
						pos = lastEntries[0].ID
						log.Printf("Worker: Starting stream %s from last message: %s", s, pos)
					} else {
						// Stream exists but is empty, use 0-0
						pos = "0-0"
					}
				}
				streamArgs = append(streamArgs, s, pos)
			}

			// Read from streams with block timeout
			result, err := w.redis.XRead(ctx, &redis.XReadArgs{
				Streams: streamArgs,
				Block:   5 * time.Second,
				Count:   10,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No new messages, continue waiting
					continue
				}
				// Unexpected error
				log.Printf("Worker: Error reading from streams: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("Worker: Received %d stream(s) with events", len(result))
			for _, stream := range result {
				log.Printf("Worker: Processing %d message(s) from %s", len(stream.Messages), stream.Stream)
				for _, msg := range stream.Messages {
					w.processEvent(ctx, stream.Stream, msg)
					// Update stream position to the last processed message
					streamPositions[stream.Stream] = msg.ID
				}
			}
		}
	}
}

// processEvent handles a single event.
func (w *Worker) processEvent(ctx context.Context, stream string, msg redis.XMessage) {
	eventJSON, ok := msg.Values["event"].(string)
	if !ok {
		return
	}

	var event notification.Event
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		log.Printf("Worker: Failed to unmarshal event: %v", err)
		return
	}

	// Get webhooks subscribed to this event type
	webhooks, err := w.webhookRepo.GetActiveWebhooksForEvent(ctx, string(event.Type))
	if err != nil {
		log.Printf("Worker: Failed to get webhooks: %v", err)
		return
	}

	// Queue webhook deliveries
	for _, webhook := range webhooks {
		w.queueWebhookDelivery(ctx, webhook, event)
	}
}

// queueWebhookDelivery queues a webhook for delivery.
func (w *Worker) queueWebhookDelivery(ctx context.Context, webhook *notification.Webhook, event notification.Event) {
	delivery := map[string]any{
		"webhook_id": webhook.ID.String(),
		"event":      event,
		"attempt":    1,
	}

	data, _ := json.Marshal(delivery)
	w.redis.LPush(ctx, "webhook_delivery_queue", data)
}

// deliverWebhooks processes the webhook delivery queue.
func (w *Worker) deliverWebhooks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Pop from queue with timeout
			result, err := w.redis.BRPop(ctx, 5*time.Second, "webhook_delivery_queue").Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				time.Sleep(time.Second)
				continue
			}

			if len(result) < 2 {
				continue
			}

			var delivery struct {
				WebhookID string             `json:"webhook_id"`
				Event     notification.Event `json:"event"`
				Attempt   int                `json:"attempt"`
			}

			if err := json.Unmarshal([]byte(result[1]), &delivery); err != nil {
				log.Printf("Worker: Failed to unmarshal delivery: %v", err)
				continue
			}

			webhookID, err := uuid.Parse(delivery.WebhookID)
			if err != nil {
				continue
			}

			webhook, err := w.webhookRepo.GetWebhookByID(ctx, webhookID)
			if err != nil {
				continue
			}

			// Deliver webhook
			err = w.notificationService.DeliverWebhook(ctx, webhook, delivery.Event)
			if err != nil {
				log.Printf("Worker: Webhook delivery failed (attempt %d): %v", delivery.Attempt, err)
				w.webhookRepo.RecordWebhookFailure(ctx, webhookID)

				// Retry with exponential backoff (max 5 attempts)
				if delivery.Attempt < 5 {
					delivery.Attempt++
					data, _ := json.Marshal(delivery)

					// Delay based on attempt number
					delay := time.Duration(delivery.Attempt*delivery.Attempt) * time.Second
					time.AfterFunc(delay, func() {
						w.redis.LPush(context.Background(), "webhook_delivery_queue", data)
					})
				}
			} else {
				w.webhookRepo.RecordWebhookSuccess(ctx, webhookID)
			}
		}
	}
}

// processAuctions checks for auctions that need to be ended.
func (w *Worker) processAuctions(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.startScheduledAuctions(ctx)
			w.notifyAuctionsEndingSoon(ctx)
			w.endExpiredAuctions(ctx)
		}
	}
}

// notifyAuctionsEndingSoon publishes auction.ending_soon for auctions approaching their end time.
func (w *Worker) notifyAuctionsEndingSoon(ctx context.Context) {
	status := auction.AuctionStatusActive
	result, err := w.auctionRepo.SearchAuctions(ctx, auction.SearchAuctionsParams{
		Status: &status,
		Limit:  200,
	})
	if err != nil {
		log.Printf("Worker: Failed to search active auctions for ending_soon: %v", err)
		return
	}

	now := time.Now().UTC()
	threshold := 5 * time.Minute

	for _, auc := range result.Auctions {
		if auc.EndsAt.Before(now) {
			continue
		}

		remaining := auc.EndsAt.Sub(now)
		if remaining > threshold {
			continue
		}

		if w.redis == nil {
			continue
		}

		key := fmt.Sprintf("auction:%s:ending_soon", auc.ID)
		ok, err := w.redis.SetNX(ctx, key, "1", remaining+time.Minute).Result()
		if err != nil {
			log.Printf("Worker: Failed to set ending_soon key for auction %s: %v", auc.ID, err)
			continue
		}
		if !ok {
			continue
		}

		if w.notificationService != nil {
			_ = w.notificationService.Publish(ctx, "auction.ending_soon", map[string]any{
				"auction_id":   auc.ID,
				"seller_id":    auc.SellerID,
				"auction_type": auc.AuctionType,
				"title":        auc.Title,
				"ends_at":      auc.EndsAt,
			})
		}
	}
}

// startScheduledAuctions activates auctions whose start time has passed.
func (w *Worker) startScheduledAuctions(ctx context.Context) {
	status := auction.AuctionStatusScheduled
	result, err := w.auctionRepo.SearchAuctions(ctx, auction.SearchAuctionsParams{
		Status: &status,
		Limit:  100,
	})
	if err != nil {
		log.Printf("Worker: Failed to search scheduled auctions: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, auc := range result.Auctions {
		if auc.StartsAt.After(now) {
			continue
		}

		if err := w.auctionRepo.UpdateAuctionStatus(ctx, auc.ID, auction.AuctionStatusActive); err != nil {
			log.Printf("Worker: Failed to activate auction %s: %v", auc.ID, err)
			continue
		}

		// Publish event now that the auction is active
		if w.notificationService != nil {
			_ = w.notificationService.Publish(ctx, "auction.started", map[string]any{
				"auction_id":     auc.ID,
				"seller_id":      auc.SellerID,
				"auction_type":   auc.AuctionType,
				"title":          auc.Title,
				"starting_price": auc.StartingPrice,
				"ends_at":        auc.EndsAt,
			})
		}
	}
}

// endExpiredAuctions finds and ends auctions past their end time.
func (w *Worker) endExpiredAuctions(ctx context.Context) {
	// Search for active auctions that have ended
	status := auction.AuctionStatusActive
	result, err := w.auctionRepo.SearchAuctions(ctx, auction.SearchAuctionsParams{
		Status: &status,
		Limit:  100,
	})
	if err != nil {
		log.Printf("Worker: Failed to search auctions: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, auc := range result.Auctions {
		if auc.EndsAt.Before(now) {
			// End the auction
			log.Printf("Worker: Ending expired auction %s", auc.ID)

			// Get highest bid
			highestBid, err := w.auctionRepo.GetHighestBid(ctx, auc.ID)
			if err != nil {
				log.Printf("Worker: Failed to get highest bid: %v", err)
				continue
			}

			if highestBid != nil {
				// Check reserve price
				metReserve := true
				if auc.ReservePrice != nil && highestBid.Amount < *auc.ReservePrice {
					metReserve = false
				}

				if metReserve {
					w.auctionRepo.SetAuctionWinner(ctx, auc.ID, highestBid.ID, highestBid.BidderID)
					w.auctionRepo.UpdateBidStatus(ctx, highestBid.ID, auction.BidStatusWon)
				} else {
					w.auctionRepo.UpdateAuctionStatus(ctx, auc.ID, auction.AuctionStatusEnded)
				}
			} else {
				w.auctionRepo.UpdateAuctionStatus(ctx, auc.ID, auction.AuctionStatusEnded)
			}
		}
	}
}

// processEmailQueue processes the email queue periodically.
func (w *Worker) processEmailQueue(ctx context.Context) {
	if w.emailService == nil {
		log.Printf("Worker: Email service not configured, skipping email queue processing")
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.emailService.ProcessQueue(ctx); err != nil {
				log.Printf("Worker: Failed to process email queue: %v", err)
			}
		}
	}
}
