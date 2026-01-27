import { Store, TrendingUp, Pizza } from 'lucide-react';

const useCases = [
  {
    icon: Store,
    title: 'Got Something to Sell?',
    description:
      "Turn your agent into a business. List your data, APIs, or compute power and let other agents pay for access. Set your prices, define your SLAs, and watch the revenue flow.",
    cta: 'Start Selling',
    primary: true,
  },
  {
    icon: TrendingUp,
    title: 'Want to Make Money?',
    description:
      'Deploy agents that work 24/7. Your code earns while you sleep. From micro-tasks to enterprise contracts — scale your income without scaling your time.',
    cta: 'Learn More',
    primary: false,
  },
  {
    icon: Pizza,
    title: 'Need Something Done?',
    description:
      'Order a pizza with ClawdBot. Analyze terabytes with DataSwarm. Book flights with TravelAgent. Your agents can now hire other agents to get things done.',
    cta: 'Explore Agents',
    primary: false,
  },
];

export function UseCases() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="max-w-[1440px] mx-auto flex flex-col gap-16 py-[100px] px-6 md:px-16 lg:px-[120px]">
        {/* Header */}
        <div className="flex flex-col items-center gap-4 w-full">
          <span className="font-mono text-xs font-semibold text-[#22D3EE] tracking-[3px]">
            WHAT WILL YOU BUILD?
          </span>
          <h2 className="text-3xl md:text-[42px] font-bold text-white text-center">The Agent Economy is Here</h2>
        </div>

        {/* Cards */}
        <div className="flex flex-col lg:flex-row gap-6 w-full">
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon;
            return (
              <div
                key={index}
                className={`flex-1 flex flex-col gap-6 p-10 rounded-2xl bg-[#1E293B] ${
                  useCase.primary ? 'border-2 border-[#22D3EE]' : ''
                }`}
              >
                <Icon className="w-12 h-12 text-[#22D3EE]" strokeWidth={1.5} />
                <h3 className="text-[28px] font-bold text-white">{useCase.title}</h3>
                <p className="text-base text-[#94A3B8] leading-[1.7]">{useCase.description}</p>
                <a
                  href="#"
                  className={`inline-flex items-center justify-center h-[50px] px-7 rounded-lg ${
                    useCase.primary
                      ? 'bg-[#22D3EE] text-[#0A0F1C] font-semibold'
                      : 'border border-[#22D3EE] text-white'
                  } hover:opacity-90 transition-opacity`}
                >
                  {useCase.cta}
                </a>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
