import { useState, useEffect, useCallback } from 'react';
import { User, Bell, Shield, CreditCard, Key, ExternalLink, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { api } from '../../lib/api';
import type { ConnectStatus } from '../../lib/api';

type SettingsTab = 'profile' | 'notifications' | 'security' | 'billing' | 'api_keys';

interface ToggleSetting {
  id: string;
  title: string;
  description: string;
  enabled: boolean;
}

const notificationSettings: ToggleSetting[] = [
  {
    id: 'email',
    title: 'Email Notifications',
    description: 'Receive email updates for your agents and tasks',
    enabled: true,
  },
  {
    id: 'task_completed',
    title: 'Task Completed',
    description: 'Get notified when an agent completes a task',
    enabled: true,
  },
  {
    id: 'payment_received',
    title: 'Payment Received',
    description: 'Get notified when you receive a payment',
    enabled: true,
  },
  {
    id: 'agent_status',
    title: 'Agent Status Changes',
    description: 'Get notified when an agent goes online or offline',
    enabled: false,
  },
  {
    id: 'marketing',
    title: 'Marketing & Updates',
    description: 'Receive news about new features and updates',
    enabled: false,
  },
];

function Toggle({
  enabled,
  onChange,
}: {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
}) {
  return (
    <button
      onClick={() => onChange(!enabled)}
      className="relative rounded-full transition-colors"
      style={{
        width: '44px',
        height: '24px',
        backgroundColor: enabled ? '#A855F7' : '#334155',
      }}
    >
      <div
        className="absolute rounded-full bg-white transition-transform"
        style={{
          width: '16px',
          height: '16px',
          top: '4px',
          left: enabled ? '24px' : '4px',
        }}
      />
    </button>
  );
}

function SettingRow({
  setting,
  onToggle,
}: {
  setting: ToggleSetting;
  onToggle: (id: string, enabled: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between" style={{ padding: '16px 0' }}>
      <div className="flex flex-col" style={{ gap: '2px' }}>
        <span className="text-[14px] font-medium text-white">{setting.title}</span>
        <span className="text-[12px] text-[#64748B]">{setting.description}</span>
      </div>
      <Toggle enabled={setting.enabled} onChange={(enabled) => onToggle(setting.id, enabled)} />
    </div>
  );
}

function StatusBadge({ ok, label }: { ok: boolean; label: string }) {
  return (
    <div className="flex items-center" style={{ gap: '8px' }}>
      {ok ? (
        <CheckCircle className="w-4 h-4 text-green-400" />
      ) : (
        <AlertCircle className="w-4 h-4 text-yellow-400" />
      )}
      <span className={`text-[13px] ${ok ? 'text-green-400' : 'text-yellow-400'}`}>{label}</span>
    </div>
  );
}

function ConnectSection() {
  const [status, setStatus] = useState<ConnectStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const s = await api.getConnectStatus();
      setStatus(s);
    } catch {
      setError('Failed to load payout status');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  const handleOnboard = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await api.startConnectOnboarding();
      window.location.href = result.onboarding_url;
    } catch {
      setError('Failed to start onboarding');
      setActionLoading(false);
    }
  };

  const handleLoginLink = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await api.getConnectLoginLink();
      window.open(result.url, '_blank');
    } catch {
      setError('Failed to open Stripe dashboard');
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center" style={{ padding: '48px' }}>
        <Loader2 className="w-6 h-6 text-[#64748B] animate-spin" />
      </div>
    );
  }

  const hasAccount = status?.account_id != null;
  const isComplete = status?.charges_enabled && status?.payouts_enabled;
  const inProgress = hasAccount && !isComplete;

  return (
    <div className="flex flex-col" style={{ gap: '24px' }}>
      <div className="flex flex-col" style={{ gap: '4px' }}>
        <h3 className="text-[16px] font-medium text-white">Receive Payouts</h3>
        <p className="text-[13px] text-[#64748B]">
          Set up your payout account so your agents can receive payments from marketplace transactions.
        </p>
      </div>

      {error && (
        <div
          className="rounded-lg text-[13px] text-red-400"
          style={{ padding: '12px 16px', backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
        >
          {error}
        </div>
      )}

      {/* Status indicators */}
      {hasAccount && (
        <div
          className="rounded-lg flex flex-col"
          style={{ padding: '16px', gap: '12px', backgroundColor: '#0F172A' }}
        >
          <StatusBadge ok={status!.details_submitted} label={status!.details_submitted ? 'Details submitted' : 'Details incomplete'} />
          <StatusBadge ok={status!.charges_enabled} label={status!.charges_enabled ? 'Charges enabled' : 'Charges pending'} />
          <StatusBadge ok={status!.payouts_enabled} label={status!.payouts_enabled ? 'Payouts enabled' : 'Payouts pending'} />
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center" style={{ gap: '12px' }}>
        {!hasAccount && (
          <button
            onClick={handleOnboard}
            disabled={actionLoading}
            className="flex items-center rounded-lg text-[14px] font-medium text-white bg-[#A855F7] hover:bg-[#9333EA] transition-colors disabled:opacity-50"
            style={{ padding: '10px 20px', gap: '8px' }}
          >
            {actionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <CreditCard className="w-4 h-4" />}
            Set up payouts
          </button>
        )}

        {inProgress && (
          <button
            onClick={handleOnboard}
            disabled={actionLoading}
            className="flex items-center rounded-lg text-[14px] font-medium text-white bg-[#A855F7] hover:bg-[#9333EA] transition-colors disabled:opacity-50"
            style={{ padding: '10px 20px', gap: '8px' }}
          >
            {actionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <ExternalLink className="w-4 h-4" />}
            Resume onboarding
          </button>
        )}

        {isComplete && (
          <button
            onClick={handleLoginLink}
            disabled={actionLoading}
            className="flex items-center rounded-lg text-[14px] font-medium text-white bg-[#334155] hover:bg-[#475569] transition-colors disabled:opacity-50"
            style={{ padding: '10px 20px', gap: '8px' }}
          >
            {actionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <ExternalLink className="w-4 h-4" />}
            Stripe Dashboard
          </button>
        )}
      </div>
    </div>
  );
}

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('notifications');
  const [settings, setSettings] = useState(notificationSettings);

  // Check URL params for tab override (e.g. returning from Stripe onboarding)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const tab = params.get('tab');
    if (tab === 'billing') {
      setActiveTab('billing');
    }
  }, []);

  const handleToggle = (id: string, enabled: boolean) => {
    setSettings((prev) => prev.map((s) => (s.id === id ? { ...s, enabled } : s)));
  };

  const tabs = [
    { id: 'profile' as SettingsTab, label: 'Profile', icon: User },
    { id: 'notifications' as SettingsTab, label: 'Notifications', icon: Bell },
    { id: 'security' as SettingsTab, label: 'Security', icon: Shield },
    { id: 'billing' as SettingsTab, label: 'Billing', icon: CreditCard },
    { id: 'api_keys' as SettingsTab, label: 'API Keys', icon: Key },
  ];

  return (
    <div className="flex flex-col" style={{ gap: '32px' }}>
      {/* Header */}
      <div className="flex flex-col" style={{ gap: '4px' }}>
        <h1 className="text-[28px] font-bold text-white leading-tight">Settings</h1>
        <p className="text-[14px] text-[#64748B]">Manage your account and preferences</p>
      </div>

      {/* Content */}
      <div className="flex" style={{ gap: '32px' }}>
        {/* Sidebar */}
        <div className="flex flex-col" style={{ width: '240px', gap: '4px' }}>
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center rounded-lg text-left transition-colors ${
                  isActive ? 'bg-[#1E293B]' : 'hover:bg-[#1E293B]/50'
                }`}
                style={{ padding: '12px 16px', gap: '12px' }}
              >
                <Icon
                  className="w-5 h-5"
                  style={{ color: isActive ? '#A855F7' : '#64748B' }}
                />
                <span
                  className="text-[14px] font-medium"
                  style={{ color: isActive ? '#FFFFFF' : '#94A3B8' }}
                >
                  {tab.label}
                </span>
              </button>
            );
          })}
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col" style={{ gap: '32px' }}>
          {/* Notification Settings */}
          {activeTab === 'notifications' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Notification Settings
              </h2>
              <div className="flex flex-col">
                {settings.map((setting, index) => (
                  <div
                    key={setting.id}
                    style={{
                      borderBottom:
                        index < settings.length - 1 ? '1px solid #334155' : 'none',
                    }}
                  >
                    <SettingRow setting={setting} onToggle={handleToggle} />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Profile Settings
              </h2>
              <p className="text-[14px] text-[#64748B]">Profile settings will be available soon.</p>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Security Settings
              </h2>
              <p className="text-[14px] text-[#64748B]">Security settings will be available soon.</p>
            </div>
          )}

          {/* Billing Tab */}
          {activeTab === 'billing' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Billing & Payouts
              </h2>
              <ConnectSection />
            </div>
          )}

          {/* API Keys Tab */}
          {activeTab === 'api_keys' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                API Keys
              </h2>
              <p className="text-[14px] text-[#64748B]">API key management will be available soon.</p>
            </div>
          )}

          {/* Danger Zone */}
          <div
            className="rounded-xl bg-[#1E293B]"
            style={{ padding: '32px', border: '1px solid rgba(239, 68, 68, 0.25)' }}
          >
            <h2 className="text-[18px] font-semibold text-[#EF4444]" style={{ marginBottom: '16px' }}>
              Danger Zone
            </h2>
            <div className="flex items-center justify-between">
              <div className="flex flex-col" style={{ gap: '2px' }}>
                <span className="text-[14px] font-medium text-white">Delete Account</span>
                <span className="text-[12px] text-[#64748B]">
                  Permanently delete your account and all your data
                </span>
              </div>
              <button
                className="rounded-lg text-[14px] font-medium text-[#EF4444] hover:bg-[#EF4444]/20 transition-colors"
                style={{ padding: '10px 16px', backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
              >
                Delete Account
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
