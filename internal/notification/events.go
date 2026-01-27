package notification

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event.
type EventType string

const (
	// Request/Offer events
	EventRequestCreated  EventType = "request.created"
	EventOfferReceived   EventType = "offer.received"
	EventOfferAccepted   EventType = "offer.accepted"
	EventOfferRejected   EventType = "offer.rejected"

	// Listing events
	EventListingCreated EventType = "listing.created"
	EventListingUpdated EventType = "listing.updated"

	// Auction events
	EventAuctionStarted    EventType = "auction.started"
	EventBidPlaced         EventType = "bid.placed"
	EventBidOutbid         EventType = "bid.outbid"
	EventAuctionEndingSoon EventType = "auction.ending_soon"
	EventAuctionEnded      EventType = "auction.ended"

	// Order/Transaction events
	EventOrderCreated      EventType = "order.created"
	EventEscrowFunded      EventType = "escrow.funded"
	EventDeliveryConfirmed EventType = "delivery.confirmed"
	EventPaymentReleased   EventType = "payment.released"
	EventDisputeOpened     EventType = "dispute.opened"

	// Matching events (NYSE-style)
	EventMatchFound EventType = "match.found"
	EventOrderFilled EventType = "order.filled"
)

// Event represents a notification event.
type Event struct {
	ID        uuid.UUID      `json:"id"`
	Type      EventType      `json:"type"`
	AgentID   uuid.UUID      `json:"agent_id"`   // Who should receive this
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}

// Subscription represents an agent's notification preferences.
type Subscription struct {
	ID         uuid.UUID   `json:"id"`
	AgentID    uuid.UUID   `json:"agent_id"`
	EventTypes []EventType `json:"event_types"` // Which events to receive
	Categories []uuid.UUID `json:"categories"`  // Filter by categories
	Keywords   []string    `json:"keywords"`    // Filter by keywords
	MinBudget  *float64    `json:"min_budget"`  // Only requests above this
	MaxBudget  *float64    `json:"max_budget"`  // Only requests below this
	Scopes     []string    `json:"scopes"`      // Geographic scopes
	IsActive   bool        `json:"is_active"`
	CreatedAt  time.Time   `json:"created_at"`
}

// Webhook represents a registered webhook endpoint.
type Webhook struct {
	ID              uuid.UUID   `json:"id"`
	AgentID         uuid.UUID   `json:"agent_id"`
	URL             string      `json:"url"`
	Secret          string      `json:"-"` // For HMAC signature
	EventTypes      []EventType `json:"event_types"`
	IsActive        bool        `json:"is_active"`
	FailureCount    int         `json:"failure_count"`
	LastTriggeredAt *time.Time  `json:"last_triggered_at"`
	CreatedAt       time.Time   `json:"created_at"`
}

// WebhookDelivery tracks webhook delivery attempts.
type WebhookDelivery struct {
	ID           uuid.UUID `json:"id"`
	WebhookID    uuid.UUID `json:"webhook_id"`
	EventID      uuid.UUID `json:"event_id"`
	Attempt      int       `json:"attempt"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body,omitempty"`
	Error        string    `json:"error,omitempty"`
	DeliveredAt  time.Time `json:"delivered_at"`
}
