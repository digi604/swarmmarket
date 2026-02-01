import { useState } from 'react';
import {
  ArrowUpRight,
  ArrowDownLeft,
  Loader2,
  CreditCard,
  Plus,
  X,
  AlertCircle,
  CheckCircle,
  Clock,
} from 'lucide-react';
import { useAgentTransactions } from '../../../hooks/useDashboard';
import { api } from '../../../lib/api';
import type { Transaction, Deposit } from '../../../lib/api';

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

function TransactionRow({
  tx,
  isLast,
}: {
  tx: Transaction;
  isLast: boolean;
}) {
  const isIncoming = true; // Agent transactions are seller transactions (incoming)
  const title = tx.status === 'escrow_funded' ? 'Escrow received' : 'Payment received';
  const description = `From ${tx.buyer_name || 'buyer'} • ${formatTimeAgo(tx.created_at)}`;

  let amount = `+$${tx.amount.toFixed(2)}`;
  let amountColor = '#22C55E';

  if (tx.status === 'escrow_funded') {
    amount = `$${tx.amount.toFixed(2)}`;
    amountColor = '#F59E0B';
  } else if (tx.status === 'refunded' || tx.status === 'cancelled') {
    amount = `-$${tx.amount.toFixed(2)}`;
    amountColor = '#EF4444';
  }

  return (
    <div
      className="flex items-center justify-between"
      style={{
        padding: '16px 20px',
        borderBottom: !isLast ? '1px solid #334155' : 'none',
      }}
    >
      <div className="flex items-center" style={{ gap: '12px' }}>
        <div
          className="rounded-full flex items-center justify-center"
          style={{
            width: '40px',
            height: '40px',
            backgroundColor: isIncoming ? 'rgba(34, 197, 94, 0.125)' : 'rgba(239, 68, 68, 0.125)',
          }}
        >
          {isIncoming ? (
            <ArrowDownLeft className="w-5 h-5 text-[#22C55E]" />
          ) : (
            <ArrowUpRight className="w-5 h-5 text-[#EF4444]" />
          )}
        </div>
        <div className="flex flex-col" style={{ gap: '2px' }}>
          <span className="text-[14px] font-medium text-white">{title}</span>
          <span className="text-[12px] text-[#64748B]">{description}</span>
        </div>
      </div>
      <span
        className="font-mono text-[14px] font-semibold"
        style={{ color: amountColor }}
      >
        {amount}
      </span>
    </div>
  );
}

function DepositRow({ deposit, isLast }: { deposit: Deposit; isLast: boolean }) {
  const statusConfig = {
    pending: { color: '#F59E0B', icon: Clock, label: 'Pending' },
    processing: { color: '#A855F7', icon: Loader2, label: 'Processing' },
    completed: { color: '#22C55E', icon: CheckCircle, label: 'Completed' },
    failed: { color: '#EF4444', icon: AlertCircle, label: 'Failed' },
    cancelled: { color: '#64748B', icon: X, label: 'Cancelled' },
  };

  const status = statusConfig[deposit.status] || statusConfig.pending;
  const StatusIcon = status.icon;

  return (
    <div
      className="flex items-center justify-between"
      style={{
        padding: '16px 20px',
        borderBottom: !isLast ? '1px solid #334155' : 'none',
      }}
    >
      <div className="flex items-center" style={{ gap: '12px' }}>
        <div
          className="rounded-full flex items-center justify-center"
          style={{
            width: '40px',
            height: '40px',
            backgroundColor: 'rgba(34, 211, 238, 0.125)',
          }}
        >
          <Plus className="w-5 h-5 text-[#22D3EE]" />
        </div>
        <div className="flex flex-col" style={{ gap: '2px' }}>
          <span className="text-[14px] font-medium text-white">Deposit</span>
          <div className="flex items-center" style={{ gap: '6px' }}>
            <StatusIcon
              className={`w-3 h-3 ${deposit.status === 'processing' ? 'animate-spin' : ''}`}
              style={{ color: status.color }}
            />
            <span className="text-[12px]" style={{ color: status.color }}>
              {status.label}
            </span>
            <span className="text-[12px] text-[#64748B]">
              • {formatTimeAgo(deposit.created_at)}
            </span>
          </div>
        </div>
      </div>
      <span
        className="font-mono text-[14px] font-semibold"
        style={{ color: deposit.status === 'completed' ? '#22C55E' : '#64748B' }}
      >
        +${deposit.amount.toFixed(2)}
      </span>
    </div>
  );
}

function DepositModal({
  isOpen,
  onClose,
  onSuccess,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [depositResult, setDepositResult] = useState<{
    checkoutUrl: string;
    instructions: string;
  } | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const numAmount = parseFloat(amount);
    if (isNaN(numAmount) || numAmount <= 0) {
      setError('Please enter a valid amount');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const result = await api.createDeposit(numAmount, 'USD', window.location.href);
      if (result.checkout_url) {
        // Redirect to Stripe Checkout
        window.location.href = result.checkout_url;
      } else {
        setDepositResult({
          checkoutUrl: result.checkout_url,
          instructions: result.instructions,
        });
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create deposit');
      setLoading(false);
    }
  };

  const handleClose = () => {
    setAmount('');
    setError(null);
    setDepositResult(null);
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

        <h2 className="text-xl font-bold text-white mb-2">Deposit Funds</h2>
        <p className="text-sm text-[#94A3B8] mb-6">
          Add funds to your agent's wallet via Stripe.
        </p>

        {depositResult ? (
          <div className="flex flex-col items-center py-4">
            <CheckCircle className="w-12 h-12 text-[#22C55E] mb-4" />
            <p className="text-lg font-semibold text-white mb-2">Deposit Created!</p>
            <p className="text-sm text-[#94A3B8] text-center mb-4">
              {depositResult.instructions}
            </p>
            <button
              onClick={() => {
                onSuccess();
                handleClose();
              }}
              className="mt-6 px-6 py-3 rounded-lg bg-[#22D3EE] text-[#0A0F1C] font-semibold hover:bg-[#06B6D4] transition-colors"
            >
              Done
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label className="block text-sm font-medium text-[#94A3B8] mb-2">
                Amount (USD)
              </label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-[#64748B] font-mono">
                  $
                </span>
                <input
                  type="number"
                  step="0.01"
                  min="1"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="100.00"
                  className="w-full pl-8 pr-4 py-3 bg-[#0F172A] border border-[#334155] rounded-lg text-white placeholder-[#64748B] focus:outline-none focus:border-[#22D3EE] font-mono text-lg"
                  disabled={loading}
                />
              </div>
            </div>

            {/* Quick amounts */}
            <div className="flex gap-2 mb-6">
              {[10, 50, 100, 500].map((preset) => (
                <button
                  key={preset}
                  type="button"
                  onClick={() => setAmount(preset.toString())}
                  className="flex-1 py-2 rounded-lg bg-[#0F172A] text-[#94A3B8] text-sm font-medium hover:bg-[#334155] transition-colors"
                >
                  ${preset}
                </button>
              ))}
            </div>

            {error && (
              <div className="mb-4 flex items-center gap-2 text-red-400 text-sm">
                <AlertCircle className="w-4 h-4" />
                <span>{error}</span>
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !amount}
              className="w-full py-3 rounded-lg bg-[#22D3EE] text-[#0A0F1C] font-semibold hover:bg-[#06B6D4] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Creating deposit...
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4" />
                  Deposit ${amount || '0.00'}
                </>
              )}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}

interface WalletTabProps {
  agentId: string;
}

export function WalletTab({ agentId }: WalletTabProps) {
  const { transactions, loading } = useAgentTransactions(agentId);
  const [showDepositModal, setShowDepositModal] = useState(false);

  // Agent deposits would need a dedicated API endpoint
  // For now, we use an empty array as agents deposit via API, not UI
  const deposits: Deposit[] = [];

  // Suppress unused variable warning - agentId will be used when we add agent deposit API
  void agentId;

  if (loading) {
    return (
      <div className="flex items-center justify-center flex-1">
        <Loader2 className="w-8 h-8 text-[#22D3EE] animate-spin" />
      </div>
    );
  }

  // Calculate wallet stats for this agent
  const completedAmount = transactions
    .filter((tx) => tx.status === 'completed')
    .reduce((sum, tx) => sum + tx.amount, 0);

  const depositedAmount = deposits
    .filter((d) => d.status === 'completed')
    .reduce((sum, d) => sum + d.amount, 0);

  const inEscrowAmount = transactions
    .filter((tx) => tx.status === 'escrow_funded')
    .reduce((sum, tx) => sum + tx.amount, 0);

  const pendingAmount = transactions
    .filter((tx) => tx.status === 'delivered')
    .reduce((sum, tx) => sum + tx.amount, 0);

  const pendingDeposits = deposits
    .filter((d) => d.status === 'pending' || d.status === 'processing')
    .reduce((sum, d) => sum + d.amount, 0);

  const totalEarned = completedAmount;
  const availableBalance = completedAmount + depositedAmount;

  const recentTransactions = transactions
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 10);

  const recentDeposits = deposits
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5);

  return (
    <div className="flex flex-col flex-1" style={{ gap: '24px' }}>
      {/* Top Row */}
      <div className="flex" style={{ gap: '20px' }}>
        {/* Balance Card */}
        <div
          className="flex-1 rounded-2xl flex flex-col justify-between"
          style={{
            background: 'linear-gradient(135deg, #1E293B 0%, #0F172A 100%)',
            padding: '24px',
            border: '1px solid #334155',
          }}
        >
          <div className="flex flex-col" style={{ gap: '20px' }}>
            <div className="flex flex-col" style={{ gap: '8px' }}>
              <span className="text-[14px] font-medium text-[#64748B]">Available Balance</span>
              <span className="font-mono text-[48px] font-bold text-white leading-none">
                ${availableBalance.toFixed(2)}
              </span>
            </div>
            <div className="flex" style={{ gap: '12px' }}>
              <button
                onClick={() => setShowDepositModal(true)}
                className="flex items-center rounded-lg bg-[#22D3EE] text-[#0A0F1C] text-[14px] font-semibold hover:bg-[#06B6D4] transition-colors"
                style={{ padding: '12px 20px', gap: '8px' }}
              >
                <Plus className="w-4 h-4" />
                Deposit
              </button>
              <button
                className="flex items-center rounded-lg bg-[#1E293B] text-white text-[14px] font-semibold hover:bg-[#2D3B4F] transition-colors"
                style={{ padding: '12px 20px', gap: '8px', border: '1px solid #334155' }}
              >
                <ArrowUpRight className="w-4 h-4" />
                Withdraw
              </button>
            </div>
          </div>
        </div>

        {/* Stats Column */}
        <div className="flex flex-col" style={{ width: '280px', gap: '12px' }}>
          <div
            className="rounded-xl bg-[#1E293B] flex flex-col"
            style={{ padding: '16px', gap: '4px' }}
          >
            <span className="text-[12px] font-medium text-[#64748B]">Total Earned</span>
            <span className="font-mono text-[20px] font-bold text-[#22C55E]">
              ${totalEarned.toFixed(2)}
            </span>
          </div>
          <div
            className="rounded-xl bg-[#1E293B] flex flex-col"
            style={{ padding: '16px', gap: '4px' }}
          >
            <span className="text-[12px] font-medium text-[#64748B]">In Escrow</span>
            <span className="font-mono text-[20px] font-bold text-[#F59E0B]">
              ${inEscrowAmount.toFixed(2)}
            </span>
          </div>
          <div
            className="rounded-xl bg-[#1E293B] flex flex-col"
            style={{ padding: '16px', gap: '4px' }}
          >
            <span className="text-[12px] font-medium text-[#64748B]">Pending</span>
            <span className="font-mono text-[20px] font-bold text-[#A855F7]">
              ${(pendingAmount + pendingDeposits).toFixed(2)}
            </span>
          </div>
        </div>
      </div>

      {/* Transactions Section */}
      <div className="flex flex-col flex-1" style={{ gap: '16px' }}>
        <div className="flex items-center justify-between">
          <h2 className="text-[18px] font-semibold text-white">Recent Activity</h2>
          <button className="text-[14px] font-medium text-[#A855F7] hover:text-[#9333EA] transition-colors">
            View All →
          </button>
        </div>

        {recentTransactions.length === 0 && recentDeposits.length === 0 ? (
          <div
            className="flex-1 rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
            style={{ padding: '48px 16px' }}
          >
            <CreditCard className="w-12 h-12 text-[#64748B]" style={{ marginBottom: '16px' }} />
            <p className="text-[16px] font-medium text-white" style={{ marginBottom: '4px' }}>
              No activity yet
            </p>
            <p className="text-[14px] text-[#64748B] text-center" style={{ marginBottom: '16px' }}>
              Make a deposit or complete a transaction to see activity here
            </p>
            <button
              onClick={() => setShowDepositModal(true)}
              className="flex items-center rounded-lg bg-[#22D3EE] text-[#0A0F1C] text-[14px] font-semibold hover:bg-[#06B6D4] transition-colors"
              style={{ padding: '10px 16px', gap: '6px' }}
            >
              <Plus className="w-4 h-4" />
              Make First Deposit
            </button>
          </div>
        ) : (
          <div className="rounded-xl bg-[#1E293B] overflow-hidden flex-1">
            {/* Show deposits first, then transactions */}
            {recentDeposits.map((deposit, index) => (
              <DepositRow
                key={deposit.id}
                deposit={deposit}
                isLast={index === recentDeposits.length - 1 && recentTransactions.length === 0}
              />
            ))}
            {recentTransactions.map((tx, index) => (
              <TransactionRow
                key={tx.id}
                tx={tx}
                isLast={index === recentTransactions.length - 1}
              />
            ))}
          </div>
        )}
      </div>

      {/* Deposit Modal */}
      <DepositModal
        isOpen={showDepositModal}
        onClose={() => setShowDepositModal(false)}
        onSuccess={() => {
          // Refresh deposits
          // In a real implementation, you'd refetch the deposits here
        }}
      />
    </div>
  );
}
