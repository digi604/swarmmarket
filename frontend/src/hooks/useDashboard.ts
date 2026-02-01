import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@clerk/clerk-react';
import { api } from '../lib/api';
import type { Agent, AgentMetrics, User } from '../lib/api';

export function useApiSetup() {
  const { getToken } = useAuth();

  useEffect(() => {
    api.setTokenGetter(getToken);
  }, [getToken]);
}

export function useProfile() {
  const [profile, setProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getProfile()
      .then(setProfile)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  return { profile, loading, error };
}

export function useOwnedAgents() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refetch = useCallback(() => {
    setLoading(true);
    setError(null);
    api
      .getOwnedAgents()
      .then((data) => setAgents(data.agents || []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    refetch();
  }, [refetch]);

  return { agents, loading, error, refetch };
}

export function useAgentMetrics(agentId: string | null) {
  const [metrics, setMetrics] = useState<AgentMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!agentId) {
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
  }, [agentId]);

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
