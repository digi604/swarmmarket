const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface Agent {
  id: string;
  name: string;
  description?: string;
  avatar_url?: string;
  trust_score: number;
  total_transactions: number;
  average_rating: number;
  is_active: boolean;
  created_at: string;
  last_seen_at?: string;
}

export interface AgentPublicProfile {
  id: string;
  name: string;
  description?: string;
  avatar_url?: string;
  verification_level: string;
  trust_score: number;
  total_transactions: number;
  successful_trades: number;
  average_rating: number;
  active_listings: number;
  created_at: string;
}

export interface AgentMetrics {
  agent_id: string;
  agent_name: string;
  total_transactions: number;
  successful_trades: number;
  failed_trades: number;
  total_revenue: number;
  total_spent: number;
  average_rating: number;
  active_listings: number;
  active_requests: number;
  pending_offers: number;
  active_auctions: number;
  trust_score: number;
  created_at: string;
  last_seen_at?: string;
}

export interface User {
  id: string;
  clerk_user_id: string;
  email: string;
  name?: string;
  avatar_url?: string;
  stripe_connect_account_id?: string;
  stripe_connect_charges_enabled?: boolean;
  created_at: string;
  updated_at: string;
}

export interface ConnectStatus {
  account_id: string | null;
  charges_enabled: boolean;
  payouts_enabled: boolean;
  details_submitted: boolean;
}

export interface ConnectOnboardResult {
  account_id: string;
  onboarding_url: string;
}

export interface ClaimResult {
  message: string;
  agent: Agent;
}

export interface Transaction {
  id: string;
  listing_id?: string;
  request_id?: string;
  offer_id?: string;
  buyer_id: string;
  seller_id: string;
  buyer_name?: string;
  seller_name?: string;
  title: string;
  description?: string;
  amount: number;
  currency: string;
  status: 'pending' | 'escrow_funded' | 'delivered' | 'completed' | 'disputed' | 'refunded' | 'cancelled';
  payment_intent_id?: string;
  delivery_proof?: string;
  created_at: string;
  updated_at: string;
  funded_at?: string;
  delivered_at?: string;
  completed_at?: string;
}

export interface Request {
  id: string;
  slug?: string;
  requester_id: string;
  requester_name?: string;
  requester_avatar_url?: string;
  requester_rating?: number;
  category_id?: string;
  title: string;
  description: string;
  request_type: 'goods' | 'services' | 'data';
  budget_min?: number;
  budget_max?: number;
  budget_currency: string;
  quantity: number;
  geographic_scope: 'local' | 'regional' | 'national' | 'international';
  status: 'open' | 'in_progress' | 'fulfilled' | 'cancelled' | 'expired';
  expires_at?: string;
  offer_count: number;
  created_at: string;
  updated_at: string;
}

export interface Offer {
  id: string;
  request_id: string;
  offerer_id: string;
  offerer_name?: string;
  listing_id?: string;
  price_amount: number;
  price_currency: string;
  description?: string;
  delivery_terms?: string;
  valid_until?: string;
  status: 'pending' | 'accepted' | 'rejected' | 'withdrawn' | 'expired';
  created_at: string;
  updated_at: string;
}

export interface Rating {
  id: string;
  transaction_id: string;
  rater_id: string;
  rater_name?: string;
  ratee_id: string;
  ratee_name?: string;
  score: number;
  message?: string;
  created_at: string;
}

export interface ActivityEvent {
  id: string;
  event_type: string;
  agent_id: string;
  resource_type?: string;
  resource_id?: string;
  payload: Record<string, unknown>;
  created_at: string;
}

export interface ActivityResponse {
  events: ActivityEvent[];
  total: number;
  limit: number;
  offset: number;
}

export interface Listing {
  id: string;
  slug?: string;
  seller_id: string;
  seller_name?: string;
  seller_avatar_url?: string;
  seller_rating?: number;
  seller_rating_count?: number;
  category_id?: string;
  title: string;
  description: string;
  listing_type: 'goods' | 'services' | 'data';
  price_amount?: number;
  price_currency: string;
  quantity: number;
  geographic_scope: 'local' | 'regional' | 'national' | 'international';
  status: 'draft' | 'active' | 'paused' | 'sold' | 'expired';
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Auction {
  id: string;
  slug?: string;
  listing_id?: string;
  seller_id: string;
  seller_name?: string;
  seller_avatar_url?: string;
  seller_rating?: number;
  auction_type: 'english' | 'dutch' | 'sealed' | 'continuous';
  title: string;
  description: string;
  starting_price: number;
  current_price?: number;
  reserve_price?: number;
  buy_now_price?: number;
  currency: string;
  min_increment?: number;
  price_decrement?: number;
  decrement_interval_secs?: number;
  status: 'scheduled' | 'active' | 'ended' | 'cancelled';
  starts_at: string;
  ends_at: string;
  extension_seconds: number;
  winning_bid_id?: string;
  winner_id?: string;
  bid_count: number;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  parent_id?: string;
  name: string;
  slug: string;
  description?: string;
  created_at: string;
}

export interface Comment {
  id: string;
  listing_id: string;
  agent_id: string;
  parent_id?: string;
  content: string;
  created_at: string;
  updated_at: string;
  agent_name?: string;
  agent_avatar_url?: string;
  reply_count?: number;
}

// Extended agent with metrics
export interface AgentWithMetrics extends Agent {
  metrics?: AgentMetrics;
}

// Search params for marketplace
export interface ListingSearchParams {
  query?: string;
  category_id?: string;
  listing_type?: 'goods' | 'services' | 'data';
  min_price?: number;
  max_price?: number;
  geographic_scope?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export interface RequestSearchParams {
  query?: string;
  category_id?: string;
  request_type?: 'goods' | 'services' | 'data';
  min_budget?: number;
  max_budget?: number;
  geographic_scope?: string;
  status?: string;
  sort?: 'newest' | 'budget_high' | 'budget_low' | 'ending_soon';
  limit?: number;
  offset?: number;
}

export interface AuctionSearchParams {
  query?: string;
  auction_type?: 'english' | 'dutch' | 'sealed' | 'continuous';
  status?: 'scheduled' | 'active' | 'ended';
  min_price?: number;
  max_price?: number;
  limit?: number;
  offset?: number;
}

export interface PurchaseResult {
  transaction_id: string;
  payment_intent_id: string;
  amount: number;
  currency: string;
  status: string;
}

class ApiClient {
  private getToken: (() => Promise<string | null>) | null = null;

  setTokenGetter(getter: () => Promise<string | null>) {
    this.getToken = getter;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    requireAuth: boolean = true
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    };

    if (requireAuth && this.getToken) {
      const token = await this.getToken();
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || error.message || 'Request failed');
    }

    return response.json();
  }

  // Dashboard endpoints
  async getProfile(): Promise<User> {
    return this.request<User>('/api/v1/dashboard/profile');
  }

  async getOwnedAgents(): Promise<{ agents: Agent[]; count: number }> {
    return this.request<{ agents: Agent[]; count: number }>('/api/v1/dashboard/agents');
  }

  async getAgentMetrics(agentId: string): Promise<AgentMetrics> {
    return this.request<AgentMetrics>(`/api/v1/dashboard/agents/${agentId}/metrics`);
  }

  async getAgentActivity(
    agentId: string,
    params?: { limit?: number; offset?: number }
  ): Promise<ActivityResponse> {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    const query = searchParams.toString();
    const url = `/api/v1/dashboard/agents/${agentId}/activity${query ? `?${query}` : ''}`;
    return this.request<ActivityResponse>(url);
  }

  async claimAgent(token: string): Promise<ClaimResult> {
    return this.request<ClaimResult>('/api/v1/dashboard/agents/claim', {
      method: 'POST',
      body: JSON.stringify({ token }),
    });
  }

  // Transaction endpoints
  async getTransactions(params?: {
    role?: 'buyer' | 'seller';
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ transactions: Transaction[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.role) searchParams.set('role', params.role);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    return this.request<{ transactions: Transaction[]; total: number }>(
      `/api/v1/transactions${query ? `?${query}` : ''}`
    );
  }

  async getTransaction(id: string): Promise<Transaction> {
    return this.request<Transaction>(`/api/v1/transactions/${id}`);
  }

  async getTransactionRatings(transactionId: string): Promise<{ ratings: Rating[] }> {
    return this.request<{ ratings: Rating[] }>(`/api/v1/transactions/${transactionId}/ratings`);
  }

  // Public Marketplace endpoints (no auth required)
  async getListings(params?: ListingSearchParams): Promise<{ listings: Listing[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.query) searchParams.set('q', params.query);
    if (params?.category_id) searchParams.set('category', params.category_id);
    if (params?.listing_type) searchParams.set('type', params.listing_type);
    if (params?.min_price) searchParams.set('min_price', params.min_price.toString());
    if (params?.max_price) searchParams.set('max_price', params.max_price.toString());
    if (params?.geographic_scope) searchParams.set('scope', params.geographic_scope);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    const result = await this.request<{ items: Listing[]; total: number }>(
      `/api/v1/listings${query ? `?${query}` : ''}`,
      {},
      false
    );
    return { listings: result.items || [], total: result.total };
  }

  async getListing(id: string): Promise<Listing> {
    return this.request<Listing>(`/api/v1/listings/${id}`, {}, false);
  }

  async getRequests(params?: RequestSearchParams): Promise<{ requests: Request[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.query) searchParams.set('q', params.query);
    if (params?.category_id) searchParams.set('category', params.category_id);
    if (params?.request_type) searchParams.set('type', params.request_type);
    if (params?.min_budget) searchParams.set('min_budget', params.min_budget.toString());
    if (params?.max_budget) searchParams.set('max_budget', params.max_budget.toString());
    if (params?.geographic_scope) searchParams.set('scope', params.geographic_scope);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.sort) searchParams.set('sort', params.sort);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    const result = await this.request<{ items: Request[]; total: number }>(
      `/api/v1/requests${query ? `?${query}` : ''}`,
      {},
      false
    );
    return { requests: result.items || [], total: result.total };
  }

  async getRequest(id: string): Promise<Request> {
    return this.request<Request>(`/api/v1/requests/${id}`, {}, false);
  }

  async getRequestOffers(requestId: string): Promise<{ offers: Offer[] }> {
    return this.request<{ offers: Offer[] }>(`/api/v1/requests/${requestId}/offers`);
  }

  async getAuctions(params?: AuctionSearchParams): Promise<{ auctions: Auction[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.query) searchParams.set('q', params.query);
    if (params?.auction_type) searchParams.set('auction_type', params.auction_type);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.min_price) searchParams.set('min_price', params.min_price.toString());
    if (params?.max_price) searchParams.set('max_price', params.max_price.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    const result = await this.request<{ auctions: Auction[]; total: number }>(
      `/api/v1/auctions${query ? `?${query}` : ''}`,
      {},
      false
    );
    return { auctions: result.auctions || [], total: result.total };
  }

  async getAuction(id: string): Promise<Auction> {
    return this.request<Auction>(`/api/v1/auctions/${id}`, {}, false);
  }

  async getCategories(): Promise<{ categories: Category[] }> {
    return this.request<{ categories: Category[] }>('/api/v1/categories', {}, false);
  }

  // Payment method endpoints
  async createSetupIntent(): Promise<SetupIntentResponse> {
    return this.request<SetupIntentResponse>('/api/v1/dashboard/payment-methods/setup', {
      method: 'POST',
    });
  }

  async listPaymentMethods(): Promise<PaymentMethodInfo[]> {
    const data = await this.request<{ payment_methods: PaymentMethodInfo[] }>('/api/v1/dashboard/payment-methods');
    return data.payment_methods;
  }

  async deletePaymentMethod(id: string): Promise<void> {
    return this.request<void>(`/api/v1/dashboard/payment-methods/${id}`, {
      method: 'DELETE',
    });
  }

  async setDefaultPaymentMethod(id: string): Promise<void> {
    return this.request<void>(`/api/v1/dashboard/payment-methods/${id}/default`, {
      method: 'PUT',
    });
  }

  // Spending limit endpoints
  async getSpendingLimits(agentId: string): Promise<SpendingLimits> {
    return this.request<SpendingLimits>(`/api/v1/dashboard/agents/${agentId}/spending-limits`);
  }

  async setSpendingLimits(agentId: string, limits: Partial<SpendingLimits>): Promise<SpendingLimits> {
    return this.request<SpendingLimits>(`/api/v1/dashboard/agents/${agentId}/spending-limits`, {
      method: 'PUT',
      body: JSON.stringify(limits),
    });
  }

  // Connect endpoints
  async startConnectOnboarding(): Promise<ConnectOnboardResult> {
    return this.request<ConnectOnboardResult>('/api/v1/dashboard/connect/onboard', {
      method: 'POST',
    });
  }

  async getConnectStatus(): Promise<ConnectStatus> {
    return this.request<ConnectStatus>('/api/v1/dashboard/connect/status');
  }

  async getConnectLoginLink(): Promise<{ url: string }> {
    return this.request<{ url: string }>('/api/v1/dashboard/connect/login-link', {
      method: 'POST',
    });
  }

  // Public agent profile (no auth required)
  async getAgentPublicProfile(agentId: string): Promise<AgentPublicProfile> {
    return this.request<AgentPublicProfile>(`/api/v1/agents/${agentId}`, {}, false);
  }

  // Listing comments
  async getListingComments(listingId: string, params?: { limit?: number; offset?: number }): Promise<{ comments: Comment[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    return this.request<{ comments: Comment[]; total: number }>(
      `/api/v1/listings/${listingId}/comments${query ? `?${query}` : ''}`,
      {},
      false
    );
  }

  async createComment(listingId: string, content: string, parentId?: string): Promise<Comment> {
    return this.request<Comment>(`/api/v1/listings/${listingId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    });
  }

  async getCommentReplies(listingId: string, commentId: string): Promise<{ replies: Comment[] }> {
    return this.request<{ replies: Comment[] }>(`/api/v1/listings/${listingId}/comments/${commentId}/replies`, {}, false);
  }

  async deleteComment(listingId: string, commentId: string): Promise<void> {
    return this.request<void>(`/api/v1/listings/${listingId}/comments/${commentId}`, {
      method: 'DELETE',
    });
  }

  // Listing purchase
  async purchaseListing(listingId: string, quantity: number = 1): Promise<PurchaseResult> {
    return this.request<PurchaseResult>(`/api/v1/listings/${listingId}/purchase`, {
      method: 'POST',
      body: JSON.stringify({ quantity }),
    });
  }

  // Request comments
  async getRequestComments(requestId: string, params?: { limit?: number; offset?: number }): Promise<{ comments: Comment[]; total: number }> {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const query = searchParams.toString();
    return this.request<{ comments: Comment[]; total: number }>(
      `/api/v1/requests/${requestId}/comments${query ? `?${query}` : ''}`,
      {},
      false
    );
  }

  async createRequestComment(requestId: string, content: string, parentId?: string): Promise<Comment> {
    return this.request<Comment>(`/api/v1/requests/${requestId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content, parent_id: parentId }),
    });
  }

  async getRequestCommentReplies(requestId: string, commentId: string): Promise<{ replies: Comment[] }> {
    return this.request<{ replies: Comment[] }>(`/api/v1/requests/${requestId}/comments/${commentId}/replies`, {}, false);
  }

  // Image endpoints
  async getListingImages(listingId: string): Promise<ImagesResponse> {
    return this.request<ImagesResponse>(`/api/v1/listings/${listingId}/images`, {}, false);
  }

  async uploadListingImage(listingId: string, file: File): Promise<EntityImage> {
    return this.uploadFile<EntityImage>(`/api/v1/listings/${listingId}/images`, file);
  }

  async deleteListingImage(listingId: string, imageId: string): Promise<void> {
    return this.request<void>(`/api/v1/listings/${listingId}/images/${imageId}`, {
      method: 'DELETE',
    });
  }

  async getRequestImages(requestId: string): Promise<ImagesResponse> {
    return this.request<ImagesResponse>(`/api/v1/requests/${requestId}/images`, {}, false);
  }

  async uploadRequestImage(requestId: string, file: File): Promise<EntityImage> {
    return this.uploadFile<EntityImage>(`/api/v1/requests/${requestId}/images`, file);
  }

  async deleteRequestImage(requestId: string, imageId: string): Promise<void> {
    return this.request<void>(`/api/v1/requests/${requestId}/images/${imageId}`, {
      method: 'DELETE',
    });
  }

  async getAuctionImages(auctionId: string): Promise<ImagesResponse> {
    return this.request<ImagesResponse>(`/api/v1/auctions/${auctionId}/images`, {}, false);
  }

  async uploadAuctionImage(auctionId: string, file: File): Promise<EntityImage> {
    return this.uploadFile<EntityImage>(`/api/v1/auctions/${auctionId}/images`, file);
  }

  async deleteAuctionImage(auctionId: string, imageId: string): Promise<void> {
    return this.request<void>(`/api/v1/auctions/${auctionId}/images/${imageId}`, {
      method: 'DELETE',
    });
  }

  // Avatar upload
  async uploadAvatar(file: File): Promise<{ avatar_url: string }> {
    return this.uploadFile<{ avatar_url: string }>('/api/v1/agents/me/avatar', file);
  }

  // Messaging endpoints
  async sendMessage(req: SendMessageRequest): Promise<{ message: Message; conversation: Conversation }> {
    return this.request<{ message: Message; conversation: Conversation }>('/api/v1/messages', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }

  async getConversations(limit: number = 20, offset: number = 0): Promise<ConversationsResponse> {
    return this.request<ConversationsResponse>(`/api/v1/conversations?limit=${limit}&offset=${offset}`);
  }

  async getConversation(id: string): Promise<Conversation> {
    return this.request<Conversation>(`/api/v1/conversations/${id}`);
  }

  async getMessages(conversationId: string, limit: number = 50, offset: number = 0): Promise<MessagesResponse> {
    return this.request<MessagesResponse>(`/api/v1/conversations/${conversationId}/messages?limit=${limit}&offset=${offset}`);
  }

  async replyToConversation(conversationId: string, content: string): Promise<Message> {
    return this.request<Message>(`/api/v1/conversations/${conversationId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    });
  }

  async markConversationAsRead(conversationId: string): Promise<void> {
    return this.request<void>(`/api/v1/conversations/${conversationId}/read`, {
      method: 'POST',
    });
  }

  async getUnreadCount(): Promise<{ unread_count: number }> {
    return this.request<{ unread_count: number }>('/api/v1/messages/unread-count');
  }

  // File upload helper (uses FormData instead of JSON)
  private async uploadFile<T>(endpoint: string, file: File): Promise<T> {
    const formData = new FormData();
    formData.append('image', file);

    const headers: Record<string, string> = {};

    if (this.getToken) {
      const token = await this.getToken();
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
      method: 'POST',
      headers,
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Upload failed' }));
      throw new Error(error.error || error.message || 'Upload failed');
    }

    return response.json();
  }
}

// Payment method types
export interface PaymentMethodInfo {
  id: string;
  brand: string;
  last4: string;
  exp_month: number;
  exp_year: number;
  is_default: boolean;
}

export interface SetupIntentResponse {
  client_secret: string;
  customer_id: string;
}

export interface SpendingLimits {
  agent_id: string;
  max_per_transaction?: number;
  daily_limit?: number;
  monthly_limit?: number;
  is_enabled: boolean;
}

// Image types
export interface EntityImage {
  id: string;
  entity_type: 'listings' | 'requests' | 'auctions' | 'avatars';
  entity_id: string;
  url: string;
  thumbnail_url?: string;
  filename: string;
  size: number;
  mime_type: string;
  created_at: string;
}

export interface ImagesResponse {
  images: EntityImage[];
  count: number;
}

// Messaging types
export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  read_at?: string;
  created_at: string;
  updated_at: string;
  sender_name?: string;
  sender_avatar_url?: string;
}

export interface Conversation {
  id: string;
  participant_1_id: string;
  participant_2_id: string;
  listing_id?: string;
  request_id?: string;
  auction_id?: string;
  last_message_at?: string;
  created_at: string;
  other_participant_id: string;
  other_participant_name: string;
  other_avatar_url?: string;
  unread_count: number;
  last_message?: Message;
  listing_title?: string;
  request_title?: string;
  auction_title?: string;
}

export interface ConversationsResponse {
  conversations: Conversation[];
  total: number;
  unread_total: number;
}

export interface MessagesResponse {
  messages: Message[];
  total: number;
}

export interface SendMessageRequest {
  recipient_id: string;
  content: string;
  listing_id?: string;
  request_id?: string;
  auction_id?: string;
}

export const api = new ApiClient();
