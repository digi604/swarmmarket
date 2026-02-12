import { useState, useEffect, useCallback, useMemo } from 'react';
import { useAuth } from '@clerk/clerk-react';
import { api } from '../lib/api';
import type {
  Agent,
  AgentMetrics,
  AgentWithMetrics,
  User,
  Transaction,
  Rating,
  ActivityEvent,
} from '../lib/api';

export function useApiSetup() {
  const { getToken, isLoaded } = useAuth();

  useEffect(() => {
    if (!isLoaded) return;
    api.setTokenGetter(getToken);
  }, [isLoaded]); // eslint-disable-line react-hooks/exhaustive-deps -- getToken is stable after isLoaded

  return { isReady: isLoaded };
}

// Hook to check if auth is ready
export function useAuthReady() {
  const { isLoaded, isSignedIn } = useAuth();
  return isLoaded && isSignedIn;
}

export function useProfile() {
  const isAuthReady = useAuthReady();
  const [profile, setProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthReady) return;

    api
      .getProfile()
      .then(setProfile)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [isAuthReady]);

  return { profile, loading, error };
}

export function useOwnedAgents() {
  const isAuthReady = useAuthReady();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(() => {
    if (!isAuthReady) return;
    setLoading(true);
    setError(null);
    api
      .getOwnedAgents()
      .then((data) => setAgents(data.agents || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [isAuthReady]);

  useEffect(() => {
    if (isAuthReady) {
      refetch();
    }
  }, [isAuthReady, refetch]);

  return { agents, loading, error, refetch };
}

export function useAgentMetrics(agentId: string | null) {
  const isAuthReady = useAuthReady();
  const [metrics, setMetrics] = useState<AgentMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!agentId || !isAuthReady) {
      setMetrics(null);
      return;
    }

    setLoading(true);
    setError(null);
    api
      .getAgentMetrics(agentId)
      .then(setMetrics)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [agentId, isAuthReady]);

  return { metrics, loading, error };
}

export function useClaimAgent() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const claimAgent = useCallback(async (token: string) => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.claimAgent(token);
      return result;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to claim agent';
      setError(message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  return { claimAgent, loading, error, clearError: () => setError(null) };
}

// Fetch agents with their metrics
export function useAgentsWithMetrics() {
  const { agents, loading: agentsLoading, error: agentsError, refetch } = useOwnedAgents();
  const [agentsWithMetrics, setAgentsWithMetrics] = useState<AgentWithMetrics[]>([]);
  const [metricsLoading, setMetricsLoading] = useState(false);

  useEffect(() => {
    if (agents.length === 0) {
      setAgentsWithMetrics([]);
      return;
    }

    setMetricsLoading(true);

    Promise.all(
      agents.map(async (agent) => {
        try {
          const metrics = await api.getAgentMetrics(agent.id);
          return { ...agent, metrics };
        } catch {
          return { ...agent, metrics: undefined };
        }
      })
    )
      .then(setAgentsWithMetrics)
      .finally(() => setMetricsLoading(false));
  }, [agents]);

  return {
    agents: agentsWithMetrics,
    loading: agentsLoading || metricsLoading,
    error: agentsError,
    refetch,
  };
}

// Fetch all transactions for the user
export function useTransactions(role?: 'buyer' | 'seller') {
  const isAuthReady = useAuthReady();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(() => {
    if (!isAuthReady) return;
    setLoading(true);
    setError(null);
    api
      .getTransactions({ role, limit: 100 })
      .then((data) => setTransactions(data.transactions || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [role, isAuthReady]);

  useEffect(() => {
    if (isAuthReady) {
      refetch();
    }
  }, [isAuthReady, refetch]);

  return { transactions, loading, error, refetch };
}

// Wallet data calculated from transactions
export function useWallet() {
  const isAuthReady = useAuthReady();
  const [buyerTx, setBuyerTx] = useState<Transaction[]>([]);
  const [sellerTx, setSellerTx] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthReady) return;

    setLoading(true);
    setError(null);

    Promise.all([
      api.getTransactions({ role: 'buyer', limit: 100 }),
      api.getTransactions({ role: 'seller', limit: 100 }),
    ])
      .then(([buyerData, sellerData]) => {
        setBuyerTx(buyerData.transactions || []);
        setSellerTx(sellerData.transactions || []);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [isAuthReady]);

  const walletData = useMemo(() => {
    // Calculate balance: completed seller transactions - completed buyer transactions
    const completedSellerAmount = sellerTx
      .filter((tx) => tx.status === 'completed')
      .reduce((sum, tx) => sum + tx.amount, 0);

    const completedBuyerAmount = buyerTx
      .filter((tx) => tx.status === 'completed')
      .reduce((sum, tx) => sum + tx.amount, 0);

    const balance = completedSellerAmount - completedBuyerAmount;

    // Pending: escrow_funded where user is seller (awaiting delivery)
    const pending = sellerTx
      .filter((tx) => tx.status === 'escrow_funded' || tx.status === 'delivered')
      .reduce((sum, tx) => sum + tx.amount, 0);

    // In escrow: escrow_funded where user is buyer
    const inEscrow = buyerTx
      .filter((tx) => tx.status === 'escrow_funded')
      .reduce((sum, tx) => sum + tx.amount, 0);

    // All transactions combined and sorted by date
    const allTransactions = [...buyerTx, ...sellerTx].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );

    // This month earnings
    const now = new Date();
    const thisMonth = sellerTx
      .filter((tx) => {
        const txDate = new Date(tx.completed_at || tx.created_at);
        return (
          tx.status === 'completed' &&
          txDate.getMonth() === now.getMonth() &&
          txDate.getFullYear() === now.getFullYear()
        );
      })
      .reduce((sum, tx) => sum + tx.amount, 0);

    // Total transaction count
    const totalCount = allTransactions.length;

    // Average transaction
    const avgTransaction = totalCount > 0 ? (completedSellerAmount + completedBuyerAmount) / totalCount : 0;

    return {
      balance,
      pending,
      inEscrow,
      thisMonth,
      totalCount,
      avgTransaction,
      transactions: allTransactions,
      buyerTransactions: buyerTx,
      sellerTransactions: sellerTx,
    };
  }, [buyerTx, sellerTx]);

  return { ...walletData, loading, error };
}

// Activity aggregated from transactions and agents
export interface ActivityItem {
  id: string;
  type: 'task_completed' | 'task_started' | 'payment_received' | 'payment_sent' | 'agent_registered';
  title: string;
  description: string;
  time: string;
  amount?: string;
  isPositive?: boolean;
  timestamp: Date;
}

export function useActivity() {
  const { transactions: buyerTx, loading: buyerLoading } = useTransactions('buyer');
  const { transactions: sellerTx, loading: sellerLoading } = useTransactions('seller');
  const { agents, loading: agentsLoading } = useOwnedAgents();

  const activities = useMemo(() => {
    const items: ActivityItem[] = [];

    // Add seller transactions as income activities
    sellerTx.forEach((tx) => {
      if (tx.status === 'completed') {
        items.push({
          id: `tx-seller-${tx.id}`,
          type: 'payment_received',
          title: 'Payment received',
          description: `From ${tx.buyer_name || 'buyer'} for "${tx.title}"`,
          time: formatTimeAgo(tx.completed_at || tx.created_at),
          amount: `+$${tx.amount.toFixed(2)}`,
          isPositive: true,
          timestamp: new Date(tx.completed_at || tx.created_at),
        });
      } else if (tx.status === 'escrow_funded') {
        items.push({
          id: `tx-started-${tx.id}`,
          type: 'task_started',
          title: 'Task started',
          description: `"${tx.title}" - escrow funded`,
          time: formatTimeAgo(tx.funded_at || tx.created_at),
          timestamp: new Date(tx.funded_at || tx.created_at),
        });
      }
    });

    // Add buyer transactions as expense activities
    buyerTx.forEach((tx) => {
      if (tx.status === 'completed') {
        items.push({
          id: `tx-buyer-${tx.id}`,
          type: 'task_completed',
          title: 'Task completed',
          description: `"${tx.title}" by ${tx.seller_name || 'agent'}`,
          time: formatTimeAgo(tx.completed_at || tx.created_at),
          amount: `-$${tx.amount.toFixed(2)}`,
          isPositive: false,
          timestamp: new Date(tx.completed_at || tx.created_at),
        });
      } else if (tx.status === 'escrow_funded') {
        items.push({
          id: `tx-funded-${tx.id}`,
          type: 'payment_sent',
          title: 'Escrow funded',
          description: `For "${tx.title}"`,
          time: formatTimeAgo(tx.funded_at || tx.created_at),
          amount: `-$${tx.amount.toFixed(2)}`,
          isPositive: false,
          timestamp: new Date(tx.funded_at || tx.created_at),
        });
      }
    });

    // Add agent registration events
    agents.forEach((agent) => {
      items.push({
        id: `agent-${agent.id}`,
        type: 'agent_registered',
        title: 'Agent registered',
        description: `${agent.name} was linked to your account`,
        time: formatTimeAgo(agent.created_at),
        timestamp: new Date(agent.created_at),
      });
    });

    // Sort by timestamp descending
    return items.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
  }, [buyerTx, sellerTx, agents]);

  // Group by date
  const groupedActivities = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const todayItems: ActivityItem[] = [];
    const yesterdayItems: ActivityItem[] = [];
    const earlierItems: ActivityItem[] = [];

    activities.forEach((item) => {
      const itemDate = new Date(item.timestamp);
      itemDate.setHours(0, 0, 0, 0);

      if (itemDate.getTime() === today.getTime()) {
        todayItems.push(item);
      } else if (itemDate.getTime() === yesterday.getTime()) {
        yesterdayItems.push(item);
      } else {
        earlierItems.push(item);
      }
    });

    return { today: todayItems, yesterday: yesterdayItems, earlier: earlierItems };
  }, [activities]);

  return {
    activities,
    groupedActivities,
    loading: buyerLoading || sellerLoading || agentsLoading,
  };
}

// Tasks mapped from transactions
export interface Task {
  id: string;
  title: string;
  category: string;
  categoryColor: string;
  price: string;
  status: 'in_progress' | 'pending' | 'completed';
  agent?: string;
  time?: string;
  sourceType: 'transaction';
  sourceId: string;
}

export function useTasks() {
  const { transactions: buyerTx, loading: buyerLoading } = useTransactions('buyer');
  const { transactions: sellerTx, loading: sellerLoading } = useTransactions('seller');

  const tasks = useMemo(() => {
    const categoryColors: Record<string, string> = {
      'Data Analysis': '#22D3EE',
      Research: '#A855F7',
      Travel: '#F59E0B',
      Design: '#22C55E',
      Content: '#EC4899',
      Finance: '#6366F1',
      default: '#64748B',
    };

    const items: Task[] = [];

    // Buyer transactions = tasks the user ordered
    buyerTx.forEach((tx) => {
      let status: Task['status'] = 'pending';
      if (tx.status === 'completed') {
        status = 'completed';
      } else if (tx.status === 'escrow_funded' || tx.status === 'delivered') {
        status = 'in_progress';
      }

      items.push({
        id: tx.id,
        title: tx.title,
        category: 'Task',
        categoryColor: categoryColors.default,
        price: `$${tx.amount.toFixed(2)}`,
        status,
        agent: tx.seller_name,
        time: formatTimeAgo(tx.created_at),
        sourceType: 'transaction',
        sourceId: tx.id,
      });
    });

    // Seller transactions = tasks assigned to user's agents
    sellerTx.forEach((tx) => {
      let status: Task['status'] = 'pending';
      if (tx.status === 'completed') {
        status = 'completed';
      } else if (tx.status === 'escrow_funded' || tx.status === 'delivered') {
        status = 'in_progress';
      }

      items.push({
        id: `seller-${tx.id}`,
        title: tx.title,
        category: 'Task',
        categoryColor: categoryColors.default,
        price: `$${tx.amount.toFixed(2)}`,
        status,
        agent: tx.seller_name,
        time: formatTimeAgo(tx.created_at),
        sourceType: 'transaction',
        sourceId: tx.id,
      });
    });

    return items;
  }, [buyerTx, sellerTx]);

  const inProgressTasks = useMemo(() => tasks.filter((t) => t.status === 'in_progress'), [tasks]);
  const pendingTasks = useMemo(() => tasks.filter((t) => t.status === 'pending'), [tasks]);
  const completedTasks = useMemo(() => tasks.filter((t) => t.status === 'completed'), [tasks]);

  return {
    tasks,
    inProgressTasks,
    pendingTasks,
    completedTasks,
    loading: buyerLoading || sellerLoading,
  };
}

// Agent's tasks and ratings
export function useAgentTransactions(agentId: string | null) {
  const isAuthReady = useAuthReady();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!agentId || !isAuthReady) {
      setTransactions([]);
      return;
    }

    setLoading(true);
    setError(null);

    // Get transactions where this agent is the seller
    api
      .getTransactions({ role: 'seller', limit: 50 })
      .then((data) => {
        // Filter to only this agent's transactions
        const agentTx = (data.transactions || []).filter(
          (tx) => tx.seller_id === agentId
        );
        setTransactions(agentTx);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [agentId, isAuthReady]);

  return { transactions, loading, error };
}

export function useAgentRatings(agentId: string | null) {
  const { transactions, loading: txLoading } = useAgentTransactions(agentId);
  const [ratings, setRatings] = useState<Rating[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (transactions.length === 0) {
      setRatings([]);
      return;
    }

    setLoading(true);

    // Fetch ratings for completed transactions
    const completedTx = transactions.filter((tx) => tx.status === 'completed');

    Promise.all(
      completedTx.map((tx) =>
        api.getTransactionRatings(tx.id).catch(() => ({ ratings: [] }))
      )
    )
      .then((results) => {
        const allRatings = results.flatMap((r) => r.ratings);
        setRatings(allRatings);
      })
      .finally(() => setLoading(false));
  }, [transactions]);

  return { ratings, loading: txLoading || loading };
}

// Fetch activity events for a specific agent
export function useAgentActivity(agentId: string | null, limit = 50) {
  const isAuthReady = useAuthReady();
  const [events, setEvents] = useState<ActivityEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(() => {
    if (!agentId || !isAuthReady) {
      setEvents([]);
      setTotal(0);
      return;
    }

    setLoading(true);
    setError(null);

    api
      .getAgentActivity(agentId, { limit })
      .then((data) => {
        setEvents(data.events || []);
        setTotal(data.total || 0);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [agentId, isAuthReady, limit]);

  useEffect(() => {
    refetch();
  }, [refetch]);

  // Group events by date for display
  const groupedEvents = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const todayItems: ActivityEvent[] = [];
    const yesterdayItems: ActivityEvent[] = [];
    const earlierItems: ActivityEvent[] = [];

    events.forEach((event) => {
      const eventDate = new Date(event.created_at);
      eventDate.setHours(0, 0, 0, 0);

      if (eventDate.getTime() === today.getTime()) {
        todayItems.push(event);
      } else if (eventDate.getTime() === yesterday.getTime()) {
        yesterdayItems.push(event);
      } else {
        earlierItems.push(event);
      }
    });

    return { today: todayItems, yesterday: yesterdayItems, earlier: earlierItems };
  }, [events]);

  return { events, groupedEvents, total, loading, error, refetch };
}

// Utility function
function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}

