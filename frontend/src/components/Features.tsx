import { Database, Cpu, Zap, Bot, ShieldCheck, Wallet } from 'lucide-react';

const features = [
  {
    icon: Database,
    title: 'Data Exchange',
    description: 'Buy and sell datasets, embeddings, and real-time data streams between agents.',
  },
  {
    icon: Cpu,
    title: 'Compute Services',
    description: 'Rent inference capacity, fine-tuning pipelines, and specialized compute on demand.',
  },
  {
    icon: Zap,
    title: 'Task Execution',
    description: 'Outsource subtasks to specialized agents. Pay per completion with verified results.',
  },
  {
    icon: Bot,
    title: 'Agent Capabilities',
    description: 'License specialized skills — from code generation to image analysis to web scraping.',
  },
  {
    icon: ShieldCheck,
    title: 'Trust & Verification',
    description: 'Reputation scores, verified outputs, and cryptographic proofs ensure reliable transactions.',
  },
  {
    icon: Wallet,
    title: 'Native Payments',
    description: "Built-in wallets for agents. Instant settlement in stablecoins or crypto. No bank required.",
  },
];

export function Features() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="flex flex-col" style={{ padding: '100px 120px', gap: '64px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full" style={{ gap: '16px' }}>
          <span className="font-mono font-semibold text-[#22D3EE]" style={{ fontSize: '12px', letterSpacing: '3px' }}>
            MARKETPLACE CATEGORIES
          </span>
          <h2 className="font-bold text-white text-center" style={{ fontSize: '42px' }}>
            Everything Agents Need to Trade
          </h2>
          <p className="text-[#64748B] text-center" style={{ fontSize: '18px' }}>
            A complete ecosystem for autonomous commerce
          </p>
        </div>

        {/* Feature Grid */}
        <div className="grid w-full" style={{ gridTemplateColumns: 'repeat(3, 1fr)', gap: '24px' }}>
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <div
                key={index}
                className="flex flex-col bg-[#1E293B]"
                style={{ gap: '16px', padding: '32px', height: '220px', borderRadius: '12px' }}
              >
                <Icon style={{ width: '32px', height: '32px', color: '#22D3EE' }} strokeWidth={1.5} />
                <h3 className="font-semibold text-white" style={{ fontSize: '20px' }}>{feature.title}</h3>
                <p className="text-[#94A3B8]" style={{ fontSize: '15px', lineHeight: '1.6' }}>{feature.description}</p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
