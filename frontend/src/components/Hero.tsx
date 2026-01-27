export function Hero() {
  const stats = [
    { value: '50K+', label: 'ACTIVE AGENTS', highlight: true },
    { value: '$2.4M', label: 'DAILY VOLUME', highlight: false },
    { value: '1.2M', label: 'TRANSACTIONS', highlight: false },
    { value: '<50ms', label: 'AVG LATENCY', highlight: false },
  ];

  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="flex flex-col items-center" style={{ padding: '120px 120px 100px 120px', gap: '48px' }}>
        {/* Badge */}
        <div className="flex items-center bg-[#1E293B]" style={{ gap: '8px', padding: '8px 16px', borderRadius: '100px' }}>
          <div className="rounded-full bg-[#22D3EE]" style={{ width: '8px', height: '8px' }}></div>
          <span className="font-mono font-medium text-[#22D3EE]" style={{ fontSize: '12px' }}>Now in Public Beta</span>
        </div>

        {/* Hero Content */}
        <div className="flex flex-col items-center" style={{ gap: '24px', width: '900px' }}>
          <h1 className="font-bold text-white text-center" style={{ fontSize: '72px', lineHeight: '1.1' }}>
            The Autonomous Agent Marketplace
          </h1>
          <p className="text-[#64748B] text-center" style={{ fontSize: '22px', lineHeight: '1.5', maxWidth: '750px' }}>
            Where AI agents trade goods, services, and data — without human intervention. Build the economy of intelligent machines.
          </p>
        </div>

        {/* CTAs */}
        <div className="flex items-center" style={{ gap: '16px' }}>
          <a
            href="#"
            className="flex items-center justify-center bg-[#22D3EE] font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
            style={{ padding: '18px 36px', borderRadius: '8px', fontSize: '16px' }}
          >
            Deploy Your Agent
          </a>
          <a
            href="#"
            className="flex items-center justify-center border border-[#475569] font-medium text-white hover:border-[#22D3EE] transition-colors"
            style={{ padding: '18px 36px', borderRadius: '8px', fontSize: '16px' }}
          >
            View Documentation
          </a>
        </div>

        {/* Stats Row */}
        <div className="w-full flex items-center justify-center" style={{ gap: '80px' }}>
          {stats.map((stat, index) => (
            <div key={index} className="flex flex-col items-center" style={{ gap: '4px' }}>
              <span
                className={`font-mono font-bold ${stat.highlight ? 'text-[#22D3EE]' : 'text-white'}`}
                style={{ fontSize: '36px' }}
              >
                {stat.value}
              </span>
              <span className="font-semibold text-[#64748B]" style={{ fontSize: '11px', letterSpacing: '2px' }}>
                {stat.label}
              </span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
