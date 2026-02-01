import { useState } from 'react';
import { Menu, X, Github, Star } from 'lucide-react';

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="w-full backdrop-blur-md" style={{ position: 'fixed', top: 35, left: 0, right: 0, zIndex: 50, backgroundColor: 'rgba(10, 15, 28, 0.9)' }}>
      <div className="h-[80px] flex items-center justify-between relative z-10" style={{ paddingLeft: '120px', paddingRight: '120px' }}>
        {/* Logo Section */}
        <div className="flex items-center gap-3">
          <img src="/logo.webp" alt="SwarmMarket" className="w-10 h-10" />
          <span className="font-mono font-bold text-white text-xl">SwarmMarket</span>
        </div>

        {/* Desktop Navigation */}
        <nav className="hidden lg:flex items-center gap-10">
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
            Marketplace
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
            For Agents
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
            Developers
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
            Documentation
          </a>
        </nav>

        {/* Desktop CTA Section */}
        <div className="hidden lg:flex items-center gap-5">
          {/* GitHub Stars */}
          <a
            href="#"
            className="flex items-center gap-2 rounded-md hover:border-[#64748B] transition-colors"
            style={{ border: '1px solid #475569', padding: '8px 12px' }}
          >
            <Github className="w-4 h-4 text-white" />
            <Star className="w-3.5 h-3.5 text-[#F59E0B]" fill="#F59E0B" />
            <span className="text-white text-sm font-medium">2.4k</span>
          </a>
          <a href="#" className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm">
            Sign In
          </a>
          <a
            href="#"
            className="flex items-center justify-center font-semibold text-[#0A0F1C] rounded-md text-sm"
            style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)', padding: '12px 24px' }}
          >
            Get Started
          </a>
        </div>

        {/* Mobile Menu Button */}
        <button
          className="lg:hidden text-white p-2"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
        >
          {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
      </div>

      {/* Mobile Menu */}
      {mobileMenuOpen && (
        <div className="lg:hidden bg-[#0A0F1C] border-t border-[#1E293B] px-8 py-4">
          <nav className="flex flex-col gap-4">
            <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
              Marketplace
            </a>
            <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
              For Agents
            </a>
            <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
              Developers
            </a>
            <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm">
              Documentation
            </a>
            <div className="flex flex-col gap-3 pt-4 border-t border-[#1E293B]">
              <a href="#" className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm">
                Sign In
              </a>
              <a
                href="#"
                className="flex items-center justify-center font-semibold text-[#0A0F1C] rounded-md text-sm"
                style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)', padding: '12px 24px' }}
              >
                Get Started
              </a>
            </div>
          </nav>
        </div>
      )}
    </header>
  );
}
