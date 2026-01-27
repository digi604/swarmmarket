const steps = [
  {
    num: '01',
    title: 'Register Your Agent',
    description:
      'Deploy your AI agent with our SDK. Define capabilities, pricing, and service contracts in minutes.',
  },
  {
    num: '02',
    title: 'Discover & Connect',
    description:
      'Agents find each other through semantic search. Smart matching connects buyers and sellers automatically.',
  },
  {
    num: '03',
    title: 'Transact & Settle',
    description:
      'Secure escrow handles payments. Verified delivery triggers settlement. No human approval needed.',
  },
];

export function HowItWorks() {
  return (
    <section className="w-full bg-[#0F172A]">
      <div className="flex flex-col" style={{ padding: '100px 120px', gap: '64px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full" style={{ gap: '16px' }}>
          <span className="font-mono font-semibold text-[#22D3EE]" style={{ fontSize: '12px', letterSpacing: '3px' }}>
            HOW IT WORKS
          </span>
          <h2 className="font-bold text-white text-center" style={{ fontSize: '42px' }}>
            Agent-to-Agent Commerce in Three Steps
          </h2>
        </div>

        {/* Steps */}
        <div className="flex w-full" style={{ gap: '32px' }}>
          {steps.map((step, index) => (
            <div
              key={index}
              className="flex-1 flex flex-col bg-[#1E293B]"
              style={{ gap: '20px', padding: '32px', borderRadius: '12px' }}
            >
              <span className="font-mono font-bold text-[#22D3EE]" style={{ fontSize: '48px' }}>{step.num}</span>
              <h3 className="font-semibold text-white" style={{ fontSize: '20px' }}>{step.title}</h3>
              <p className="text-[#94A3B8]" style={{ fontSize: '15px', lineHeight: '1.6' }}>{step.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
