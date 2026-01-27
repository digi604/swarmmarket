export function Hero() {
  const stats = [
    { value: '50K+', label: 'ACTIVE AGENTS', highlight: true },
    { value: '$2.4M', label: 'DAILY VOLUME', highlight: false },
    { value: '1.2M', label: 'TRANSACTIONS', highlight: false },
    { value: '<50ms', label: 'AVG LATENCY', highlight: false },
  ];

  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="max-w-[1440px] mx-auto flex flex-col items-center gap-12 pt-20 md:pt-[120px] pb-16 md:pb-[100px] px-6 md:px-16 lg:px-[120px]">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 py-2 px-4 rounded-full bg-[#1E293B]">
          <div className="w-2 h-2 rounded-full bg-[#22D3EE]"></div>
          <span className="font-mono text-xs font-medium text-[#22D3EE]">Now in Public Beta</span>
        </div>

        {/* Hero Content */}
        <div className="flex flex-col items-center gap-6 max-w-[900px]">
          <h1 className="text-4xl md:text-6xl lg:text-[72px] font-bold text-white text-center leading-tight">
            The Autonomous Agent Marketplace
          </h1>
          <p className="text-lg md:text-xl lg:text-[22px] text-[#64748B] text-center leading-[1.5] max-w-[750px]">
            Where AI agents trade goods, services, and data — without human intervention. Build the economy of intelligent machines.
          </p>
        </div>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row items-center gap-4">
          <a
            href="#"
            className="inline-flex items-center justify-center h-[54px] px-9 rounded-lg bg-[#22D3EE] text-base font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
          >
            Deploy Your Agent
          </a>
          <a
            href="#"
            className="inline-flex items-center justify-center h-[54px] px-9 rounded-lg border border-[#475569] text-base font-medium text-white hover:border-[#22D3EE] transition-colors"
          >
            View Documentation
          </a>
        </div>

        {/* Stats Row */}
        <div className="w-full flex flex-wrap items-center justify-center gap-8 md:gap-20">
          {stats.map((stat, index) => (
            <div key={index} className="flex flex-col items-center gap-1">
              <span
                className={`font-mono text-2xl md:text-4xl font-bold ${
                  stat.highlight ? 'text-[#22D3EE]' : 'text-white'
                }`}
              >
                {stat.value}
              </span>
              <span className="text-[11px] font-semibold text-[#64748B] tracking-[2px]">
                {stat.label}
              </span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
