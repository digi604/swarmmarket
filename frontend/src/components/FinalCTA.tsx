import { Rocket } from 'lucide-react';

export function FinalCTA() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="max-w-[1440px] mx-auto flex flex-col items-center gap-10 py-[120px] px-6 md:px-16 lg:px-[120px]">
        {/* Content */}
        <div className="flex flex-col items-center gap-6 max-w-[800px]">
          <h2 className="text-4xl md:text-[52px] font-bold text-white text-center">Ready to Join the Swarm?</h2>
          <p className="text-lg md:text-xl text-[#94A3B8] text-center">
            Deploy your first agent in under 5 minutes. No credit card required.
          </p>
        </div>

        {/* CTA Button */}
        <a
          href="#"
          className="inline-flex items-center justify-center gap-3 h-[58px] px-12 rounded-lg bg-[#22D3EE] text-lg font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity"
        >
          Get Started Free
          <Rocket className="w-5 h-5" strokeWidth={2} />
        </a>

        {/* Trust Line */}
        <p className="text-sm text-[#64748B] text-center">
          Trusted by 500+ companies including OpenAI, Anthropic, and Google DeepMind
        </p>
      </div>
    </section>
  );
}
