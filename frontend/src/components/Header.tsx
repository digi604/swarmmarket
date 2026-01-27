export function Header() {
  return (
    <header className="w-full h-20 bg-[#0A0F1C]">
      <div className="max-w-[1440px] mx-auto h-full flex items-center justify-between px-6 md:px-16 lg:px-[120px]">
        {/* Logo Section */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-[#22D3EE]"></div>
          <span className="font-mono font-bold text-xl text-white">SwarmMarket</span>
        </div>

        {/* Navigation */}
        <nav className="hidden md:flex items-center gap-10">
          <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
            Marketplace
          </a>
          <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
            For Agents
          </a>
          <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
            Developers
          </a>
          <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
            Documentation
          </a>
        </nav>

        {/* CTA Section */}
        <div className="flex items-center gap-4">
          <a href="#" className="hidden sm:block text-sm font-medium text-white hover:text-[#22D3EE] transition-colors">
            Sign In
          </a>
          <a
            href="#"
            className="inline-flex items-center justify-center h-[42px] px-6 rounded-[6px] bg-[#22D3EE] text-sm font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
          >
            Get Started
          </a>
        </div>
      </div>
    </header>
  );
}
