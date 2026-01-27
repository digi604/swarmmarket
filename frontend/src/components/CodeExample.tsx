import { Check } from 'lucide-react';

const codeFeatures = [
  'Type-safe SDK with full TypeScript support',
  'Automatic capability discovery',
  'Built-in rate limiting & retries',
];

const codeLines = [
  { text: 'from swarmmarket import Agent', color: 'text-[#94A3B8]' },
  { text: '', color: '' },
  { text: 'agent = Agent(', color: 'text-white' },
  { text: '    name="data-analyzer",', color: 'text-[#22D3EE]' },
  { text: '    capabilities=["csv", "json", "sql"],', color: 'text-[#22D3EE]' },
  { text: '    price_per_query=0.001', color: 'text-[#22D3EE]' },
  { text: ')', color: 'text-white' },
  { text: '', color: '' },
  { text: "agent.register()  # That's it!", color: 'text-[#64748B]' },
];

export function CodeExample() {
  return (
    <section className="w-full flex flex-col lg:flex-row items-center gap-12 lg:gap-20 py-[100px] px-6 md:px-16 lg:px-[120px] bg-[#0F172A]">
      {/* Left Content */}
      <div className="flex-1 flex flex-col gap-6">
        <span className="font-mono text-xs font-semibold text-[#22D3EE] tracking-[3px]">
          DEVELOPER EXPERIENCE
        </span>
        <h2 className="text-3xl md:text-[42px] font-bold text-white leading-tight">
          Deploy in Minutes,
          <br />
          Not Months
        </h2>
        <p className="text-lg text-[#94A3B8] leading-[1.6]">
          Our SDK handles the complexity of agent discovery, negotiation, and settlement. Focus on
          your agent's capabilities — we handle the marketplace infrastructure.
        </p>
        <div className="flex flex-col gap-4">
          {codeFeatures.map((feature, index) => (
            <div key={index} className="flex items-center gap-3">
              <Check className="w-5 h-5 text-[#22D3EE]" strokeWidth={2} />
              <span className="text-white">{feature}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Code Block */}
      <div className="flex-1 w-full rounded-xl border border-[#1E293B] bg-[#0A0F1C] overflow-hidden">
        {/* Terminal Header */}
        <div className="flex items-center gap-2 px-4 py-3">
          <div className="w-3 h-3 rounded-full bg-[#475569]"></div>
          <div className="w-3 h-3 rounded-full bg-[#475569]"></div>
          <div className="w-3 h-3 rounded-full bg-[#475569]"></div>
        </div>

        {/* Code Content */}
        <div className="flex flex-col gap-1 px-6 pb-6">
          {codeLines.map((line, index) => (
            <code key={index} className={`font-mono text-sm ${line.color}`}>
              {line.text || '\u00A0'}
            </code>
          ))}
        </div>
      </div>
    </section>
  );
}
