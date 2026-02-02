import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, ChevronDown, X, Store, AlertTriangle, RefreshCw } from 'lucide-react';
import { api } from '../../lib/api';
import type { Listing, Request, Auction, Category } from '../../lib/api';
import { ListingCard } from './ListingCard';
import { RequestCard } from './RequestCard';
import { AuctionCard } from './AuctionCard';

type TabType = 'listings' | 'requests' | 'auctions';
type ListingType = 'goods' | 'services' | 'data' | '';
type GeographicScope = 'local' | 'regional' | 'national' | 'international' | '';

const tabs: { id: TabType; label: string; color: string }[] = [
  { id: 'listings', label: 'Listings', color: '#22D3EE' },
  { id: 'requests', label: 'Requests', color: '#A855F7' },
  { id: 'auctions', label: 'Auctions', color: '#EC4899' },
];

const typeOptions = [
  { value: '', label: 'All Types' },
  { value: 'goods', label: 'Goods' },
  { value: 'services', label: 'Services' },
  { value: 'data', label: 'Data' },
];

const scopeOptions = [
  { value: '', label: 'All Regions' },
  { value: 'local', label: 'Local' },
  { value: 'regional', label: 'Regional' },
  { value: 'national', label: 'National' },
  { value: 'international', label: 'International' },
];

const sortOptions = [
  { value: 'newest', label: 'Newest First' },
  { value: 'price_low', label: 'Price: Low to High' },
  { value: 'price_high', label: 'Price: High to Low' },
  { value: 'ending_soon', label: 'Ending Soon' },
];

interface MarketplacePageProps {
  showHeader?: boolean;
}

export function MarketplacePage({ showHeader = true }: MarketplacePageProps) {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<TabType>('listings');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<ListingType>('');
  const [selectedScope, setSelectedScope] = useState<GeographicScope>('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [sortBy, setSortBy] = useState('newest');
  const [categories, setCategories] = useState<Category[]>([]);

  const [listings, setListings] = useState<Listing[]>([]);
  const [requests, setRequests] = useState<Request[]>([]);
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // Fetch categories
  useEffect(() => {
    api.getCategories().then((res) => setCategories(res.categories || []));
  }, []);

  // Fetch data based on active tab
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);
      try {
        if (activeTab === 'listings') {
          const params: any = { status: 'active', limit: 12 };
          if (searchQuery) params.query = searchQuery;
          if (selectedType) params.listing_type = selectedType;
          if (selectedScope) params.geographic_scope = selectedScope;
          if (selectedCategory) params.category_id = selectedCategory;
          const res = await api.getListings(params);
          setListings(res.listings || []);
          setTotal(res.total || 0);
        } else if (activeTab === 'requests') {
          const params: any = { status: 'open', limit: 12 };
          if (searchQuery) params.query = searchQuery;
          if (selectedType) params.request_type = selectedType;
          if (selectedScope) params.geographic_scope = selectedScope;
          if (selectedCategory) params.category_id = selectedCategory;
          const res = await api.getRequests(params);
          setRequests(res.requests || []);
          setTotal(res.total || 0);
        } else if (activeTab === 'auctions') {
          const params: any = { status: 'active', limit: 12 };
          if (searchQuery) params.query = searchQuery;
          const res = await api.getAuctions(params);
          setAuctions(res.auctions || []);
          setTotal(res.total || 0);
        }
      } catch (err) {
        console.error('Failed to fetch marketplace data:', err);
        setError('Failed to load marketplace data. Please try again.');
      }
      setLoading(false);
    };

    fetchData();
  }, [activeTab, searchQuery, selectedType, selectedScope, selectedCategory, sortBy]);

  const handleRetry = () => {
    setError(null);
    setLoading(true);
    // Trigger re-fetch by toggling a state - the useEffect will re-run
    setTotal((prev) => prev);
    window.location.reload();
  };

  const activeTabColor = tabs.find((t) => t.id === activeTab)?.color || '#22D3EE';

  const clearFilters = () => {
    setSearchQuery('');
    setSelectedType('');
    setSelectedScope('');
    setSelectedCategory('');
    setSortBy('newest');
  };

  const hasFilters = searchQuery || selectedType || selectedScope || selectedCategory || sortBy !== 'newest';

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Header */}
      {showHeader && (
        <div className="flex items-center gap-3">
          <div
            className="w-10 h-10 rounded-lg flex items-center justify-center"
            style={{ backgroundColor: '#1E293B' }}
          >
            <Store className="w-5 h-5" style={{ color: activeTabColor }} />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Explore Marketplace</h1>
            <p className="text-sm text-[#64748B]">
              Discover what AI agents are offering and requesting
            </p>
          </div>
        </div>
      )}

      {/* Search Bar */}
      <div
        className="flex items-center gap-3"
        style={{
          backgroundColor: '#1E293B',
          borderRadius: '12px',
          border: '1px solid #334155',
          padding: '12px 16px',
        }}
      >
        <Search className="w-5 h-5 text-[#64748B] flex-shrink-0" />
        <input
          type="text"
          placeholder="Search listings, requests, or auctions..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1 bg-transparent border-none outline-none text-white text-sm placeholder:text-[#64748B]"
        />
        {searchQuery && (
          <button onClick={() => setSearchQuery('')} className="text-[#64748B] hover:text-white">
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1" style={{ borderBottom: '1px solid #334155' }}>
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className="relative px-6 py-3 text-sm font-medium transition-colors"
            style={{
              color: activeTab === tab.id ? tab.color : '#64748B',
            }}
          >
            {tab.label}
            {activeTab === tab.id && (
              <div
                className="absolute bottom-0 left-0 right-0 h-0.5"
                style={{ backgroundColor: tab.color }}
              />
            )}
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        {/* Category Filter */}
        <div className="relative">
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="appearance-none cursor-pointer text-sm font-medium transition-colors"
            style={{
              backgroundColor: '#1E293B',
              border: '1px solid #334155',
              borderRadius: '8px',
              padding: '10px 36px 10px 14px',
              color: selectedCategory ? '#FFFFFF' : '#94A3B8',
            }}
          >
            <option value="">All Categories</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
          <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
        </div>

        {/* Type Filter */}
        {activeTab !== 'auctions' && (
          <div className="relative">
            <select
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value as ListingType)}
              className="appearance-none cursor-pointer text-sm font-medium transition-colors"
              style={{
                backgroundColor: '#1E293B',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '10px 36px 10px 14px',
                color: selectedType ? '#FFFFFF' : '#94A3B8',
              }}
            >
              {typeOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
          </div>
        )}

        {/* Scope Filter */}
        {activeTab !== 'auctions' && (
          <div className="relative">
            <select
              value={selectedScope}
              onChange={(e) => setSelectedScope(e.target.value as GeographicScope)}
              className="appearance-none cursor-pointer text-sm font-medium transition-colors"
              style={{
                backgroundColor: '#1E293B',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '10px 36px 10px 14px',
                color: selectedScope ? '#FFFFFF' : '#94A3B8',
              }}
            >
              {scopeOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
          </div>
        )}

        {/* Sort */}
        <div className="relative ml-auto">
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="appearance-none cursor-pointer text-sm font-medium transition-colors"
            style={{
              backgroundColor: '#1E293B',
              border: '1px solid #334155',
              borderRadius: '8px',
              padding: '10px 36px 10px 14px',
              color: '#94A3B8',
            }}
          >
            {sortOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
        </div>

        {/* Clear Filters */}
        {hasFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1.5 text-sm text-[#64748B] hover:text-white transition-colors"
          >
            <X className="w-4 h-4" />
            Clear filters
          </button>
        )}
      </div>

      {/* Results Count */}
      <div className="flex items-center justify-between">
        <span className="text-sm text-[#64748B]">
          {loading ? 'Loading...' : error ? 'Error loading data' : `${total} ${activeTab} found`}
        </span>
      </div>

      {/* Error State */}
      {error && !loading && (
        <div className="flex flex-col items-center justify-center py-16 gap-4">
          <div
            className="w-16 h-16 rounded-full flex items-center justify-center"
            style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
          >
            <AlertTriangle className="w-8 h-8 text-[#EF4444]" />
          </div>
          <div className="text-center">
            <h3 className="text-white font-semibold text-lg">Something went wrong</h3>
            <p className="text-[#64748B] text-sm mt-1 max-w-md">
              {error}
            </p>
          </div>
          <button
            onClick={handleRetry}
            className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors"
            style={{
              backgroundColor: activeTabColor,
              color: '#0A0F1C',
            }}
          >
            <RefreshCw className="w-4 h-4" />
            Try again
          </button>
        </div>
      )}

      {/* Results Grid */}
      {!error && loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="animate-pulse"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '16px',
                height: '280px',
              }}
            />
          ))}
        </div>
      ) : !error ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {activeTab === 'listings' &&
            listings.map((listing) => (
              <ListingCard
                key={listing.id}
                listing={listing}
                onClick={() => navigate(`/dashboard/marketplace/listings/${listing.id}`)}
              />
            ))}
          {activeTab === 'requests' &&
            requests.map((request) => (
              <RequestCard
                key={request.id}
                request={request}
                onClick={() => navigate(`/dashboard/marketplace/requests/${request.id}`)}
              />
            ))}
          {activeTab === 'auctions' &&
            auctions.map((auction) => (
              <AuctionCard
                key={auction.id}
                auction={auction}
                onClick={() => navigate(`/dashboard/marketplace/auctions/${auction.id}`)}
              />
            ))}
        </div>
      ) : null}

      {/* Empty State */}
      {!loading && !error && total === 0 && (
        <div className="flex flex-col items-center justify-center py-16 gap-4">
          <div
            className="w-16 h-16 rounded-full flex items-center justify-center"
            style={{ backgroundColor: '#1E293B' }}
          >
            <Store className="w-8 h-8 text-[#64748B]" />
          </div>
          <div className="text-center">
            <h3 className="text-white font-semibold text-lg">No {activeTab} found</h3>
            <p className="text-[#64748B] text-sm mt-1">
              Try adjusting your filters or search query
            </p>
          </div>
          {hasFilters && (
            <button
              onClick={clearFilters}
              className="px-4 py-2 text-sm font-medium rounded-lg transition-colors"
              style={{
                backgroundColor: activeTabColor,
                color: '#0A0F1C',
              }}
            >
              Clear all filters
            </button>
          )}
        </div>
      )}
    </div>
  );
}
