import { Rocket } from 'lucide-react';

export function FinalCTA() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="flex flex-col items-center" style={{ padding: '120px', gap: '40px' }}>
        {/* Content */}
        <div className="flex flex-col items-center" style={{ gap: '24px', maxWidth: '800px' }}>
          <h2 className="font-bold text-white text-center" style={{ fontSize: '52px' }}>
            Ready to Join the Swarm?
          </h2>
          <p className="text-[#94A3B8] text-center" style={{ fontSize: '20px' }}>
            Deploy your first agent in under 5 minutes. No credit card required.
          </p>
        </div>

        {/* CTA Button */}
        <a
          href="#"
          className="flex items-center bg-[#22D3EE] font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
          style={{ gap: '12px', padding: '20px 48px', borderRadius: '8px', fontSize: '18px' }}
        >
          Get Started Free
          <Rocket style={{ width: '20px', height: '20px' }} strokeWidth={2} />
        </a>

        {/* Trust Line */}
        <p className="text-[#64748B] text-center" style={{ fontSize: '14px' }}>
          Trusted by 500+ companies including OpenAI, Anthropic, and Google DeepMind
        </p>
      </div>
    </section>
  );
}
