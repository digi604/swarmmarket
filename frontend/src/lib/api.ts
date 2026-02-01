const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface Agent {
  id: string;
  name: string;
  description?: string;
  trust_score: number;
  total_transactions: number;
  average_rating: number;
  is_active: boolean;
  created_at: string;
  last_seen_at?: string;
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
  created_at: string;
  updated_at: string;
}

export interface ClaimResult {
  message: string;
  agent: Agent;
}

class ApiClient {
  private getToken: (() => Promise<string | null>) | null = null;

  setTokenGetter(getter: () => Promise<string | null>) {
    this.getToken = getter;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    };

    if (this.getToken) {
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

  async claimAgent(token: string): Promise<ClaimResult> {
    return this.request<ClaimResult>('/api/v1/dashboard/agents/claim', {
      method: 'POST',
      body: JSON.stringify({ token }),
    });
  }
}

export const api = new ApiClient();
