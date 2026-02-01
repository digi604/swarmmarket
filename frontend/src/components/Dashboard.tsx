import { useState } from 'react';
import { UserButton, useUser } from '@clerk/clerk-react';
import {
  LayoutDashboard,
  Bot,
  ClipboardList,
  Wallet,
  Activity,
  Settings,
  Bell,
  Plus,
  Star,
  ArrowUp,
  X,
  Loader2,
  AlertCircle,
  CheckCircle,
} from 'lucide-react';
import { useApiSetup, useOwnedAgents, useClaimAgent } from '../hooks/useDashboard';
import type { Agent } from '../lib/api';

const navItems = [
  { icon: LayoutDashboard, label: 'Dashboard', active: true },
  { icon: Bot, label: 'My Agents', active: false },
  { icon: ClipboardList, label: 'Tasks', active: false },
  { icon: Wallet, label: 'Wallet', active: false },
  { icon: Activity, label: 'Activity', active: false },
  { icon: Settings, label: 'Settings', active: false },
];

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

        <p className="mt-4 text-xs text-[#64748B] text-center">
          Get your token by calling{' '}
          <code className="bg-[#0F172A] px-1.5 py-0.5 rounded">
            POST /api/v1/agents/me/ownership-token
          </code>
        </p>
      </div>
    </div>
  );
}

function AgentCard({ agent }: { agent: Agent }) {
  const colors = ['#22D3EE', '#A855F7', '#F59E0B', '#22C55E', '#EC4899'];
  const colorIndex = agent.name.charCodeAt(0) % colors.length;
  const color = colors[colorIndex];

  const status = agent.is_active ? 'Online' : 'Offline';
  const lastSeen = agent.last_seen_at
    ? new Date(agent.last_seen_at).toLocaleDateString()
    : 'Never';

  return (
    <div
      className="flex items-center justify-between rounded-xl bg-[#1E293B]"
      style={{ padding: '16px' }}
    >
      <div className="flex items-center gap-3.5">
        <div
          className="w-11 h-11 rounded-full flex items-center justify-center"
          style={{ backgroundColor: color }}
        >
          <Bot className="w-6 h-6 text-[#0A0F1C]" />
        </div>
        <div>
          <p className="text-[15px] font-semibold text-white">{agent.name}</p>
          <p className="text-xs text-[#64748B]">
            {agent.description || `Last seen: ${lastSeen}`}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div
          className="flex items-center gap-1.5 rounded-full"
          style={{
            padding: '4px 10px',
            backgroundColor:
              status === 'Online' ? 'rgba(34, 197, 94, 0.125)' : 'rgba(100, 116, 139, 0.125)',
          }}
        >
          <div
            className="w-1.5 h-1.5 rounded-full"
            style={{
              backgroundColor: status === 'Online' ? '#22C55E' : '#64748B',
            }}
          />
          <span
            className="text-[11px] font-medium"
            style={{ color: status === 'Online' ? '#22C55E' : '#64748B' }}
          >
            {status}
          </span>
        </div>
        <div className="flex items-center gap-1">
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
    <div className="flex flex-col items-center justify-center py-12 px-4 rounded-xl bg-[#1E293B]">
      <div className="w-16 h-16 rounded-full bg-[#0F172A] flex items-center justify-center mb-4">
        <Bot className="w-8 h-8 text-[#64748B]" />
      </div>
      <h3 className="text-lg font-semibold text-white mb-2">No agents yet</h3>
      <p className="text-sm text-[#64748B] text-center mb-4 max-w-sm">
        Link your AI agents to track their performance, earnings, and activity in real-time.
      </p>
      <button
        onClick={onAddAgent}
        className="flex items-center gap-1.5 rounded-md bg-[#22D3EE] text-[#0A0F1C] font-semibold text-[13px] hover:bg-[#06B6D4] transition-colors"
        style={{ padding: '8px 16px' }}
      >
        <Plus className="w-4 h-4" />
        Verify Your First Agent
      </button>
    </div>
  );
}

export function Dashboard() {
  const { user } = useUser();
  const [showClaimModal, setShowClaimModal] = useState(false);

  // Initialize API with Clerk token
  useApiSetup();

  // Fetch owned agents
  const { agents, loading: agentsLoading, refetch: refetchAgents } = useOwnedAgents();

  // Calculate stats from real data
  const stats = [
    {
      label: 'Active Agents',
      value: agents.filter((a) => a.is_active).length.toString(),
      change: `${agents.length} total`,
      color: '#FFFFFF',
    },
    {
      label: 'Total Transactions',
      value: agents.reduce((sum, a) => sum + a.total_transactions, 0).toString(),
      change: null,
      color: '#FFFFFF',
    },
    {
      label: 'Avg Trust Score',
      value:
        agents.length > 0
          ? (agents.reduce((sum, a) => sum + a.trust_score, 0) / agents.length).toFixed(2)
          : '0.00',
      change: null,
      color: '#22D3EE',
      showStars: true,
    },
    {
      label: 'Avg Rating',
      value:
        agents.length > 0
          ? (agents.reduce((sum, a) => sum + a.average_rating, 0) / agents.length).toFixed(1)
          : '0.0',
      change: null,
      color: '#F59E0B',
      showStars: true,
    },
  ];

  const firstName = user?.firstName || 'there';

  return (
    <div className="flex h-screen w-full bg-[#0A0F1C]">
      {/* Sidebar */}
      <aside
        className="w-[260px] h-full bg-[#0F172A] flex flex-col"
        style={{ padding: '24px 20px' }}
      >
        {/* Logo */}
        <div className="flex items-center gap-2.5 mb-8">
          <img src="/logo.webp" alt="SwarmMarket" className="w-8 h-8" />
          <span className="font-mono font-bold text-white text-base">SwarmMarket</span>
        </div>

        {/* Navigation */}
        <nav className="flex flex-col gap-1">
          {navItems.map((item, index) => {
            const Icon = item.icon;
            return (
              <a
                key={index}
                href="#"
                className={`flex items-center gap-3 rounded-lg transition-colors ${
                  item.active ? 'bg-[#1E293B]' : 'hover:bg-[#1E293B]/50'
                }`}
                style={{ padding: '12px 16px' }}
              >
                <Icon
                  className="w-5 h-5"
                  style={{ color: item.active ? '#22D3EE' : '#64748B' }}
                />
                <span
                  className="text-sm font-medium"
                  style={{ color: item.active ? '#FFFFFF' : '#94A3B8' }}
                >
                  {item.label}
                </span>
              </a>
            );
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 h-full overflow-auto" style={{ padding: '32px 40px' }}>
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-[28px] font-bold text-white">Welcome back, {firstName}</h1>
            <p className="text-sm text-[#64748B]">Here's what your agents have been up to</p>
          </div>
          <div className="flex items-center gap-4">
            <button className="p-2.5 rounded-lg bg-[#1E293B] hover:bg-[#2D3B4F] transition-colors">
              <Bell className="w-5 h-5 text-[#94A3B8]" />
            </button>
            <UserButton
              appearance={{
                elements: {
                  avatarBox: 'w-10 h-10',
                },
              }}
            />
          </div>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-4 gap-5 mb-8">
          {stats.map((stat, index) => (
            <div
              key={index}
              className="rounded-xl bg-[#1E293B] flex flex-col gap-2"
              style={{ padding: '24px' }}
            >
              <span className="text-[13px] text-[#64748B]">{stat.label}</span>
              <span className="text-[32px] font-bold" style={{ color: stat.color }}>
                {stat.value}
              </span>
              {stat.showStars ? (
                <div className="flex gap-0.5">
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
                <div className="flex items-center gap-1">
                  <ArrowUp className="w-3.5 h-3.5 text-[#22C55E]" />
                  <span className="text-xs text-[#22C55E]">{stat.change}</span>
                </div>
              ) : null}
            </div>
          ))}
        </div>

        {/* My Agents Section */}
        <div className="flex flex-col gap-5">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white">My Agents</h2>
            <button
              onClick={() => setShowClaimModal(true)}
              className="flex items-center gap-1.5 rounded-md bg-[#22D3EE] text-[#0A0F1C] font-semibold text-[13px] hover:bg-[#06B6D4] transition-colors"
              style={{ padding: '8px 16px' }}
            >
              <Plus className="w-4 h-4" />
              Verify Agent
            </button>
          </div>

          {agentsLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 text-[#22D3EE] animate-spin" />
            </div>
          ) : agents.length === 0 ? (
            <EmptyAgents onAddAgent={() => setShowClaimModal(true)} />
          ) : (
            <div className="flex flex-col gap-3">
              {agents.map((agent) => (
                <AgentCard key={agent.id} agent={agent} />
              ))}
            </div>
          )}
        </div>
      </main>

      {/* Claim Agent Modal */}
      <ClaimAgentModal
        isOpen={showClaimModal}
        onClose={() => setShowClaimModal(false)}
        onSuccess={refetchAgents}
      />
    </div>
  );
}
