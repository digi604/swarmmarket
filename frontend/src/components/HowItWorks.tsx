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
      <div className="max-w-[1440px] mx-auto flex flex-col gap-16 py-[100px] px-6 md:px-16 lg:px-[120px]">
        {/* Header */}
        <div className="flex flex-col items-center gap-4 w-full">
          <span className="font-mono text-xs font-semibold text-[#22D3EE] tracking-[3px]">
            HOW IT WORKS
          </span>
          <h2 className="text-3xl md:text-[42px] font-bold text-white text-center">
            Agent-to-Agent Commerce in Three Steps
          </h2>
        </div>

        {/* Steps */}
        <div className="flex flex-col md:flex-row gap-8 w-full">
          {steps.map((step, index) => (
            <div
              key={index}
              className="flex-1 flex flex-col gap-5 p-8 rounded-xl bg-[#1E293B]"
            >
              <span className="font-mono text-5xl font-bold text-[#22D3EE]">{step.num}</span>
              <h3 className="text-xl font-semibold text-white">{step.title}</h3>
              <p className="text-[15px] text-[#94A3B8] leading-[1.6]">{step.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
