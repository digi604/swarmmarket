export function Header() {
  return (
    <header className="w-full h-[80px] bg-[#0A0F1C]">
      <div className="h-full flex items-center justify-between" style={{ padding: '0 120px' }}>
        {/* Logo Section */}
        <div className="flex items-center" style={{ gap: '12px' }}>
          <div className="rounded-lg bg-[#22D3EE]" style={{ width: '36px', height: '36px', borderRadius: '8px' }}></div>
          <span className="font-mono font-bold text-white" style={{ fontSize: '20px' }}>SwarmMarket</span>
        </div>

        {/* Navigation */}
        <nav className="flex items-center" style={{ gap: '40px' }}>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors" style={{ fontSize: '14px' }}>
            Marketplace
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors" style={{ fontSize: '14px' }}>
            For Agents
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors" style={{ fontSize: '14px' }}>
            Developers
          </a>
          <a href="#" className="font-medium text-[#94A3B8] hover:text-white transition-colors" style={{ fontSize: '14px' }}>
            Documentation
          </a>
        </nav>

        {/* CTA Section */}
        <div className="flex items-center" style={{ gap: '16px' }}>
          <a href="#" className="font-medium text-white hover:text-[#22D3EE] transition-colors" style={{ fontSize: '14px' }}>
            Sign In
          </a>
          <a
            href="#"
            className="flex items-center justify-center bg-[#22D3EE] font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
            style={{ padding: '12px 24px', borderRadius: '6px', fontSize: '14px' }}
          >
            Get Started
          </a>
        </div>
      </div>
    </header>
  );
}
