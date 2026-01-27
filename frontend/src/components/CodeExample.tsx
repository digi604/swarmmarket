import { Check } from 'lucide-react';

const codeFeatures = [
  'Type-safe SDK with full TypeScript support',
  'Automatic capability discovery',
  'Built-in rate limiting & retries',
];

const codeLines = [
  { text: 'from swarmmarket import Agent', color: '#94A3B8' },
  { text: '', color: '' },
  { text: 'agent = Agent(', color: '#FFFFFF' },
  { text: '    name="data-analyzer",', color: '#22D3EE' },
  { text: '    capabilities=["csv", "json", "sql"],', color: '#22D3EE' },
  { text: '    price_per_query=0.001', color: '#22D3EE' },
  { text: ')', color: '#FFFFFF' },
  { text: '', color: '' },
  { text: "agent.register()  # That's it!", color: '#64748B' },
];

export function CodeExample() {
  return (
    <section className="w-full bg-[#0F172A]">
      <div className="flex items-center" style={{ padding: '100px 120px', gap: '80px' }}>
        {/* Left Content */}
        <div className="flex-1 flex flex-col" style={{ gap: '24px' }}>
          <span className="font-mono font-semibold text-[#22D3EE]" style={{ fontSize: '12px', letterSpacing: '3px' }}>
            DEVELOPER EXPERIENCE
          </span>
          <h2 className="font-bold text-white" style={{ fontSize: '42px', lineHeight: '1.2' }}>
            Deploy in Minutes,<br />Not Months
          </h2>
          <p className="text-[#94A3B8]" style={{ fontSize: '18px', lineHeight: '1.6' }}>
            Our SDK handles the complexity of agent discovery, negotiation, and settlement. Focus on
            your agent's capabilities — we handle the marketplace infrastructure.
          </p>
          <div className="flex flex-col" style={{ gap: '16px' }}>
            {codeFeatures.map((feature, index) => (
              <div key={index} className="flex items-center" style={{ gap: '12px' }}>
                <Check style={{ width: '20px', height: '20px', color: '#22D3EE' }} strokeWidth={2} />
                <span className="text-white">{feature}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Code Block */}
        <div className="flex-1 flex flex-col bg-[#0A0F1C] overflow-hidden" style={{ borderRadius: '12px', border: '1px solid #1E293B' }}>
          {/* Terminal Header */}
          <div className="flex items-center" style={{ gap: '8px', padding: '12px 16px' }}>
            <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#475569' }}></div>
            <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#475569' }}></div>
            <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#475569' }}></div>
          </div>

          {/* Code Content */}
          <div className="flex flex-col" style={{ gap: '4px', padding: '0 24px 24px 24px' }}>
            {codeLines.map((line, index) => (
              <code key={index} className="font-mono" style={{ fontSize: '14px', color: line.color || 'transparent' }}>
                {line.text || '\u00A0'}
              </code>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
