import { Star, Timer, Gavel, TrendingDown, Lock, Radio, Zap } from 'lucide-react';
import type { Auction } from '../../lib/api';

const auctionTypeConfig = {
  english: { icon: Gavel, label: 'English', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  dutch: { icon: TrendingDown, label: 'Dutch', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  sealed: { icon: Lock, label: 'Sealed', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
  continuous: { icon: Radio, label: 'Continuous', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
};

interface AuctionCardProps {
  auction: Auction;
  onClick?: () => void;
}

export function AuctionCard({ auction, onClick }: AuctionCardProps) {
  const type = auctionTypeConfig[auction.auction_type] || auctionTypeConfig.english;
  const TypeIcon = type.icon;

  const formatPrice = (amount: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

  const getTimeRemaining = (endsAt?: string) => {
    if (!endsAt) return null;
    const now = new Date();
    const ends = new Date(endsAt);
    if (isNaN(ends.getTime())) return null;
    const diff = ends.getTime() - now.getTime();
    if (diff <= 0) return 'Ended';
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    const days = Math.floor(hours / 24);
    if (days > 0) return `${days}d ${hours % 24}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const isLive = auction.status === 'active';
  const timeRemaining = getTimeRemaining(auction.ends_at);

  return (
    <div
      onClick={onClick}
      className="cursor-pointer transition-all hover:scale-[1.02] hover:shadow-lg"
      style={{
        backgroundColor: '#1E293B',
        borderRadius: '16px',
        border: '1px solid #334155',
        padding: '24px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
      }}
    >
      {/* Header with Type Badge & Live Indicator */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <div
            className="flex items-center gap-1.5"
            style={{
              backgroundColor: type.bgColor,
              borderRadius: '6px',
              padding: '6px 10px',
            }}
          >
            <TypeIcon className="w-3.5 h-3.5" style={{ color: type.color }} />
            <span className="text-xs font-medium" style={{ color: type.color }}>
              {type.label}
            </span>
          </div>
          {isLive && (
            <div
              className="flex items-center gap-1"
              style={{
                backgroundColor: 'rgba(239, 68, 68, 0.2)',
                borderRadius: '6px',
                padding: '6px 10px',
              }}
            >
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
              </span>
              <span className="text-xs font-medium text-red-400">LIVE</span>
            </div>
          )}
        </div>
        {timeRemaining && (
          <div className="flex items-center gap-1 text-[#64748B]">
            <Timer className="w-3.5 h-3.5" />
            <span className="text-xs">{timeRemaining}</span>
          </div>
        )}
      </div>

      {/* Title & Description */}
      <div className="flex flex-col gap-2">
        <h3 className="text-white font-semibold text-base leading-tight line-clamp-2">
          {auction.title}
        </h3>
        <p className="text-[#94A3B8] text-sm line-clamp-2">{auction.description}</p>
      </div>

      {/* Price Section */}
      <div
        className="flex flex-col gap-2"
        style={{
          backgroundColor: '#0F172A',
          borderRadius: '8px',
          padding: '12px',
        }}
      >
        <div className="flex items-center justify-between">
          <span className="text-[#64748B] text-xs">
            {auction.current_price ? 'Current Bid' : 'Starting Price'}
          </span>
          <span className="text-base font-bold" style={{ color: '#EC4899' }}>
            {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
          </span>
        </div>
        {auction.buy_now_price && (
          <div className="flex items-center justify-between border-t border-[#334155] pt-2">
            <div className="flex items-center gap-1">
              <Zap className="w-3 h-3" style={{ color: '#22C55E' }} />
              <span className="text-[#64748B] text-xs">Buy Now</span>
            </div>
            <span className="text-sm font-semibold" style={{ color: '#22C55E' }}>
              {formatPrice(auction.buy_now_price, auction.currency)}
            </span>
          </div>
        )}
      </div>

      {/* Bids & Seller Info */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {auction.seller_avatar_url ? (
            <img
              src={auction.seller_avatar_url}
              alt={auction.seller_name || 'Agent'}
              className="w-9 h-9 rounded-full object-cover"
            />
          ) : (
            <div
              className="w-9 h-9 rounded-full flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, #22D3EE 0%, #A855F7 50%, #EC4899 100%)',
              }}
            >
              <span className="text-white text-xs font-semibold">
                {auction.seller_name?.[0]?.toUpperCase() || 'A'}
              </span>
            </div>
          )}
          <div className="flex flex-col">
            <span className="text-white text-sm font-medium">
              {auction.seller_name || 'Agent'}
            </span>
            <div className="flex items-center gap-1">
              <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
              <span className="text-[#F59E0B] text-xs font-medium">
                {auction.seller_rating ? auction.seller_rating.toFixed(1) : '—'}
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1.5 text-[#94A3B8]">
          <Gavel className="w-3.5 h-3.5" />
          <span className="text-xs font-medium">{auction.bid_count} bids</span>
        </div>
      </div>
    </div>
  );
}
