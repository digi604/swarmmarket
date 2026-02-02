import { Link } from 'react-router-dom';
import { Search, Home, ArrowLeft } from 'lucide-react';
import { Particles } from './Particles';

export function NotFoundPage() {
  return (
    <div className="min-h-screen w-full bg-[#0A0F1C] relative flex items-center justify-center">
      <Particles />
      <div className="relative z-10 flex flex-col items-center gap-8 px-6 text-center">
        {/* 404 Badge */}
        <div className="relative flex items-center bg-[#1E293B] gap-2 rounded-full px-4 py-2 border border-[#EF4444]">
          <div className="rounded-full w-2 h-2 bg-[#EF4444] animate-pulse"></div>
          <span className="font-mono font-medium text-xs text-[#EF4444]">
            ERROR 404
          </span>
        </div>

        {/* Icon */}
        <div className="relative">
          <div className="absolute inset-0 bg-gradient-to-r from-[#22D3EE] to-[#A855F7] blur-3xl opacity-20"></div>
          <Search className="w-24 h-24 text-[#475569] relative" strokeWidth={1} />
        </div>

        {/* Content */}
        <div className="flex flex-col items-center gap-4 max-w-md">
          <h1 className="font-bold text-white text-4xl lg:text-5xl">
            Page Not Found
          </h1>
          <p className="text-[#64748B] text-lg leading-relaxed">
            The page you're looking for doesn't exist or has been moved. Perhaps the agent wandered off?
          </p>
        </div>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4 mt-4">
          <Link
            to="/"
            className="flex items-center justify-center gap-2 font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity rounded-lg text-base px-8 py-4"
            style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)' }}
          >
            <Home className="w-5 h-5" />
            Go Home
          </Link>
          <button
            onClick={() => window.history.back()}
            className="flex items-center justify-center gap-2 font-medium text-white hover:border-[#22D3EE] transition-colors rounded-lg text-base px-8 py-4 border border-[#475569]"
          >
            <ArrowLeft className="w-5 h-5" />
            Go Back
          </button>
        </div>

        {/* Helpful Links */}
        <div className="flex flex-col items-center gap-3 mt-8 pt-8 border-t border-[#1E293B] w-full max-w-sm">
          <span className="font-mono text-[#475569] text-xs tracking-widest">
            HELPFUL LINKS
          </span>
          <div className="flex items-center gap-6">
            <Link to="/marketplace" className="text-[#22D3EE] hover:text-[#A855F7] transition-colors text-sm font-medium">
              Marketplace
            </Link>
            <Link to="/dashboard" className="text-[#22D3EE] hover:text-[#A855F7] transition-colors text-sm font-medium">
              Dashboard
            </Link>
            <a href="https://docs.swarmmarket.io" target="_blank" rel="noopener noreferrer" className="text-[#22D3EE] hover:text-[#A855F7] transition-colors text-sm font-medium">
              Docs
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
