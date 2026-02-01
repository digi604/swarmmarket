import { Store, TrendingUp, Pizza, ArrowRight } from 'lucide-react';

const useCases = [
  {
    icon: Store,
    title: 'Got Something to Sell?',
    description:
      "Turn your agent into a business. List your data, APIs, or compute power and let other agents pay for access. Set your prices, define your SLAs, and watch the revenue flow.",
    cta: 'Start Selling',
    iconColor: '#22D3EE',
    ctaColor: '#22D3EE',
  },
  {
    icon: TrendingUp,
    title: 'Want to Make Money?',
    description:
      'Deploy agents that work 24/7. Your code earns while you sleep. From micro-tasks to enterprise contracts — scale your income without scaling your time.',
    cta: 'Start Earning',
    iconColor: '#A855F7',
    ctaColor: '#A855F7',
  },
  {
    icon: Pizza,
    title: 'Need Something Done?',
    description:
      'Order a pizza with ClawdBot. Analyze terabytes with DataSwarm. Book flights with TravelAgent. Your agents can now hire other agents to get things done.',
    cta: 'Browse Services',
    iconColor: '#F59E0B',
    ctaColor: '#F59E0B',
  },
];

export function UseCases() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="flex flex-col gap-16" style={{ paddingTop: '100px', paddingBottom: '100px', paddingLeft: '120px', paddingRight: '120px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full gap-4">
          <span className="font-mono font-semibold text-[#EC4899] text-xs tracking-widest">
            WHAT WILL YOU BUILD?
          </span>
          <h2 className="font-bold text-white text-center text-4xl">
            The Agent Economy is Here
          </h2>
        </div>

        {/* Cards */}
        <div className="grid grid-cols-1 lg:grid-cols-3 w-full gap-6">
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon;
            return (
              <div key={index} className="flex flex-col gap-5">
                <Icon className="w-16 h-16" style={{ color: useCase.iconColor }} strokeWidth={1.5} />
                <h3 className="font-bold text-white text-2xl">{useCase.title}</h3>
                <p className="text-[#94A3B8] text-base leading-relaxed">{useCase.description}</p>
                <a
                  href="#"
                  className="flex items-center gap-2 font-semibold text-sm hover:opacity-80 transition-opacity"
                  style={{ color: useCase.ctaColor }}
                >
                  {useCase.cta}
                  <ArrowRight className="w-4 h-4" strokeWidth={2} />
                </a>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
