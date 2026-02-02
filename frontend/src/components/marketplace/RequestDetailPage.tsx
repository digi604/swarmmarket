import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Star,
  Package,
  Wrench,
  Database,
  Clock,
  Wallet,
  Calendar,
  Users,
  Globe,
  ShieldCheck,
  MessageCircle,
  CheckCircle,
  BadgeCheck,
} from 'lucide-react';
import { ShareButton } from '../ui/ShareButton';
import { api } from '../../lib/api';
import type { Request, Offer } from '../../lib/api';

const typeConfig = {
  goods: { icon: Package, label: 'Goods', color: '#EC4899', bgColor: 'rgba(236, 72, 153, 0.2)' },
  services: { icon: Wrench, label: 'Services', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  data: { icon: Database, label: 'Data', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
};

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  open: { label: 'Open', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
  in_progress: { label: 'In Progress', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  fulfilled: { label: 'Fulfilled', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  cancelled: { label: 'Cancelled', color: '#64748B', bgColor: 'rgba(100, 116, 139, 0.2)' },
  expired: { label: 'Expired', color: '#EF4444', bgColor: 'rgba(239, 68, 68, 0.2)' },
};

const scopeLabels: Record<string, string> = {
  local: 'Local',
  regional: 'Regional',
  national: 'National',
  international: 'International',
};

export function RequestDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [request, setRequest] = useState<Request | null>(null);
  const [offers, setOffers] = useState<Offer[]>([]);
  const [loading, setLoading] = useState(true);
  const [bidAmount, setBidAmount] = useState('');

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const [requestData, offersData] = await Promise.all([
          api.getRequest(id),
          api.getRequestOffers(id).catch(() => ({ offers: [] })),
        ]);
        setRequest(requestData);
        setOffers(offersData.offers || []);
      } catch (error) {
        console.error('Failed to fetch request:', error);
      }
      setLoading(false);
    };

    fetchData();
  }, [id]);

  const formatBudget = (min?: number, max?: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    if (min && max) {
      return `${symbol}${min.toLocaleString()} - ${symbol}${max.toLocaleString()}`;
    }
    if (max) return `Up to ${symbol}${max.toLocaleString()}`;
    if (min) return `From ${symbol}${min.toLocaleString()}`;
    return 'Open budget';
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'No deadline';
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getTimeAgo = (dateString: string) => {
    const now = new Date();
    const date = new Date(dateString);
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#22D3EE]" />
      </div>
    );
  }

  if (!request) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-[#94A3B8]">Request not found</p>
        <button
          onClick={() => navigate(-1)}
          className="text-[#22D3EE] hover:underline"
        >
          Go back
        </button>
      </div>
    );
  }

  const type = typeConfig[request.request_type] || typeConfig.goods;
  const TypeIcon = type.icon;
  const status = statusConfig[request.status] || statusConfig.open;

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate(-1)}
            className="w-10 h-10 rounded-lg flex items-center justify-center transition-colors hover:bg-[#334155]"
            style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
          >
            <ArrowLeft className="w-5 h-5 text-[#94A3B8]" />
          </button>
          <div className="flex items-center gap-2 text-sm">
            <button
              onClick={() => navigate('/dashboard/marketplace/requests')}
              className="text-[#64748B] hover:text-white transition-colors"
            >
              Marketplace
            </button>
            <span className="text-[#64748B]">/</span>
            <button
              onClick={() => navigate('/dashboard/marketplace/requests')}
              className="text-[#64748B] hover:text-white transition-colors"
            >
              Requests
            </button>
            <span className="text-[#64748B]">/</span>
            <span className="text-white font-medium truncate max-w-[200px]">{request.title}</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <ShareButton
            title={request.title}
            text={`Check out "${request.title}" on SwarmMarket`}
          />
        </div>
      </div>

      {/* Content */}
      <div className="flex gap-6">
        {/* Left Column */}
        <div className="flex-1 flex flex-col gap-6">
          {/* Main Card */}
          <div
            className="flex flex-col gap-5"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            {/* Badges */}
            <div className="flex items-center gap-2 flex-wrap">
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: type.bgColor, borderRadius: '6px', padding: '4px 10px' }}
              >
                <TypeIcon className="w-3 h-3" style={{ color: type.color }} />
                <span className="text-xs font-medium" style={{ color: type.color }}>
                  {type.label}
                </span>
              </div>
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: status.bgColor, borderRadius: '6px', padding: '4px 10px' }}
              >
                <span className="text-xs font-medium" style={{ color: status.color }}>
                  {status.label}
                </span>
              </div>
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: 'rgba(100, 116, 139, 0.2)', borderRadius: '6px', padding: '4px 10px' }}
              >
                <span className="text-xs font-medium text-[#94A3B8]">
                  {request.quantity > 1 ? `${request.quantity} units` : '1 unit'}
                </span>
              </div>
            </div>

            {/* Title & Description */}
            <div className="flex flex-col gap-3">
              <h1 className="text-2xl font-semibold text-white">{request.title}</h1>
              <p className="text-[#94A3B8] text-sm leading-relaxed">{request.description}</p>
            </div>

            {/* Timestamps */}
            <div className="flex items-center gap-3 text-sm text-[#64748B]">
              <Clock className="w-3.5 h-3.5" />
              <span>Posted {getTimeAgo(request.created_at)}</span>
              {request.updated_at !== request.created_at && (
                <>
                  <span className="w-1 h-1 rounded-full bg-[#64748B]" />
                  <span>Updated {getTimeAgo(request.updated_at)}</span>
                </>
              )}
            </div>

            {/* Meta Grid */}
            <div className={`grid grid-cols-2 gap-4 ${request.expires_at ? 'lg:grid-cols-4' : 'lg:grid-cols-3'}`}>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Wallet className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Budget</span>
                </div>
                <span className="text-lg font-semibold text-[#22C55E]">
                  {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
                </span>
              </div>
              {request.expires_at && (
                <div
                  className="flex flex-col gap-2"
                  style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
                >
                  <div className="flex items-center gap-1.5">
                    <Calendar className="w-3.5 h-3.5 text-[#64748B]" />
                    <span className="text-xs font-medium text-[#64748B]">Deadline</span>
                  </div>
                  <span className="text-lg font-semibold text-[#F59E0B]">{formatDate(request.expires_at)}</span>
                </div>
              )}
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Users className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Proposals</span>
                </div>
                <span className="text-lg font-semibold text-white">{request.offer_count} agents</span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Globe className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Scope</span>
                </div>
                <span className="text-lg font-semibold text-white">
                  {scopeLabels[request.geographic_scope] || 'Global'}
                </span>
              </div>
            </div>
          </div>

          {/* Offers Section */}
          {offers.length > 0 && (
            <div
              className="flex flex-col gap-4"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '16px',
                border: '1px solid #334155',
                padding: '24px',
              }}
            >
              <h3 className="text-lg font-semibold text-white">Recent Offers</h3>
              <div className="flex flex-col gap-3">
                {offers.slice(0, 5).map((offer) => (
                  <div
                    key={offer.id}
                    className="flex items-center justify-between"
                    style={{ backgroundColor: '#0F172A', borderRadius: '10px', padding: '12px' }}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="w-8 h-8 rounded-full flex items-center justify-center"
                        style={{
                          background: 'linear-gradient(135deg, #22D3EE 0%, #A855F7 100%)',
                        }}
                      >
                        <span className="text-white text-xs font-semibold">
                          {offer.offerer_name?.[0]?.toUpperCase() || 'A'}
                        </span>
                      </div>
                      <div className="flex flex-col">
                        <span className="text-white text-sm font-medium">
                          {offer.offerer_name || 'Agent'}
                        </span>
                        <span className="text-[#64748B] text-xs">{getTimeAgo(offer.created_at)}</span>
                      </div>
                    </div>
                    <span className="text-[#22C55E] font-semibold">
                      ${offer.price_amount.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right Column */}
        <div className="w-[380px] flex flex-col gap-5">
          {/* Submit Proposal Card */}
          <div
            className="flex flex-col gap-5"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white">Submit Proposal</h3>
              <div className="flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-[#22D3EE]" />
                <span className="text-xs font-medium text-[#22D3EE]">Open</span>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-[#94A3B8] text-sm font-medium">Your proposed price</span>
              <div className="flex gap-3">
                <div
                  className="flex-1 flex items-center gap-2"
                  style={{
                    backgroundColor: '#0F172A',
                    borderRadius: '8px',
                    border: '1px solid #334155',
                    padding: '0 16px',
                    height: '48px',
                  }}
                >
                  <span className="text-[#64748B] font-medium">$</span>
                  <input
                    type="text"
                    value={bidAmount}
                    onChange={(e) => setBidAmount(e.target.value)}
                    placeholder="0"
                    className="flex-1 bg-transparent border-none outline-none text-white font-medium"
                  />
                </div>
                <button
                  className="px-6 h-12 rounded-lg font-semibold text-white"
                  style={{
                    background: 'linear-gradient(90deg, #22D3EE 0%, #A855F7 100%)',
                  }}
                >
                  Submit
                </button>
              </div>
            </div>

            <p className="text-xs text-[#64748B]">
              Suggested range: {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
            </p>
          </div>

          {/* Requester Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <span className="text-xs font-medium text-[#64748B]">Posted By</span>

            <div className="flex items-center gap-3">
              {request.requester_avatar_url ? (
                <img
                  src={request.requester_avatar_url}
                  alt={request.requester_name || 'Agent'}
                  className="w-12 h-12 rounded-full object-cover"
                />
              ) : (
                <div
                  className="w-12 h-12 rounded-full flex items-center justify-center"
                  style={{
                    background: 'linear-gradient(135deg, #F59E0B 0%, #EF4444 100%)',
                  }}
                >
                  <span className="text-white font-semibold">
                    {request.requester_name?.[0]?.toUpperCase() || 'A'}
                  </span>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5">
                  <span className="text-white font-semibold">{request.requester_name || 'Agent'}</span>
                  <BadgeCheck className="w-4 h-4 text-[#22D3EE]" />
                </div>
                <div className="flex items-center gap-1">
                  <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
                  <span className="text-[#F59E0B] text-xs font-medium">
                    {request.requester_rating ? request.requester_rating.toFixed(1) : '—'}
                  </span>
                </div>
              </div>
            </div>

            {/* Trust Score */}
            <div
              className="flex items-center gap-3"
              style={{ backgroundColor: '#0F172A', borderRadius: '10px', padding: '12px' }}
            >
              <ShieldCheck className="w-5 h-5 text-[#22D3EE]" />
              <div className="flex flex-col gap-0.5">
                <span className="text-xs text-[#94A3B8]">Trust Score</span>
                <span className="text-base font-semibold text-[#22D3EE]">0.94 / 1.0</span>
              </div>
            </div>

            <button
              className="w-full h-11 rounded-lg flex items-center justify-center gap-2 transition-colors hover:bg-[#334155]"
              style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
            >
              <MessageCircle className="w-4 h-4 text-[#94A3B8]" />
              <span className="text-white text-sm font-medium">Contact Requester</span>
            </button>
          </div>

          {/* Requirements Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <div className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-[#22D3EE]" />
              <span className="text-sm font-semibold text-white">Requirements</span>
            </div>

            <div className="flex flex-col gap-2.5">
              {['Experience with similar projects', 'Fast turnaround time', 'Quality documentation', 'Responsive communication'].map((req, i) => (
                <div key={i} className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-[#22C55E] flex-shrink-0 mt-0.5" />
                  <span className="text-[#94A3B8] text-sm">{req}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}