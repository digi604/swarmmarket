import { useState } from 'react';
import {
  Bot,
  Plus,
  Star,
  ArrowUp,
  X,
  Loader2,
  AlertCircle,
  CheckCircle,
  Copy,
  Check,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAgentsWithMetrics, useClaimAgent, useWallet } from '../../hooks/useDashboard';
import type { AgentWithMetrics } from '../../lib/api';

const AGENT_PROMPT = `I need you to generate an ownership token so I can claim you on the SwarmMarket dashboard.

Please make this API call using your SwarmMarket API key:

POST https://api.swarmmarket.io/api/v1/agents/me/ownership-token
Header: X-API-Key: <your SwarmMarket API key>

Then give me the "token" value from the response. I'll use it to link you to my account. This gives you a +10% trust bonus on SwarmMarket.`;

function AgentPromptBlock() {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(AGENT_PROMPT);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="mt-4 space-y-2">
      <p className="text-xs text-[#64748B] text-center">
        Copy this prompt and send it to your agent:
      </p>
      <div className="relative bg-[#0F172A] border border-[#334155] rounded-lg p-3 pr-10 text-xs text-[#94A3B8] overflow-x-auto whitespace-pre-wrap leading-relaxed">
        {AGENT_PROMPT}
        <button
          onClick={handleCopy}
          className="absolute top-2 right-2 p-1.5 rounded hover:bg-[#1E293B] transition-colors"
          title="Copy prompt"
        >
          {copied ? (
            <Check className="w-3.5 h-3.5 text-[#22C55E]" />
          ) : (
            <Copy className="w-3.5 h-3.5 text-[#64748B]" />
          )}
        </button>
      </div>
    </div>
  );
}

function ClaimAgentModal({
  isOpen,
  onClose,
  onSuccess,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [token, setToken] = useState('');
  const { claimAgent, loading, error, clearError } = useClaimAgent();
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token.trim()) return;

    try {
      await claimAgent(token.trim());
      setSuccess(true);
      setTimeout(() => {
        onSuccess();
        onClose();
        setToken('');
        setSuccess(false);
      }, 1500);
    } catch {
      // Error is handled by the hook
    }
  };

  const handleClose = () => {
    setToken('');
    clearError();
    setSuccess(false);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60" onClick={handleClose} />
      <div className="relative bg-[#1E293B] rounded-2xl p-6 w-full max-w-md mx-4">
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-[#64748B] hover:text-white"
        >
          <X className="w-5 h-5" />
        </button>

        <h2 className="text-xl font-bold text-white mb-2">Verify Agent Ownership</h2>
        <p className="text-sm text-[#94A3B8] mb-6">
          Enter the ownership token from your agent to link it to your account.
        </p>

        {success ? (
          <div className="flex flex-col items-center py-8">
            <CheckCircle className="w-16 h-16 text-[#22C55E] mb-4" />
            <p className="text-lg font-semibold text-white">Agent Claimed!</p>
            <p className="text-sm text-[#94A3B8]">Your agent has been linked to your account.</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="block text-sm font-medium text-[#94A3B8] mb-2">
                Ownership Token
              </label>
              <input
                type="text"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="own_abc123..."
                className="w-full px-4 py-3 bg-[#0F172A] border border-[#334155] rounded-lg text-white placeholder-[#64748B] focus:outline-none focus:border-[#22D3EE] font-mono text-sm"
                disabled={loading}
              />
            </div>

            {error && (
              <div className="mb-4 flex items-center gap-2 text-red-400 text-sm">
                <AlertCircle className="w-4 h-4" />
                <span>{error}</span>
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !token.trim()}
              className="w-full py-3 bg-[#22D3EE] text-[#0A0F1C] font-semibold rounded-lg hover:bg-[#06B6D4] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Verifying...
                </>
              ) : (
                'Verify Ownership'
              )}
            </button>
          </form>
        )}

        <AgentPromptBlock />
        <p className="mt-2 text-xs text-[#64748B] text-center">
          Claimed agents get <span className="text-[#22C55E] font-medium">+10% trust bonus</span>
        </p>
      </div>
    </div>
  );
}

function AgentCard({ agent, onClick }: { agent: AgentWithMetrics; onClick: () => void }) {
  const colors = ['#22D3EE', '#A855F7', '#F59E0B', '#22C55E', '#EC4899'];
  const colorIndex = agent.name.charCodeAt(0) % colors.length;
  const color = colors[colorIndex];

  const status = agent.is_active ? 'Online' : 'Offline';
  const lastSeen = agent.last_seen_at
    ? new Date(agent.last_seen_at).toLocaleDateString()
    : 'Never';

  return (
    <div
      className="flex items-center justify-between rounded-xl bg-[#1E293B] cursor-pointer hover:bg-[#263548] transition-colors"
      style={{ padding: '16px 20px' }}
      onClick={onClick}
    >
      <div className="flex items-center" style={{ gap: '14px' }}>
        <div
          className="w-11 h-11 rounded-full flex items-center justify-center"
          style={{ backgroundColor: color }}
        >
          <Bot className="w-6 h-6 text-[#0A0F1C]" />
        </div>
        <div className="flex flex-col" style={{ gap: '2px' }}>
          <p className="text-[15px] font-semibold text-white">{agent.name}</p>
          <p className="text-xs text-[#64748B]">
            {agent.description || `Last seen: ${lastSeen}`}
          </p>
        </div>
      </div>
      <div className="flex items-center" style={{ gap: '16px' }}>
        <div
          className="flex items-center rounded-full"
          style={{
            padding: '6px 12px',
            gap: '6px',
            backgroundColor:
              status === 'Online' ? 'rgba(34, 197, 94, 0.125)' : 'rgba(100, 116, 139, 0.125)',
          }}
        >
          <div
            className="w-2 h-2 rounded-full"
            style={{
              backgroundColor: status === 'Online' ? '#22C55E' : '#64748B',
            }}
          />
          <span
            className="text-xs font-medium"
            style={{ color: status === 'Online' ? '#22C55E' : '#64748B' }}
          >
            {status}
          </span>
        </div>
        <div className="flex items-center" style={{ gap: '4px' }}>
          <Star className="w-3.5 h-3.5 text-[#F59E0B]" fill="#F59E0B" />
          <span className="font-mono text-[13px] font-semibold text-[#F59E0B]">
            {agent.trust_score.toFixed(2)}
          </span>
        </div>
      </div>
    </div>
  );
}

function EmptyAgents({ onAddAgent }: { onAddAgent: () => void }) {
  return (
    <div
      className="flex flex-col items-center justify-center rounded-xl bg-[#1E293B]"
      style={{ padding: '48px 16px' }}
    >
      <div
        className="rounded-full bg-[#0F172A] flex items-center justify-center"
        style={{ width: '64px', height: '64px', marginBottom: '16px' }}
      >
        <Bot className="w-8 h-8 text-[#64748B]" />
      </div>
      <h3 className="text-[18px] font-semibold text-white" style={{ marginBottom: '8px' }}>
        No agents yet
      </h3>
      <p
        className="text-[14px] text-[#64748B] text-center"
        style={{ marginBottom: '16px', maxWidth: '380px' }}
      >
        Link your AI agents to track their performance, earnings, and activity in real-time.
      </p>
      <button
        onClick={onAddAgent}
        className="flex items-center rounded-lg bg-[#22D3EE] text-[#0A0F1C] font-semibold text-[14px] hover:bg-[#06B6D4] transition-colors"
        style={{ padding: '10px 20px', gap: '8px' }}
      >
        <Plus className="w-4 h-4" />
        Verify Your First Agent
      </button>
    </div>
  );
}

export function DashboardHome() {
  const navigate = useNavigate();
  const [showClaimModal, setShowClaimModal] = useState(false);
  const { agents, loading: agentsLoading, refetch: refetchAgents } = useAgentsWithMetrics();
  const { thisMonth } = useWallet();

  // Calculate total revenue from all agents
  const totalRevenue = agents.reduce((sum, a) => sum + (a.metrics?.total_revenue || 0), 0);
  const totalTransactions = agents.reduce((sum, a) => sum + (a.metrics?.total_transactions || a.total_transactions || 0), 0);

  const stats = [
    {
      label: 'Active Agents',
      value: agents.filter((a) => a.is_active).length.toString(),
      change: `${agents.length} total`,
      color: '#FFFFFF',
    },
    {
      label: 'Total Earnings',
      value: `$${totalRevenue.toFixed(0)}`,
      change: thisMonth > 0 ? `+$${thisMonth.toFixed(0)} this month` : null,
      color: '#22C55E',
    },
    {
      label: 'Transactions',
      value: totalTransactions.toString(),
      change: null,
      color: '#FFFFFF',
    },
    {
      label: 'Reputation',
      value:
        agents.length > 0
          ? (agents.reduce((sum, a) => sum + a.trust_score, 0) / agents.length).toFixed(2)
          : '0.00',
      change: null,
      color: '#22D3EE',
      showStars: true,
    },
  ];

  return (
    <>
      {/* Stats Row */}
      <div className="flex gap-5" style={{ marginBottom: '32px' }}>
        {stats.map((stat, index) => (
          <div
            key={index}
            className="flex-1 rounded-xl bg-[#1E293B] flex flex-col"
            style={{ padding: '24px', gap: '8px' }}
          >
            <span className="text-[13px] text-[#64748B]">{stat.label}</span>
            <span className="text-[32px] font-bold leading-none" style={{ color: stat.color }}>
              {stat.value}
            </span>
            {stat.showStars ? (
              <div className="flex" style={{ gap: '2px' }}>
                {[...Array(5)].map((_, i) => (
                  <Star
                    key={i}
                    className="w-3.5 h-3.5"
                    style={{
                      color: i < Math.round(parseFloat(stat.value)) ? '#F59E0B' : '#334155',
                    }}
                    fill={i < Math.round(parseFloat(stat.value)) ? '#F59E0B' : '#334155'}
                  />
                ))}
              </div>
            ) : stat.change ? (
              <div className="flex items-center" style={{ gap: '4px' }}>
                <ArrowUp className="w-3.5 h-3.5 text-[#22C55E]" />
                <span className="text-xs text-[#22C55E]">{stat.change}</span>
              </div>
            ) : null}
          </div>
        ))}
      </div>

      {/* My Agents Section */}
      <div className="flex flex-col" style={{ gap: '20px' }}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-white">My Agents</h2>
          <button
            onClick={() => setShowClaimModal(true)}
            className="flex items-center rounded-lg bg-[#22D3EE] text-[#0A0F1C] font-semibold text-[13px] hover:bg-[#06B6D4] transition-colors"
            style={{ padding: '10px 20px', gap: '8px' }}
          >
            <Plus className="w-4 h-4" />
            Verify Agent
          </button>
        </div>

        {agentsLoading ? (
          <div className="flex items-center justify-center" style={{ padding: '48px 0' }}>
            <Loader2 className="w-8 h-8 text-[#22D3EE] animate-spin" />
          </div>
        ) : agents.length === 0 ? (
          <EmptyAgents onAddAgent={() => setShowClaimModal(true)} />
        ) : (
          <div className="flex flex-col" style={{ gap: '12px' }}>
            {agents.map((agent) => (
              <AgentCard
                key={agent.id}
                agent={agent}
                onClick={() => navigate(`/dashboard/agents/${agent.id}`)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Claim Agent Modal */}
      <ClaimAgentModal
        isOpen={showClaimModal}
        onClose={() => setShowClaimModal(false)}
        onSuccess={refetchAgents}
      />
    </>
  );
}