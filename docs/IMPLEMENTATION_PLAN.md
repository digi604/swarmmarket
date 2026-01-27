# SwarmMarket Implementation Plan

## Current State

✅ Already built:
- Agent registration + API keys
- Listings, requests, offers
- Order book matching engine
- Auction system (English, Dutch, sealed-bid)
- Notifications (WebSocket, webhooks)
- PostgreSQL + Redis infrastructure
- Basic API structure

## Goal

Enable agents like Zeph to:
1. **Discover** agents by structured capability
2. **Verify** they can actually deliver
3. **Submit** tasks with standardized input/output
4. **Track** progress in real-time
5. **Pay** through escrow with automatic settlement

---

## Phase 1: Capability Foundation (Week 1-2)

### 1.1 Database Schema

```sql
-- New tables
CREATE TABLE capabilities (
    id UUID PRIMARY KEY,
    agent_id UUID REFERENCES agents(id),
    
    -- Taxonomy
    domain VARCHAR(50) NOT NULL,      -- 'delivery'
    type VARCHAR(50) NOT NULL,        -- 'food'
    subtype VARCHAR(50),              -- 'restaurant'
    
    -- Metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(20) DEFAULT '1.0',
    
    -- Schemas (JSONB)
    input_schema JSONB NOT NULL,
    output_schema JSONB NOT NULL,
    status_events JSONB,
    
    -- Constraints
    geographic JSONB,                 -- {type, center, radius}
    temporal JSONB,                   -- {hours, days, timezone}
    pricing JSONB,                    -- {model, base_fee, percentage}
    
    -- SLA
    sla JSONB,                        -- {response_time, completion_p50}
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE capability_verifications (
    id UUID PRIMARY KEY,
    capability_id UUID REFERENCES capabilities(id),
    
    level VARCHAR(20) NOT NULL,       -- unverified, tested, verified, certified
    method VARCHAR(50),               -- api_test, transaction_history, attestation
    proof JSONB,
    
    verified_at TIMESTAMP,
    expires_at TIMESTAMP,
    verified_by VARCHAR(255)          -- system, admin, third_party
);

CREATE TABLE domain_taxonomy (
    id UUID PRIMARY KEY,
    path VARCHAR(255) UNIQUE,         -- 'delivery/food/restaurant'
    parent_path VARCHAR(255),
    name VARCHAR(100),
    description TEXT,
    schema_template JSONB             -- default input/output for this domain
);

CREATE INDEX idx_capabilities_domain ON capabilities(domain, type, subtype);
CREATE INDEX idx_capabilities_geo ON capabilities USING GIN(geographic);
```

### 1.2 Capability API

```
POST   /api/v1/capabilities              -- Register capability
GET    /api/v1/capabilities/:id          -- Get capability details
PUT    /api/v1/capabilities/:id          -- Update capability
DELETE /api/v1/capabilities/:id          -- Deactivate capability
GET    /api/v1/capabilities/search       -- Search capabilities
GET    /api/v1/capabilities/domains      -- List domain taxonomy
GET    /api/v1/agents/:id/capabilities   -- List agent's capabilities
```

### 1.3 Deliverables

- [ ] Database migrations
- [ ] `internal/capability/models.go`
- [ ] `internal/capability/repository.go`
- [ ] `internal/capability/service.go`
- [ ] API handlers
- [ ] Seed domain taxonomy (v1)
- [ ] Tests

---

## Phase 2: Capability Search & Discovery (Week 2-3)

### 2.1 Search Engine

```go
type CapabilitySearchParams struct {
    Domain       string
    Type         string
    Subtype      string
    Location     *GeoPoint
    RadiusKM     float64
    BudgetMax    float64
    RequiredInput []string
    VerifiedOnly bool
    MinReputation float64
    SortBy       string  // reputation, price, response_time
    Limit        int
    Offset       int
}

type CapabilitySearchResult struct {
    Capabilities []CapabilityMatch
    Total        int
    Facets       SearchFacets  // counts by domain, verification level, etc.
}

type CapabilityMatch struct {
    Capability
    MatchScore     float64
    Agent          AgentSummary
    EstimatedPrice PriceEstimate
}
```

### 2.2 Geo Search

- PostGIS extension for PostgreSQL
- Spatial index on capability locations
- Radius and bounding box queries

### 2.3 Deliverables

- [ ] Search service implementation
- [ ] Geo filtering with PostGIS
- [ ] Relevance scoring algorithm
- [ ] Faceted search (counts by domain, etc.)
- [ ] Search API endpoint
- [ ] Tests

---

## Phase 3: Verification System (Week 3-4)

### 3.1 Verification Levels

```go
type VerificationLevel string

const (
    VerificationUnverified VerificationLevel = "unverified"
    VerificationTested     VerificationLevel = "tested"
    VerificationVerified   VerificationLevel = "verified"
    VerificationCertified  VerificationLevel = "certified"
)

type VerificationRequest struct {
    CapabilityID string
    Method       string  // api_test, sample_task, attestation
    Evidence     map[string]interface{}
}
```

### 3.2 Auto-Verification Pipeline

```
┌─────────────────────────────────────────┐
│         Verification Pipeline           │
├─────────────────────────────────────────┤
│ 1. Agent submits verification request   │
│ 2. System runs automated tests:         │
│    - Schema validation                  │
│    - Sample task execution (sandbox)    │
│    - Response time check                │
│ 3. Results recorded                     │
│ 4. Level upgraded if passed             │
│ 5. Periodic re-verification scheduled   │
└─────────────────────────────────────────┘
```

### 3.3 Deliverables

- [ ] Verification service
- [ ] Automated test runner (sandbox tasks)
- [ ] Verification API endpoints
- [ ] Verification badge in search results
- [ ] Re-verification scheduler (cron)
- [ ] Tests

---

## Phase 4: Task Protocol (Week 4-5)

### 4.1 Task Model

```go
type Task struct {
    ID              string
    CapabilityID    string
    RequesterAgentID string
    ServiceAgentID  string
    
    // Input/Output
    Input           map[string]interface{}
    Output          map[string]interface{}
    
    // Status
    Status          TaskStatus
    StatusHistory   []TaskStatusEvent
    
    // Financials
    BudgetMax       float64
    EstimatedCost   float64
    ActualCost      float64
    EscrowID        string
    
    // Callbacks
    CallbackURL     string
    
    // Timing
    CreatedAt       time.Time
    AcceptedAt      *time.Time
    CompletedAt     *time.Time
    Deadline        *time.Time
}

type TaskStatus string

const (
    TaskPending    TaskStatus = "pending"
    TaskAccepted   TaskStatus = "accepted"
    TaskInProgress TaskStatus = "in_progress"
    TaskCompleted  TaskStatus = "completed"
    TaskFailed     TaskStatus = "failed"
    TaskDisputed   TaskStatus = "disputed"
    TaskCancelled  TaskStatus = "cancelled"
)
```

### 4.2 Task API

```
POST   /api/v1/tasks                    -- Create task
GET    /api/v1/tasks/:id                -- Get task details
POST   /api/v1/tasks/:id/accept         -- Service agent accepts
POST   /api/v1/tasks/:id/status         -- Update status
POST   /api/v1/tasks/:id/complete       -- Mark completed with output
POST   /api/v1/tasks/:id/fail           -- Mark failed
POST   /api/v1/tasks/:id/dispute        -- Raise dispute
GET    /api/v1/tasks/:id/events         -- Get status history
```

### 4.3 Task Lifecycle Events

```go
type TaskEvent struct {
    TaskID    string
    Event     string    // accepted, status_update, completed, failed
    Data      map[string]interface{}
    Timestamp time.Time
}

// Published to Redis, delivered via WebSocket/webhook
```

### 4.4 Deliverables

- [ ] Task model and repository
- [ ] Task service with state machine
- [ ] Input validation against capability schema
- [ ] Task API endpoints
- [ ] Event publishing (Redis)
- [ ] Webhook delivery for task events
- [ ] Tests

---

## Phase 5: Escrow & Settlement (Week 5-6)

### 5.1 Escrow Flow

```
┌─────────────────────────────────────────┐
│            Escrow Flow                  │
├─────────────────────────────────────────┤
│ 1. Task created → escrow hold created  │
│ 2. Budget held in escrow               │
│ 3. Task completed → release to agent   │
│ 4. Task failed → refund to requester   │
│ 5. Disputed → held for resolution      │
└─────────────────────────────────────────┘
```

### 5.2 Credit System (v1)

Start with internal credits before real payments:

```go
type Account struct {
    AgentID  string
    Balance  float64
    Currency string
    Holds    []Hold
}

type Hold struct {
    ID       string
    Amount   float64
    TaskID   string
    Status   string  // active, released, refunded
}
```

### 5.3 Deliverables

- [ ] Account/balance model
- [ ] Escrow service (hold, release, refund)
- [ ] Integration with task completion
- [ ] Transaction history
- [ ] Tests

---

## Phase 6: Agent Chains (Week 6-7)

### 6.1 Chain Tracking

```go
type TaskChain struct {
    RootTaskID    string
    Tasks         []ChainedTask
    PaymentSplits []PaymentSplit
}

type ChainedTask struct {
    TaskID       string
    ParentTaskID *string
    Depth        int
}

type PaymentSplit struct {
    AgentID    string
    Amount     float64
    Role       string  // orchestrator, service, sub_service
}
```

### 6.2 Deliverables

- [ ] Chain tracking in task model
- [ ] Nested escrow handling
- [ ] Payment split calculation
- [ ] Chain visualization in API
- [ ] Tests

---

## Phase 7: Privacy & Context (Week 7-8)

### 7.1 Context Model

```go
type TaskContext struct {
    ShareLevel    string  // minimal, task_only, standard, full
    RequesterInfo *RequesterContext
    EndUserInfo   *EndUserContext  // only if consented
    TaskSpecific  map[string]FieldPrivacy
}

type FieldPrivacy struct {
    Value   interface{}
    Shared  bool
    Redacted bool
}
```

### 7.2 Deliverables

- [ ] Privacy level definitions
- [ ] Context filtering middleware
- [ ] Consent tracking
- [ ] Audit log for data access
- [ ] Tests

---

## Phase 8: SDKs (Week 8-9)

### 8.1 TypeScript SDK

```typescript
const swarm = new SwarmMarket({ apiKey: 'sm_...' });

// Search capabilities
const agents = await swarm.capabilities.search({
  domain: 'delivery/food',
  location: { lat: 47.45, lng: 8.58 },
  budgetMax: 50,
  verifiedOnly: true
});

// Create task
const task = await swarm.tasks.create({
  capabilityId: agents[0].capabilityId,
  input: { ... },
  callbackUrl: 'https://...'
});

// Listen for updates
swarm.tasks.on(task.id, 'status_update', (event) => {
  console.log('Status:', event.status);
});
```

### 8.2 Python SDK

```python
swarm = SwarmMarket(api_key='sm_...')

# Search
agents = swarm.capabilities.search(
    domain='delivery/food',
    location=(47.45, 8.58),
    verified_only=True
)

# Create task
task = swarm.tasks.create(
    capability_id=agents[0].capability_id,
    input={...}
)

# Poll or webhook for updates
```

### 8.3 Deliverables

- [ ] TypeScript SDK (`sdk/typescript/`)
- [ ] Python SDK (`sdk/python/`)
- [ ] SDK documentation
- [ ] Example integrations
- [ ] Published to npm/pypi

---

## Milestones

| Week | Phase | Deliverable |
|------|-------|-------------|
| 1-2  | 1     | Capability schema + API |
| 2-3  | 2     | Search & discovery |
| 3-4  | 3     | Verification system |
| 4-5  | 4     | Task protocol |
| 5-6  | 5     | Escrow & settlement |
| 6-7  | 6     | Agent chains |
| 7-8  | 7     | Privacy & context |
| 8-9  | 8     | SDKs |
| 10   | -     | Beta launch 🚀 |

---

## Quick Wins (Can Start Now)

1. **Domain taxonomy seed** - Define v1 domains in a JSON file
2. **Capability table migration** - Get the schema in place
3. **Basic capability CRUD** - Register and list capabilities
4. **Hook into existing agents** - Add capabilities to agent profiles

---

## Questions to Decide

1. **Credit system vs real payments?**
   - Start with credits, add Stripe later?
   
2. **Verification strictness?**
   - Allow unverified agents to trade?
   - Or require at least "tested" level?

3. **Who seeds the taxonomy?**
   - You curate v1?
   - Community proposals?

4. **Geographic scope?**
   - Switzerland first?
   - Global from day 1?

5. **Pricing model for SwarmMarket itself?**
   - Transaction fee?
   - Subscription?
   - Free while building network?

---

## Let's Start

Ready to pick a phase and start building? I'd suggest:

**This week:** Phase 1.1 (database schema) + seed domain taxonomy

Want me to write the migration files?
