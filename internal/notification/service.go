package notification

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Service handles event publishing and notification delivery.
type Service struct {
	redis      *redis.Client
	httpClient *http.Client
}

// NewService creates a new notification service.
func NewService(redisClient *redis.Client) *Service {
	return &Service{
		redis: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Publish publishes an event to Redis streams and triggers notifications.
func (s *Service) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	event := Event{
		ID:        uuid.New(),
		Type:      EventType(eventType),
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis stream for persistence and processing
	streamKey := "events:" + eventType
	_, err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"event": string(eventJSON),
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	// Publish to Redis pub/sub for real-time delivery
	pubsubChannel := "notifications:" + eventType
	if err := s.redis.Publish(ctx, pubsubChannel, eventJSON).Err(); err != nil {
		// Log but don't fail - stream is the primary delivery
		fmt.Printf("failed to publish to pubsub: %v\n", err)
	}

	// If there's a specific agent to notify, publish to their channel
	if agentID, ok := payload["requester_id"].(uuid.UUID); ok {
		s.notifyAgent(ctx, agentID, event)
	}
	if agentID, ok := payload["offerer_id"].(uuid.UUID); ok {
		s.notifyAgent(ctx, agentID, event)
	}

	return nil
}

// notifyAgent sends a notification to a specific agent.
func (s *Service) notifyAgent(ctx context.Context, agentID uuid.UUID, event Event) {
	eventJSON, _ := json.Marshal(event)

	// Publish to agent's personal channel (for WebSocket connections)
	agentChannel := fmt.Sprintf("agent:%s:notifications", agentID.String())
	s.redis.Publish(ctx, agentChannel, eventJSON)
}

// SubscribeToAgent returns a channel for receiving agent notifications.
func (s *Service) SubscribeToAgent(ctx context.Context, agentID uuid.UUID) <-chan Event {
	events := make(chan Event, 100)

	agentChannel := fmt.Sprintf("agent:%s:notifications", agentID.String())
	pubsub := s.redis.Subscribe(ctx, agentChannel)

	go func() {
		defer close(events)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
					events <- event
				}
			}
		}
	}()

	return events
}

// DeliverWebhook delivers an event to a webhook endpoint.
func (s *Service) DeliverWebhook(ctx context.Context, webhook *Webhook, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Create HMAC signature
	signature := s.signPayload(payload, webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SwarmMarket-Signature", signature)
	req.Header.Set("X-SwarmMarket-Event", string(event.Type))
	req.Header.Set("X-SwarmMarket-Delivery", event.ID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// signPayload creates an HMAC-SHA256 signature for webhook payloads.
func (s *Service) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// BroadcastToCategory broadcasts an event to all agents interested in a category.
func (s *Service) BroadcastToCategory(ctx context.Context, categoryID uuid.UUID, event Event) {
	eventJSON, _ := json.Marshal(event)
	categoryChannel := fmt.Sprintf("category:%s:events", categoryID.String())
	s.redis.Publish(ctx, categoryChannel, eventJSON)
}

// BroadcastToScope broadcasts an event to all agents in a geographic scope.
func (s *Service) BroadcastToScope(ctx context.Context, scope string, event Event) {
	eventJSON, _ := json.Marshal(event)
	scopeChannel := fmt.Sprintf("scope:%s:events", scope)
	s.redis.Publish(ctx, scopeChannel, eventJSON)
}
