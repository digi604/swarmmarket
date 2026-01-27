package matching

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OrderSide represents buy or sell.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the type of order.
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"  // Execute at specific price or better
	OrderTypeMarket OrderType = "market" // Execute at best available price
)

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusPartial   OrderStatus = "partial"   // Partially filled
	OrderStatusFilled    OrderStatus = "filled"    // Completely filled
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents a buy or sell order in the order book.
type Order struct {
	ID            uuid.UUID   `json:"id"`
	AgentID       uuid.UUID   `json:"agent_id"`
	ProductID     uuid.UUID   `json:"product_id"`     // What's being traded
	Side          OrderSide   `json:"side"`           // buy or sell
	Type          OrderType   `json:"type"`           // limit or market
	Price         float64     `json:"price"`          // Limit price (0 for market)
	Quantity      float64     `json:"quantity"`       // Original quantity
	FilledQty     float64     `json:"filled_qty"`     // How much has been filled
	RemainingQty  float64     `json:"remaining_qty"`  // Remaining to fill
	Status        OrderStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Trade represents a matched trade between two orders.
type Trade struct {
	ID         uuid.UUID `json:"id"`
	ProductID  uuid.UUID `json:"product_id"`
	BuyOrderID uuid.UUID `json:"buy_order_id"`
	SellOrderID uuid.UUID `json:"sell_order_id"`
	BuyerID    uuid.UUID `json:"buyer_id"`
	SellerID   uuid.UUID `json:"seller_id"`
	Price      float64   `json:"price"`    // Execution price
	Quantity   float64   `json:"quantity"` // Traded quantity
	CreatedAt  time.Time `json:"created_at"`
}

// PriceLevel represents aggregate quantity at a price level.
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Orders   int     `json:"orders"` // Number of orders at this level
}

// OrderBook represents the order book for a single product.
type OrderBook struct {
	ProductID uuid.UUID     `json:"product_id"`
	Bids      []PriceLevel  `json:"bids"`      // Buy orders (highest first)
	Asks      []PriceLevel  `json:"asks"`      // Sell orders (lowest first)
	LastPrice *float64      `json:"last_price"` // Last traded price
	Volume24h float64       `json:"volume_24h"` // 24h trading volume
	High24h   *float64      `json:"high_24h"`
	Low24h    *float64      `json:"low_24h"`
}

// MatchResult contains the result of order matching.
type MatchResult struct {
	Trades        []Trade `json:"trades"`
	RemainingOrder *Order `json:"remaining_order,omitempty"`
}

// EventHandler is called when trades occur.
type EventHandler func(ctx context.Context, trade Trade)

// Engine is the order matching engine.
type Engine struct {
	mu           sync.RWMutex
	buyOrders    map[uuid.UUID][]*Order // ProductID -> orders (sorted by price desc, time asc)
	sellOrders   map[uuid.UUID][]*Order // ProductID -> orders (sorted by price asc, time asc)
	lastPrices   map[uuid.UUID]float64
	eventHandler EventHandler
}

// NewEngine creates a new matching engine.
func NewEngine(handler EventHandler) *Engine {
	return &Engine{
		buyOrders:    make(map[uuid.UUID][]*Order),
		sellOrders:   make(map[uuid.UUID][]*Order),
		lastPrices:   make(map[uuid.UUID]float64),
		eventHandler: handler,
	}
}

// PlaceOrder places an order and attempts to match it.
func (e *Engine) PlaceOrder(ctx context.Context, order *Order) (*MatchResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	order.ID = uuid.New()
	order.RemainingQty = order.Quantity
	order.Status = OrderStatusOpen
	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = order.CreatedAt

	result := &MatchResult{}

	if order.Side == OrderSideBuy {
		result = e.matchBuyOrder(ctx, order)
	} else {
		result = e.matchSellOrder(ctx, order)
	}

	// If order still has remaining quantity, add to book
	if order.RemainingQty > 0 && order.Type == OrderTypeLimit {
		if order.Side == OrderSideBuy {
			e.addBuyOrder(order)
		} else {
			e.addSellOrder(order)
		}
		result.RemainingOrder = order
	}

	return result, nil
}

// matchBuyOrder matches a buy order against sell orders.
func (e *Engine) matchBuyOrder(ctx context.Context, buyOrder *Order) *MatchResult {
	result := &MatchResult{}
	productID := buyOrder.ProductID
	sellOrders := e.sellOrders[productID]

	i := 0
	for i < len(sellOrders) && buyOrder.RemainingQty > 0 {
		sellOrder := sellOrders[i]

		// Check if prices cross (buy price >= sell price for a match)
		if buyOrder.Type == OrderTypeLimit && buyOrder.Price < sellOrder.Price {
			break // No more matches possible (sell orders sorted by price asc)
		}

		// Determine trade quantity and price
		tradeQty := min(buyOrder.RemainingQty, sellOrder.RemainingQty)
		tradePrice := sellOrder.Price // Price-time priority: use resting order's price

		// Create trade
		trade := Trade{
			ID:          uuid.New(),
			ProductID:   productID,
			BuyOrderID:  buyOrder.ID,
			SellOrderID: sellOrder.ID,
			BuyerID:     buyOrder.AgentID,
			SellerID:    sellOrder.AgentID,
			Price:       tradePrice,
			Quantity:    tradeQty,
			CreatedAt:   time.Now().UTC(),
		}
		result.Trades = append(result.Trades, trade)

		// Update quantities
		buyOrder.RemainingQty -= tradeQty
		buyOrder.FilledQty += tradeQty
		sellOrder.RemainingQty -= tradeQty
		sellOrder.FilledQty += tradeQty

		// Update statuses
		if buyOrder.RemainingQty == 0 {
			buyOrder.Status = OrderStatusFilled
		} else {
			buyOrder.Status = OrderStatusPartial
		}

		if sellOrder.RemainingQty == 0 {
			sellOrder.Status = OrderStatusFilled
			// Remove filled order
			sellOrders = append(sellOrders[:i], sellOrders[i+1:]...)
		} else {
			sellOrder.Status = OrderStatusPartial
			i++
		}

		// Update last price
		e.lastPrices[productID] = tradePrice

		// Notify
		if e.eventHandler != nil {
			go e.eventHandler(ctx, trade)
		}
	}

	e.sellOrders[productID] = sellOrders
	return result
}

// matchSellOrder matches a sell order against buy orders.
func (e *Engine) matchSellOrder(ctx context.Context, sellOrder *Order) *MatchResult {
	result := &MatchResult{}
	productID := sellOrder.ProductID
	buyOrders := e.buyOrders[productID]

	i := 0
	for i < len(buyOrders) && sellOrder.RemainingQty > 0 {
		buyOrder := buyOrders[i]

		// Check if prices cross (buy price >= sell price for a match)
		if sellOrder.Type == OrderTypeLimit && buyOrder.Price < sellOrder.Price {
			break // No more matches possible (buy orders sorted by price desc)
		}

		// Determine trade quantity and price
		tradeQty := min(sellOrder.RemainingQty, buyOrder.RemainingQty)
		tradePrice := buyOrder.Price // Price-time priority: use resting order's price

		// Create trade
		trade := Trade{
			ID:          uuid.New(),
			ProductID:   productID,
			BuyOrderID:  buyOrder.ID,
			SellOrderID: sellOrder.ID,
			BuyerID:     buyOrder.AgentID,
			SellerID:    sellOrder.AgentID,
			Price:       tradePrice,
			Quantity:    tradeQty,
			CreatedAt:   time.Now().UTC(),
		}
		result.Trades = append(result.Trades, trade)

		// Update quantities
		sellOrder.RemainingQty -= tradeQty
		sellOrder.FilledQty += tradeQty
		buyOrder.RemainingQty -= tradeQty
		buyOrder.FilledQty += tradeQty

		// Update statuses
		if sellOrder.RemainingQty == 0 {
			sellOrder.Status = OrderStatusFilled
		} else {
			sellOrder.Status = OrderStatusPartial
		}

		if buyOrder.RemainingQty == 0 {
			buyOrder.Status = OrderStatusFilled
			buyOrders = append(buyOrders[:i], buyOrders[i+1:]...)
		} else {
			buyOrder.Status = OrderStatusPartial
			i++
		}

		// Update last price
		e.lastPrices[productID] = tradePrice

		// Notify
		if e.eventHandler != nil {
			go e.eventHandler(ctx, trade)
		}
	}

	e.buyOrders[productID] = buyOrders
	return result
}

// addBuyOrder adds a buy order to the book (sorted by price desc, time asc).
func (e *Engine) addBuyOrder(order *Order) {
	orders := e.buyOrders[order.ProductID]
	orders = append(orders, order)
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price != orders[j].Price {
			return orders[i].Price > orders[j].Price // Higher price first
		}
		return orders[i].CreatedAt.Before(orders[j].CreatedAt) // Earlier time first
	})
	e.buyOrders[order.ProductID] = orders
}

// addSellOrder adds a sell order to the book (sorted by price asc, time asc).
func (e *Engine) addSellOrder(order *Order) {
	orders := e.sellOrders[order.ProductID]
	orders = append(orders, order)
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price != orders[j].Price {
			return orders[i].Price < orders[j].Price // Lower price first
		}
		return orders[i].CreatedAt.Before(orders[j].CreatedAt) // Earlier time first
	})
	e.sellOrders[order.ProductID] = orders
}

// GetOrderBook returns the current order book for a product.
func (e *Engine) GetOrderBook(productID uuid.UUID, depth int) *OrderBook {
	e.mu.RLock()
	defer e.mu.RUnlock()

	book := &OrderBook{
		ProductID: productID,
	}

	// Aggregate bids
	bidLevels := make(map[float64]*PriceLevel)
	for _, order := range e.buyOrders[productID] {
		if level, ok := bidLevels[order.Price]; ok {
			level.Quantity += order.RemainingQty
			level.Orders++
		} else {
			bidLevels[order.Price] = &PriceLevel{
				Price:    order.Price,
				Quantity: order.RemainingQty,
				Orders:   1,
			}
		}
	}
	for _, level := range bidLevels {
		book.Bids = append(book.Bids, *level)
	}
	sort.Slice(book.Bids, func(i, j int) bool {
		return book.Bids[i].Price > book.Bids[j].Price
	})
	if depth > 0 && len(book.Bids) > depth {
		book.Bids = book.Bids[:depth]
	}

	// Aggregate asks
	askLevels := make(map[float64]*PriceLevel)
	for _, order := range e.sellOrders[productID] {
		if level, ok := askLevels[order.Price]; ok {
			level.Quantity += order.RemainingQty
			level.Orders++
		} else {
			askLevels[order.Price] = &PriceLevel{
				Price:    order.Price,
				Quantity: order.RemainingQty,
				Orders:   1,
			}
		}
	}
	for _, level := range askLevels {
		book.Asks = append(book.Asks, *level)
	}
	sort.Slice(book.Asks, func(i, j int) bool {
		return book.Asks[i].Price < book.Asks[j].Price
	})
	if depth > 0 && len(book.Asks) > depth {
		book.Asks = book.Asks[:depth]
	}

	// Last price
	if price, ok := e.lastPrices[productID]; ok {
		book.LastPrice = &price
	}

	return book
}

// CancelOrder cancels an order.
func (e *Engine) CancelOrder(orderID uuid.UUID, agentID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Search in buy orders
	for productID, orders := range e.buyOrders {
		for i, order := range orders {
			if order.ID == orderID {
				if order.AgentID != agentID {
					return fmt.Errorf("not authorized to cancel this order")
				}
				order.Status = OrderStatusCancelled
				e.buyOrders[productID] = append(orders[:i], orders[i+1:]...)
				return nil
			}
		}
	}

	// Search in sell orders
	for productID, orders := range e.sellOrders {
		for i, order := range orders {
			if order.ID == orderID {
				if order.AgentID != agentID {
					return fmt.Errorf("not authorized to cancel this order")
				}
				order.Status = OrderStatusCancelled
				e.sellOrders[productID] = append(orders[:i], orders[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("order not found")
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
