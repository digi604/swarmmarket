import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Star,
  Clock,
  Calendar,
  ShieldCheck,
  MessageCircle,
  CheckCircle,
  BadgeCheck,
  Timer,
  Gavel,
  TrendingDown,
  Lock,
  Radio,
  Tag,
  TrendingUp,
  Zap,
  Shield,
  ChevronRight,
} from 'lucide-react';
import { ShareButton } from '../ui/ShareButton';
import { api } from '../../lib/api';
import type { Auction } from '../../lib/api';

const auctionTypeConfig = {
  english: { icon: Gavel, label: 'English Auction', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  dutch: { icon: TrendingDown, label: 'Dutch Auction', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  sealed: { icon: Lock, label: 'Sealed Bid', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
  continuous: { icon: Radio, label: 'Continuous', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
};

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  scheduled: { label: 'Scheduled', color: '#64748B', bgColor: 'rgba(100, 116, 139, 0.2)' },
  active: { label: 'Active', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
  ended: { label: 'Ended', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  cancelled: { label: 'Cancelled', color: '#EF4444', bgColor: 'rgba(239, 68, 68, 0.2)' },
};

interface Bid {
  id: string;
  bidder_name: string;
  amount: number;
  created_at: string;
}

export function AuctionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [auction, setAuction] = useState<Auction | null>(null);
  const [loading, setLoading] = useState(true);
  const [bidAmount, setBidAmount] = useState('');
  const [timeRemaining, setTimeRemaining] = useState('');

  // Mock bids for display (in a real app, these would come from the API)
  const mockBids: Bid[] = [
    { id: '1', bidder_name: 'DataBot Pro', amount: 12500, created_at: new Date(Date.now() - 2 * 60000).toISOString() },
    { id: '2', bidder_name: 'MLTrader X', amount: 12000, created_at: new Date(Date.now() - 15 * 60000).toISOString() },
    { id: '3', bidder_name: 'AI Agent 7', amount: 11500, created_at: new Date(Date.now() - 45 * 60000).toISOString() },
  ];

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const auctionData = await api.getAuction(id);
        setAuction(auctionData);

        // Set initial bid amount
        if (auctionData.current_price && auctionData.min_increment) {
          setBidAmount((auctionData.current_price + auctionData.min_increment).toString());
        } else if (auctionData.starting_price) {
          setBidAmount(auctionData.starting_price.toString());
        }
      } catch (error) {
        console.error('Failed to fetch auction:', error);
      }
      setLoading(false);
    };

    fetchData();
  }, [id]);

  // Update time remaining every second
  useEffect(() => {
    if (!auction) return;

    const updateTime = () => {
      const now = new Date();
      const ends = new Date(auction.ends_at);
      const diff = ends.getTime() - now.getTime();

      if (diff <= 0) {
        setTimeRemaining('Ended');
        return;
      }

      const hours = Math.floor(diff / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diff % (1000 * 60)) / 1000);

      if (hours > 24) {
        const days = Math.floor(hours / 24);
        setTimeRemaining(`${days}d ${hours % 24}h ${minutes}m`);
      } else {
        setTimeRemaining(`${hours}h ${minutes}m ${seconds}s`);
      }
    };

    updateTime();
    const interval = setInterval(updateTime, 1000);
    return () => clearInterval(interval);
  }, [auction]);

  const formatPrice = (amount: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  const getTimeAgo = (dateString: string) => {
    const now = new Date();
    const date = new Date(dateString);
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);

    if (diffHours > 0) return `${diffHours}h ago`;
    if (diffMins > 0) return `${diffMins} min ago`;
    return 'Just now';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#22D3EE]" />
      </div>
    );
  }

  if (!auction) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-[#94A3B8]">Auction not found</p>
        <button
          onClick={() => navigate(-1)}
          className="text-[#22D3EE] hover:underline"
        >
          Go back
        </button>
      </div>
    );
  }

  const type = auctionTypeConfig[auction.auction_type] || auctionTypeConfig.english;
  const TypeIcon = type.icon;
  const status = statusConfig[auction.status] || statusConfig.active;
  const isLive = auction.status === 'active';
  const reserveMet = auction.reserve_price && auction.current_price && auction.current_price >= auction.reserve_price;
  const minBid = auction.current_price && auction.min_increment
    ? auction.current_price + auction.min_increment
    : auction.starting_price;

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
              onClick={() => navigate('/dashboard/marketplace/auctions')}
              className="text-[#64748B] hover:text-white transition-colors"
            >
              Auctions
            </button>
            <span className="text-[#64748B]">/</span>
            <span className="text-white font-medium truncate max-w-[200px]">{auction.title}</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <ShareButton
            title={auction.title}
            text={`Check out "${auction.title}" on SwarmMarket`}
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
              {reserveMet && (
                <div
                  className="flex items-center gap-1"
                  style={{ backgroundColor: 'rgba(34, 211, 238, 0.2)', borderRadius: '6px', padding: '4px 10px' }}
                >
                  <CheckCircle className="w-3 h-3 text-[#22D3EE]" />
                  <span className="text-xs font-medium text-[#22D3EE]">Reserve Met</span>
                </div>
              )}
            </div>

            {/* Title & Description */}
            <div className="flex flex-col gap-3">
              <h1 className="text-2xl font-semibold text-white">{auction.title}</h1>
              <p className="text-[#94A3B8] text-sm leading-relaxed">{auction.description}</p>
            </div>

            {/* Timestamps */}
            <div className="flex items-center gap-3 text-sm text-[#64748B]">
              <Clock className="w-3.5 h-3.5" />
              <span>Started {getTimeAgo(auction.starts_at)}</span>
              <span className="w-1 h-1 rounded-full bg-[#64748B]" />
              <span>{auction.bid_count} bids placed</span>
            </div>

            {/* Pricing Row */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Tag className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Starting Price</span>
                </div>
                <span className="text-lg font-semibold text-[#94A3B8]">
                  {formatPrice(auction.starting_price, auction.currency)}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <TrendingUp className="w-3.5 h-3.5 text-[#22C55E]" />
                  <span className="text-xs font-medium text-[#64748B]">Current Bid</span>
                </div>
                <span className="text-lg font-semibold text-[#22C55E]">
                  {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Lock className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Reserve Price</span>
                </div>
                <span className="text-lg font-semibold" style={{ color: reserveMet ? '#22D3EE' : '#94A3B8' }}>
                  {auction.reserve_price
                    ? `${formatPrice(auction.reserve_price, auction.currency)} ${reserveMet ? '✓' : ''}`
                    : 'None'}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Zap className="w-3.5 h-3.5 text-[#F59E0B]" />
                  <span className="text-xs font-medium text-[#64748B]">Buy Now</span>
                </div>
                <span className="text-lg font-semibold text-[#F59E0B]">
                  {auction.buy_now_price ? formatPrice(auction.buy_now_price, auction.currency) : 'N/A'}
                </span>
              </div>
            </div>

            {/* Timing Row */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Timer className="w-3.5 h-3.5 text-[#EF4444]" />
                  <span className="text-xs font-medium text-[#64748B]">Time Remaining</span>
                </div>
                <span className="text-lg font-semibold text-[#EF4444]">{timeRemaining}</span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Ends At</span>
                </div>
                <span className="text-lg font-semibold text-white">{formatDate(auction.ends_at)}</span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <TrendingUp className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Min Increment</span>
                </div>
                <span className="text-lg font-semibold text-white">
                  {auction.min_increment ? formatPrice(auction.min_increment, auction.currency) : 'N/A'}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Shield className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Anti-Sniping</span>
                </div>
                <span className="text-lg font-semibold text-white">
                  {auction.extension_seconds > 0 ? `+${Math.floor(auction.extension_seconds / 60)} min` : 'None'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column */}
        <div className="w-[380px] flex flex-col gap-5">
          {/* Bid Card */}
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
              <h3 className="text-lg font-semibold text-white">Place Bid</h3>
              {isLive && (
                <div className="flex items-center gap-1.5">
                  <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                  </span>
                  <span className="text-xs font-medium text-[#22C55E]">Live</span>
                </div>
              )}
            </div>

            {/* Current High Bid */}
            <div
              className="flex flex-col gap-1"
              style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
            >
              <span className="text-xs font-medium text-[#64748B]">Current High Bid</span>
              <div className="flex items-center justify-between">
                <span className="text-3xl font-bold text-[#22C55E]">
                  {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
                </span>
                <div className="flex items-center gap-2">
                  <div
                    className="w-6 h-6 rounded-full"
                    style={{
                      background: 'linear-gradient(135deg, #22D3EE 0%, #06B6D4 100%)',
                    }}
                  />
                  <span className="text-[#94A3B8] text-sm font-medium">DataBot Pro</span>
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-[#94A3B8] text-sm font-medium">
                Your bid (min {formatPrice(minBid, auction.currency)})
              </span>
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
                  disabled={!isLive}
                >
                  Place Bid
                </button>
              </div>
            </div>

            <div className="h-px bg-[#334155]" />

            {/* Bid History */}
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold text-white">Bid History</span>
                <span className="text-xs text-[#64748B]">{auction.bid_count} total</span>
              </div>

              {mockBids.map((bid) => (
                <div
                  key={bid.id}
                  className="flex items-center justify-between"
                  style={{ backgroundColor: '#0F172A', borderRadius: '10px', padding: '12px' }}
                >
                  <div className="flex items-center gap-2.5">
                    <div
                      className="w-8 h-8 rounded-full"
                      style={{
                        background: bid.id === '1'
                          ? 'linear-gradient(135deg, #22D3EE 0%, #06B6D4 100%)'
                          : 'linear-gradient(135deg, #A855F7 0%, #7C3AED 100%)',
                      }}
                    />
                    <div className="flex flex-col">
                      <span className="text-white text-sm font-medium">{bid.bidder_name}</span>
                      <span className="text-[#64748B] text-xs">{getTimeAgo(bid.created_at)}</span>
                    </div>
                  </div>
                  <span
                    className="font-semibold text-sm"
                    style={{ color: bid.id === '1' ? '#22C55E' : '#94A3B8' }}
                  >
                    {formatPrice(bid.amount, auction.currency)}
                  </span>
                </div>
              ))}

              <button
                className="w-full h-9 rounded-lg flex items-center justify-center gap-1.5 transition-colors hover:bg-[#334155]"
                style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
              >
                <span className="text-[#94A3B8] text-sm font-medium">View all {auction.bid_count} bids</span>
                <ChevronRight className="w-3.5 h-3.5 text-[#64748B]" />
              </button>
            </div>
          </div>

          {/* Seller Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <span className="text-xs font-medium text-[#64748B]">Seller</span>

            <div className="flex items-center gap-3">
              {auction.seller_avatar_url ? (
                <img
                  src={auction.seller_avatar_url}
                  alt={auction.seller_name || 'Agent'}
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
                    {auction.seller_name?.[0]?.toUpperCase() || 'A'}
                  </span>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5">
                  <span className="text-white font-semibold">{auction.seller_name || 'Agent'}</span>
                  <BadgeCheck className="w-4 h-4 text-[#22D3EE]" />
                </div>
                <div className="flex items-center gap-1">
                  <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
                  <span className="text-[#F59E0B] text-xs font-medium">
                    {auction.seller_rating ? auction.seller_rating.toFixed(1) : '—'}
                  </span>
                </div>
              </div>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-3 gap-4">
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">89</span>
                <span className="text-xs text-[#64748B]">Auctions</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-[#22C55E]">100%</span>
                <span className="text-xs text-[#64748B]">Delivery</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">$892K</span>
                <span className="text-xs text-[#64748B]">Total Sales</span>
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
                <span className="text-base font-semibold text-[#22D3EE]">0.98 / 1.0</span>
              </div>
            </div>

            <button
              className="w-full h-11 rounded-lg flex items-center justify-center gap-2 transition-colors hover:bg-[#334155]"
              style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
            >
              <MessageCircle className="w-4 h-4 text-[#94A3B8]" />
              <span className="text-white text-sm font-medium">Contact Seller</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
