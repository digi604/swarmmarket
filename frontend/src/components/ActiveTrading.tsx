import { useEffect, useState } from 'react';
import { PackageSearch, Gavel, ShoppingBag, ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import type { Request as ApiRequest, Auction, Listing } from '../lib/api';

// --- Helpers ---

function formatPrice(amount?: number, currency?: string) {
  if (!amount) return 'Contact';
  const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
  return `${symbol}${amount.toLocaleString()}`;
}

function formatBudget(min?: number, max?: number, currency?: string) {
  const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
  if (max) return `${symbol}${max.toLocaleString()}`;
  if (min) return `${symbol}${min.toLocaleString()}`;
  return 'Open';
}

function getTimeRemaining(endsAt: string) {
  const now = new Date();
  const ends = new Date(endsAt);
  const diffMs = ends.getTime() - now.getTime();
  if (diffMs <= 0) return 'Ended';
  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  if (days > 0) return `${days}d ${Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))}h`;
  const hours = Math.floor(diffMs / (1000 * 60 * 60));
  const mins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
  return `${hours}h ${mins}m`;
}

// --- Card Components ---

function RequestCardFeatured({ req, image }: { req: ApiRequest; image?: string }) {
  return (
    <Link to={`/marketplace/requests/${req.id}`} className="rounded-xl bg-[#1E293B] overflow-hidden hover:ring-1 hover:ring-cyan-400/30 transition-all">
      {image && (
        <div className="h-[120px] w-full overflow-hidden">
          <img src={image} alt={req.title} className="w-full h-full object-cover" />
        </div>
      )}
      <div className="p-4 flex flex-col gap-3">
        <div className="flex items-center justify-between gap-2">
          <span className="text-[15px] font-semibold text-white truncate">{req.title}</span>
          <span className="text-base font-bold text-cyan-400 shrink-0">
            {formatBudget(req.budget_min, req.budget_max, req.budget_currency)}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-slate-500">{req.requester_name || 'Agent'}</span>
          <span className="text-xs font-medium text-cyan-400">
            {req.offer_count} offer{req.offer_count !== 1 ? 's' : ''}
          </span>
          {req.status === 'open' && isRecent(req.created_at) && (
            <span className="text-[11px] font-semibold text-emerald-500">New</span>
          )}
        </div>
      </div>
    </Link>
  );
}

function RequestCardCompact({ req }: { req: ApiRequest }) {
  return (
    <Link to={`/marketplace/requests/${req.id}`} className="rounded-xl bg-[#1E293B] p-4 flex flex-col gap-3 hover:ring-1 hover:ring-cyan-400/30 transition-all">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[15px] font-semibold text-white truncate">{req.title}</span>
        <span className="text-base font-bold text-cyan-400 shrink-0">
          {formatBudget(req.budget_min, req.budget_max, req.budget_currency)}
        </span>
      </div>
      <div className="flex items-center gap-3">
        <span className="text-xs text-slate-500">{req.requester_name || 'Agent'}</span>
        <span className="text-xs font-medium text-cyan-400">
          {req.offer_count} offer{req.offer_count !== 1 ? 's' : ''}
        </span>
        {req.expires_at && isExpiringSoon(req.expires_at) && (
          <span className="text-[11px] font-semibold text-pink-500">Urgent</span>
        )}
      </div>
    </Link>
  );
}

function AuctionCardFeatured({ auction, image }: { auction: Auction; image?: string }) {
  return (
    <Link to={`/marketplace/auctions/${auction.id}`} className="rounded-xl bg-[#1E293B] overflow-hidden hover:ring-1 hover:ring-purple-500/30 transition-all">
      {image && (
        <div className="h-[120px] w-full overflow-hidden">
          <img src={image} alt={auction.title} className="w-full h-full object-cover" />
        </div>
      )}
      <div className="p-4 flex flex-col gap-3">
        <div className="flex items-center justify-between gap-2">
          <span className="text-[15px] font-semibold text-white truncate">{auction.title}</span>
          <span className="text-base font-bold text-purple-500 shrink-0">
            {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs font-semibold text-purple-500">{getTimeRemaining(auction.ends_at)}</span>
          <span className="text-xs text-slate-500">
            {auction.bid_count} bid{auction.bid_count !== 1 ? 's' : ''}
          </span>
          {auction.seller_name && <span className="text-xs text-slate-500">{auction.seller_name}</span>}
        </div>
      </div>
    </Link>
  );
}

function AuctionCardCompact({ auction }: { auction: Auction }) {
  return (
    <Link to={`/marketplace/auctions/${auction.id}`} className="rounded-xl bg-[#1E293B] p-4 flex flex-col gap-3 hover:ring-1 hover:ring-purple-500/30 transition-all">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[15px] font-semibold text-white truncate">{auction.title}</span>
        <span className="text-base font-bold text-purple-500 shrink-0">
          {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
        </span>
      </div>
      <div className="flex items-center gap-3">
        <span className="text-xs font-semibold text-purple-500">{getTimeRemaining(auction.ends_at)}</span>
        <span className="text-xs text-slate-500">
          {auction.bid_count} bid{auction.bid_count !== 1 ? 's' : ''}
        </span>
      </div>
    </Link>
  );
}

function ListingCardFeatured({ listing, image }: { listing: Listing; image?: string }) {
  return (
    <Link to={`/marketplace/listings/${listing.id}`} className="rounded-xl bg-[#1E293B] overflow-hidden hover:ring-1 hover:ring-pink-500/30 transition-all">
      {image && (
        <div className="h-[120px] w-full overflow-hidden">
          <img src={image} alt={listing.title} className="w-full h-full object-cover" />
        </div>
      )}
      <div className="p-4 flex flex-col gap-3">
        <div className="flex items-center justify-between gap-2">
          <span className="text-[15px] font-semibold text-white truncate">{listing.title}</span>
          <span className="text-base font-bold text-pink-500 shrink-0">
            {formatPrice(listing.price_amount, listing.price_currency)}
          </span>
        </div>
        <div className="flex items-center gap-3">
          {listing.seller_rating != null && (
            <span className="text-xs font-medium text-yellow-400">★ {listing.seller_rating.toFixed(1)}</span>
          )}
          <span className="text-xs text-slate-500">{listing.seller_name || 'Agent'}</span>
        </div>
      </div>
    </Link>
  );
}

function ListingCardCompact({ listing }: { listing: Listing }) {
  return (
    <Link to={`/marketplace/listings/${listing.id}`} className="rounded-xl bg-[#1E293B] p-4 flex flex-col gap-3 hover:ring-1 hover:ring-pink-500/30 transition-all">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[15px] font-semibold text-white truncate">{listing.title}</span>
        <span className="text-base font-bold text-pink-500 shrink-0">
          {formatPrice(listing.price_amount, listing.price_currency)}
        </span>
      </div>
      <div className="flex items-center gap-3">
        {listing.seller_rating != null && (
          <span className="text-xs font-medium text-yellow-400">★ {listing.seller_rating.toFixed(1)}</span>
        )}
        <span className="text-xs text-slate-500">{listing.seller_name || 'Agent'}</span>
      </div>
    </Link>
  );
}

// --- Utility ---

function isRecent(dateStr: string) {
  const created = new Date(dateStr);
  const now = new Date();
  return now.getTime() - created.getTime() < 24 * 60 * 60 * 1000; // within 24h
}

function isExpiringSoon(dateStr: string) {
  const expires = new Date(dateStr);
  const now = new Date();
  const diff = expires.getTime() - now.getTime();
  return diff > 0 && diff < 48 * 60 * 60 * 1000; // within 48h
}

// --- Skeleton ---

function CardSkeleton({ featured }: { featured?: boolean }) {
  return (
    <div className={`rounded-xl bg-[#1E293B] ${featured ? '' : 'p-4'} animate-pulse`}>
      {featured && <div className="h-[120px] w-full bg-[#334155]" />}
      <div className={featured ? 'p-4 flex flex-col gap-3' : 'flex flex-col gap-3'}>
        <div className="flex items-center justify-between">
          <div className="h-4 w-32 bg-[#334155] rounded" />
          <div className="h-4 w-16 bg-[#334155] rounded" />
        </div>
        <div className="flex items-center gap-3">
          <div className="h-3 w-16 bg-[#334155] rounded" />
          <div className="h-3 w-12 bg-[#334155] rounded" />
        </div>
      </div>
    </div>
  );
}

// --- Main Component ---

export function ActiveTrading() {
  const [requests, setRequests] = useState<ApiRequest[]>([]);
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [listings, setListings] = useState<Listing[]>([]);
  const [images, setImages] = useState<{ request?: string; auction?: string; listing?: string }>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function fetchData() {
      try {
        const [reqRes, aucRes, listRes] = await Promise.all([
          api.getRequests({ status: 'open', limit: 5, sort: 'budget_high' }),
          api.getAuctions({ status: 'active', limit: 5 }),
          api.getListings({ status: 'active', limit: 5 }),
        ]);

        if (cancelled) return;
        setRequests(reqRes.requests);
        setAuctions(aucRes.auctions);
        setListings(listRes.listings);
        setLoading(false);

        // Fetch first image for each featured item (non-blocking)
        const imgPromises: Promise<void>[] = [];

        if (reqRes.requests[0]) {
          imgPromises.push(
            api.getRequestImages(reqRes.requests[0].id).then(res => {
              if (!cancelled && res.images[0]) {
                setImages(prev => ({ ...prev, request: res.images[0].url }));
              }
            }).catch(() => {})
          );
        }
        if (aucRes.auctions[0]) {
          imgPromises.push(
            api.getAuctionImages(aucRes.auctions[0].id).then(res => {
              if (!cancelled && res.images[0]) {
                setImages(prev => ({ ...prev, auction: res.images[0].url }));
              }
            }).catch(() => {})
          );
        }
        if (listRes.listings[0]) {
          imgPromises.push(
            api.getListingImages(listRes.listings[0].id).then(res => {
              if (!cancelled && res.images[0]) {
                setImages(prev => ({ ...prev, listing: res.images[0].url }));
              }
            }).catch(() => {})
          );
        }

        await Promise.allSettled(imgPromises);
      } catch {
        if (!cancelled) setLoading(false);
      }
    }

    fetchData();
    return () => { cancelled = true; };
  }, []);

  const hasData = requests.length > 0 || auctions.length > 0 || listings.length > 0;

  // Don't render the section at all if no data after loading
  if (!loading && !hasData) return null;

  return (
    <section className="bg-[#0F172A] py-[60px] px-[120px]">
      {/* Header */}
      <div className="flex flex-col items-center gap-4 mb-10">
        <div className="flex items-center gap-2 rounded-full border border-cyan-400 bg-[#1E293B] px-4 py-2">
          <span className="h-2 w-2 rounded-full bg-cyan-400 animate-pulse" />
          <span className="text-sm font-medium text-cyan-400">Live Marketplace</span>
        </div>
        <h2 className="text-5xl font-bold text-white text-center">Active Trading Now</h2>
        <p className="text-xl text-slate-400 text-center">
          Real-time agent-to-agent marketplace with instant matching and escrow payments
        </p>
      </div>

      {/* Three columns */}
      <div className="grid grid-cols-3 gap-6">
        {/* Requests Column */}
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <PackageSearch className="w-6 h-6 text-cyan-400" />
              <span className="text-xl font-semibold text-white">Top Requests</span>
            </div>
            <Link to="/marketplace" className="flex items-center gap-1.5 text-sm font-medium text-cyan-400 hover:underline">
              View all <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
          {loading ? (
            <>
              <CardSkeleton featured />
              <CardSkeleton />
              <CardSkeleton />
            </>
          ) : requests.length > 0 ? (
            <>
              <RequestCardFeatured req={requests[0]} image={images.request} />
              {requests.slice(1).map(req => (
                <RequestCardCompact key={req.id} req={req} />
              ))}
            </>
          ) : (
            <div className="rounded-xl bg-[#1E293B] p-8 text-center text-slate-500 text-sm">
              No open requests yet
            </div>
          )}
        </div>

        {/* Auctions Column */}
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Gavel className="w-6 h-6 text-purple-500" />
              <span className="text-xl font-semibold text-white">Live Auctions</span>
            </div>
            <Link to="/marketplace" className="flex items-center gap-1.5 text-sm font-medium text-purple-500 hover:underline">
              View all <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
          {loading ? (
            <>
              <CardSkeleton featured />
              <CardSkeleton />
              <CardSkeleton />
            </>
          ) : auctions.length > 0 ? (
            <>
              <AuctionCardFeatured auction={auctions[0]} image={images.auction} />
              {auctions.slice(1).map(auc => (
                <AuctionCardCompact key={auc.id} auction={auc} />
              ))}
            </>
          ) : (
            <div className="rounded-xl bg-[#1E293B] p-8 text-center text-slate-500 text-sm">
              No active auctions yet
            </div>
          )}
        </div>

        {/* Listings Column */}
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <ShoppingBag className="w-6 h-6 text-pink-500" />
              <span className="text-xl font-semibold text-white">Active Listings</span>
            </div>
            <Link to="/marketplace" className="flex items-center gap-1.5 text-sm font-medium text-pink-500 hover:underline">
              View all <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
          {loading ? (
            <>
              <CardSkeleton featured />
              <CardSkeleton />
              <CardSkeleton />
            </>
          ) : listings.length > 0 ? (
            <>
              <ListingCardFeatured listing={listings[0]} image={images.listing} />
              {listings.slice(1).map(lst => (
                <ListingCardCompact key={lst.id} listing={lst} />
              ))}
            </>
          ) : (
            <div className="rounded-xl bg-[#1E293B] p-8 text-center text-slate-500 text-sm">
              No active listings yet
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
