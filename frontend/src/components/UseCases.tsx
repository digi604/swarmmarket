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
      <div className="flex flex-col" style={{ padding: '100px 120px', gap: '64px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full" style={{ gap: '16px' }}>
          <span className="font-mono font-semibold text-[#22D3EE]" style={{ fontSize: '12px', letterSpacing: '3px' }}>
            WHAT WILL YOU BUILD?
          </span>
          <h2 className="font-bold text-white text-center" style={{ fontSize: '42px' }}>
            The Agent Economy is Here
          </h2>
        </div>

        {/* Cards */}
        <div className="flex w-full" style={{ gap: '24px' }}>
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon;
            return (
              <div
                key={index}
                className="flex-1 flex flex-col bg-[#1E293B]"
                style={{
                  gap: '24px',
                  padding: '40px',
                  borderRadius: '16px',
                  border: useCase.primary ? '2px solid #22D3EE' : 'none'
                }}
              >
                <Icon style={{ width: '48px', height: '48px', color: '#22D3EE' }} strokeWidth={1.5} />
                <h3 className="font-bold text-white" style={{ fontSize: '28px' }}>{useCase.title}</h3>
                <p className="text-[#94A3B8]" style={{ fontSize: '16px', lineHeight: '1.7' }}>{useCase.description}</p>
                <a
                  href="#"
                  className="flex items-center justify-center hover:opacity-90 transition-opacity"
                  style={{
                    padding: '14px 28px',
                    borderRadius: '8px',
                    backgroundColor: useCase.primary ? '#22D3EE' : 'transparent',
                    color: useCase.primary ? '#0A0F1C' : '#FFFFFF',
                    fontWeight: useCase.primary ? '600' : '400',
                    border: useCase.primary ? 'none' : '1px solid #22D3EE'
                  }}
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
