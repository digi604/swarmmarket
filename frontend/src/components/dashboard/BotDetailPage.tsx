import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Settings,
  Loader2,
  Bot,
  LayoutDashboard,
  ClipboardList,
  Gavel,
  ShoppingCart,
  Activity,
} from 'lucide-react';
import { useOwnedAgents, useAgentMetrics } from '../../hooks/useDashboard';
import {
  OverviewTab,
  TasksTab,
  AuctionsTab,
  OrdersTab,
  ActivityTab,
} from './agent-tabs';

type TabId = 'overview' | 'tasks' | 'auctions' | 'orders' | 'activity';

interface Tab {
  id: TabId;
  label: string;
  icon: React.ElementType;
}

const tabs: Tab[] = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard },
  { id: 'tasks', label: 'Tasks', icon: ClipboardList },
  { id: 'auctions', label: 'Auctions', icon: Gavel },
  { id: 'orders', label: 'Orders', icon: ShoppingCart },
  { id: 'activity', label: 'Activity', icon: Activity },
];

export function BotDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<TabId>('overview');
  const { agents, loading: agentsLoading } = useOwnedAgents();
  const { metrics, loading: metricsLoading } = useAgentMetrics(id || null);

  const agent = agents.find((a) => a.id === id);

  if (agentsLoading || metricsLoading) {
    return (
      <div className="flex items-center justify-center" style={{ padding: '80px 0' }}>
        <Loader2 className="w-8 h-8 text-[#22D3EE] animate-spin" />
      </div>
    );
  }

  if (!agent) {
    return (
      <div className="flex flex-col items-center justify-center" style={{ padding: '80px 0' }}>
        <p className="text-[18px] text-white" style={{ marginBottom: '16px' }}>
          Agent not found
        </p>
        <button
          onClick={() => navigate('/dashboard/agents')}
          className="text-[14px] text-[#22D3EE] hover:underline"
        >
          Back to My Agents
        </button>
      </div>
    );
  }

  const colors = ['#22D3EE', '#A855F7', '#F59E0B', '#22C55E', '#EC4899'];
  const colorIndex = agent.name.charCodeAt(0) % colors.length;
  const color = colors[colorIndex];
  const statusColor = agent.is_active ? '#22C55E' : '#64748B';

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return (
          <OverviewTab
            agentId={agent.id}
            metrics={metrics}
            trustScore={agent.trust_score}
          />
        );
      case 'tasks':
        return <TasksTab agentId={agent.id} />;
      case 'auctions':
        return <AuctionsTab agentId={agent.id} />;
      case 'orders':
        return <OrdersTab agentId={agent.id} />;
      case 'activity':
        return <ActivityTab agentId={agent.id} />;
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full" style={{ gap: '32px' }}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center" style={{ gap: '20px' }}>
          <button
            onClick={() => navigate('/dashboard/agents')}
            className="flex items-center rounded-lg bg-[#1E293B] text-[#94A3B8] text-[14px] hover:bg-[#2D3B4F] transition-colors"
            style={{ padding: '10px 14px', gap: '8px' }}
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </button>
          <div className="flex items-center" style={{ gap: '16px' }}>
            <div
              className="rounded-full flex items-center justify-center"
              style={{ width: '56px', height: '56px', backgroundColor: color }}
            >
              <Bot className="w-7 h-7 text-[#0A0F1C]" />
            </div>
            <div className="flex flex-col" style={{ gap: '4px' }}>
              <h1 className="text-[24px] font-bold text-white leading-tight">{agent.name}</h1>
              <p className="text-[14px] text-[#64748B]">
                {agent.description || 'AI agent'}
              </p>
            </div>
          </div>
        </div>
        <div className="flex items-center" style={{ gap: '12px' }}>
          <div
            className="flex items-center rounded-full"
            style={{
              padding: '8px 16px',
              gap: '6px',
              backgroundColor: agent.is_active
                ? 'rgba(34, 197, 94, 0.125)'
                : 'rgba(100, 116, 139, 0.125)',
            }}
          >
            <div
              className="rounded-full"
              style={{ width: '8px', height: '8px', backgroundColor: statusColor }}
            />
            <span className="text-[14px] font-semibold" style={{ color: statusColor }}>
              {agent.is_active ? 'Active' : 'Inactive'}
            </span>
          </div>
          <button
            className="flex items-center rounded-lg bg-[#1E293B] text-[#94A3B8] text-[14px] hover:bg-[#2D3B4F] transition-colors"
            style={{ padding: '10px 20px', gap: '8px', border: '1px solid #334155' }}
          >
            <Settings className="w-4 h-4" />
            Settings
          </button>
        </div>
      </div>

      {/* Tabs Row */}
      <div
        className="flex"
        style={{ gap: '8px', borderBottom: '1px solid #334155' }}
      >
        {tabs.map((tab) => {
          const isActive = activeTab === tab.id;
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className="flex items-center transition-colors"
              style={{
                padding: '12px 16px',
                gap: '8px',
                borderBottom: isActive ? '2px solid #22D3EE' : '2px solid transparent',
                marginBottom: '-1px',
              }}
            >
              <Icon
                className="w-4 h-4"
                style={{ color: isActive ? '#22D3EE' : '#64748B' }}
              />
              <span
                className="text-[14px] font-medium"
                style={{ color: isActive ? '#22D3EE' : '#64748B' }}
              >
                {tab.label}
              </span>
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      <div className="flex flex-col flex-1" style={{ gap: '32px' }}>
        {renderTabContent()}
      </div>
    </div>
  );
}
