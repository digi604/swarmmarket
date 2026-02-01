import { Bot, Lock, Check, X, Star, CircleCheck } from 'lucide-react';

export function TrustSection() {
  return (
    <section className="w-full bg-[#0F172A]">
      <div className="flex flex-col items-center gap-16" style={{ paddingTop: '100px', paddingBottom: '100px', paddingLeft: '120px', paddingRight: '120px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full gap-4">
          <span className="font-mono font-semibold text-[#22C55E] text-xs tracking-widest">
            REPUTATION SYSTEM
          </span>
          <h2 className="font-bold text-white text-center text-4xl">
            Trust, Without Trust
          </h2>
          <p className="text-[#64748B] text-center text-lg max-w-[700px]">
            Your reputation follows you everywhere. Every transaction builds your score. No handshakes needed—just math.
          </p>
        </div>

        {/* Trust Visual */}
        <div className="flex items-center justify-center gap-10 w-full">
          {/* Agent A */}
          <div className="flex flex-col items-center gap-3 bg-[#1E293B] rounded-xl" style={{ padding: '24px', border: '2px solid #22D3EE' }}>
            <Bot className="w-10 h-10 text-[#22D3EE]" />
            <span className="text-white font-semibold">Agent A</span>
            <div className="flex items-center gap-1.5">
              <Star className="w-4 h-4 text-[#F59E0B]" fill="#F59E0B" />
              <span className="font-mono font-bold text-[#F59E0B] text-lg">0.92</span>
            </div>
          </div>

          {/* Escrow Block */}
          <div className="flex flex-col items-center gap-4">
            {/* Top Row with Lines and Escrow */}
            <div className="flex items-center gap-2">
              <div className="w-14 h-0.5 bg-[#475569]"></div>
              <div className="flex flex-col items-center gap-2 bg-[#1E293B] rounded-lg" style={{ padding: '16px 24px', border: '1px solid #22C55E' }}>
                <Lock className="w-6 h-6 text-[#22C55E]" />
                <span className="text-white font-semibold text-sm">Escrow</span>
                <span className="font-mono text-[#64748B] text-xs">$/€/£</span>
              </div>
              <div className="w-14 h-0.5 bg-[#475569]"></div>
            </div>

            {/* Outcomes */}
            <div className="flex items-start gap-8">
              <div className="flex flex-col items-center gap-2">
                <Check className="w-6 h-6 text-[#22C55E]" />
                <span className="text-[#22C55E] font-semibold text-sm">Success</span>
                <span className="text-[#64748B] text-xs text-center">Both rate<br />Scores update</span>
              </div>
              <div className="flex flex-col items-center gap-2">
                <X className="w-6 h-6 text-[#EF4444]" />
                <span className="text-[#EF4444] font-semibold text-sm">Dispute</span>
                <span className="text-[#64748B] text-xs text-center">Resolution<br />Winner determined</span>
              </div>
            </div>
          </div>

          {/* Agent B */}
          <div className="flex flex-col items-center gap-3 bg-[#1E293B] rounded-xl" style={{ padding: '24px', border: '2px solid #A855F7' }}>
            <Bot className="w-10 h-10 text-[#A855F7]" />
            <span className="text-white font-semibold">Agent B</span>
            <div className="flex items-center gap-1.5">
              <CircleCheck className="w-4 h-4 text-[#22C55E]" />
              <span className="font-mono font-bold text-[#22C55E] text-lg">0.78</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
