import { useState, useEffect, useCallback } from 'react';
import { CreditCard, Trash2, Loader2, Plus, Star, X } from 'lucide-react';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js';
import { api } from '../../lib/api';
import type { PaymentMethodInfo } from '../../lib/api';

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY);

const appearance = {
  theme: 'night' as const,
  variables: {
    colorPrimary: '#A855F7',
    colorBackground: '#1E293B',
    colorText: '#F8FAFC',
    colorDanger: '#EF4444',
    fontFamily: 'inherit',
    borderRadius: '8px',
    colorTextSecondary: '#94A3B8',
    colorTextPlaceholder: '#64748B',
  },
};

function SetupForm({ onSuccess, onClose }: { onSuccess: () => void; onClose: () => void }) {
  const stripe = useStripe();
  const elements = useElements();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    setSubmitting(true);
    setError(null);

    const result = await stripe.confirmSetup({
      elements,
      redirect: 'if_required',
    });

    if (result.error) {
      setError(result.error.message ?? 'Something went wrong');
      setSubmitting(false);
    } else {
      onSuccess();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col" style={{ gap: '20px' }}>
      <PaymentElement />
      {error && (
        <div
          className="rounded-lg text-[13px] text-red-400"
          style={{ padding: '12px 16px', backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
        >
          {error}
        </div>
      )}
      <div className="flex items-center justify-end" style={{ gap: '12px' }}>
        <button
          type="button"
          onClick={onClose}
          disabled={submitting}
          className="rounded-lg text-[14px] font-medium text-[#94A3B8] hover:text-white transition-colors disabled:opacity-50"
          style={{ padding: '10px 20px' }}
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={submitting || !stripe || !elements}
          className="flex items-center rounded-lg text-[14px] font-medium text-white bg-[#A855F7] hover:bg-[#9333EA] transition-colors disabled:opacity-50"
          style={{ padding: '10px 20px', gap: '8px' }}
        >
          {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
          Save payment method
        </button>
      </div>
    </form>
  );
}

function SetupFormModal({
  clientSecret,
  onSuccess,
  onClose,
}: {
  clientSecret: string;
  onSuccess: () => void;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{ backgroundColor: 'rgba(0, 0, 0, 0.6)' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div
        className="relative rounded-xl w-full"
        style={{ maxWidth: '480px', padding: '28px', backgroundColor: '#0F172A', border: '1px solid #1E293B' }}
      >
        <div className="flex items-center justify-between" style={{ marginBottom: '24px' }}>
          <h3 className="text-[16px] font-medium text-white">Add payment method</h3>
          <button onClick={onClose} className="text-[#64748B] hover:text-white transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        <Elements stripe={stripePromise} options={{ clientSecret, appearance }}>
          <SetupForm onSuccess={onSuccess} onClose={onClose} />
        </Elements>
      </div>
    </div>
  );
}

export function PaymentMethodsSection() {
  const [methods, setMethods] = useState<PaymentMethodInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [clientSecret, setClientSecret] = useState<string | null>(null);

  const fetchMethods = useCallback(async () => {
    try {
      const data = await api.listPaymentMethods();
      setMethods(data || []);
    } catch {
      setError('Failed to load payment methods');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMethods();
  }, [fetchMethods]);

  const handleAddCard = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await api.createSetupIntent();
      setClientSecret(result.client_secret);
    } catch {
      setError('Failed to start card setup');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSetupSuccess = () => {
    setClientSecret(null);
    fetchMethods();
  };

  const handleDelete = async (id: string) => {
    setActionLoading(true);
    setError(null);
    try {
      await api.deletePaymentMethod(id);
      setMethods((prev) => prev.filter((m) => m.id !== id));
    } catch {
      setError('Failed to remove payment method');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    setActionLoading(true);
    setError(null);
    try {
      await api.setDefaultPaymentMethod(id);
      setMethods((prev) =>
        prev.map((m) => ({ ...m, is_default: m.id === id }))
      );
    } catch {
      setError('Failed to set default payment method');
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

  return (
    <div className="flex flex-col" style={{ gap: '24px' }}>
      <p className="text-[13px] text-[#64748B]">
        Save a payment method so your agents can make purchases automatically.
      </p>

      {error && (
        <div
          className="rounded-lg text-[13px] text-red-400"
          style={{ padding: '12px 16px', backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
        >
          {error}
        </div>
      )}

      {methods.length > 0 && (
        <div className="flex flex-col rounded-lg overflow-hidden" style={{ backgroundColor: '#0F172A' }}>
          {methods.map((method, index) => (
            <div
              key={method.id}
              className="flex items-center justify-between"
              style={{
                padding: '16px',
                borderBottom: index < methods.length - 1 ? '1px solid #1E293B' : 'none',
              }}
            >
              <div className="flex items-center" style={{ gap: '12px' }}>
                <CreditCard className="w-5 h-5 text-[#64748B]" />
                <div className="flex flex-col" style={{ gap: '2px' }}>
                  <div className="flex items-center" style={{ gap: '8px' }}>
                    <span className="text-[14px] font-medium text-white">
                      {method.brand} ····{method.last4}
                    </span>
                    {method.is_default && (
                      <span
                        className="text-[11px] font-medium rounded-full"
                        style={{ padding: '2px 8px', backgroundColor: 'rgba(168, 85, 247, 0.2)', color: '#A855F7' }}
                      >
                        Default
                      </span>
                    )}
                  </div>
                  <span className="text-[12px] text-[#64748B]">
                    Expires {method.exp_month}/{method.exp_year}
                  </span>
                </div>
              </div>
              <div className="flex items-center" style={{ gap: '8px' }}>
                {!method.is_default && (
                  <button
                    onClick={() => handleSetDefault(method.id)}
                    disabled={actionLoading}
                    className="flex items-center text-[13px] text-[#94A3B8] hover:text-white transition-colors disabled:opacity-50"
                    style={{ padding: '6px 10px', gap: '4px' }}
                  >
                    <Star className="w-3.5 h-3.5" />
                    Set default
                  </button>
                )}
                <button
                  onClick={() => handleDelete(method.id)}
                  disabled={actionLoading}
                  className="flex items-center text-[13px] text-[#64748B] hover:text-red-400 transition-colors disabled:opacity-50"
                  style={{ padding: '6px 10px', gap: '4px' }}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {methods.length === 0 ? (
        <button
          onClick={handleAddCard}
          disabled={actionLoading}
          className="flex items-center rounded-lg text-[14px] font-medium text-white bg-[#A855F7] hover:bg-[#9333EA] transition-colors disabled:opacity-50 self-start"
          style={{ padding: '10px 20px', gap: '8px' }}
        >
          {actionLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4" />}
          Add payment method
        </button>
      ) : (
        <button
          onClick={handleAddCard}
          disabled={actionLoading}
          className="flex items-center text-[13px] text-[#94A3B8] hover:text-white transition-colors disabled:opacity-50 self-start"
          style={{ gap: '4px' }}
        >
          {actionLoading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Plus className="w-3.5 h-3.5" />}
          Add another
        </button>
      )}

      {clientSecret && (
        <SetupFormModal
          clientSecret={clientSecret}
          onSuccess={handleSetupSuccess}
          onClose={() => setClientSecret(null)}
        />
      )}
    </div>
  );
}
