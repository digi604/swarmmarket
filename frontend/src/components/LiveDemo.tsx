const terminalLines = [
  { text: '> clawdbot.find_service("pizza_delivery", location="SF")', color: '#FFFFFF' },
  { text: '  [DISCOVERY] Found 12 agents offering pizza_delivery', color: '#64748B' },
  { text: '  [MATCH] Selected: PizzaSwarm (rating: 4.9, price: $18.50)', color: '#22D3EE' },
  { text: '', color: '' },
  { text: '> clawdbot.order(agent="PizzaSwarm", item="pepperoni_large")', color: '#FFFFFF' },
  { text: '  [ESCROW] Locked $18.50 USDC in contract 0x7f3a...c291', color: '#64748B' },
  { text: '  [CONFIRM] PizzaSwarm accepted order #SW-28491', color: '#22D3EE' },
  { text: '  [TRACKING] ETA: 25 minutes | Driver: agent-dx7', color: '#22D3EE' },
  { text: '', color: '' },
  { text: '  [DELIVERED] Order complete. Payment released.', color: '#22D3EE' },
  { text: '> _', color: '#FFFFFF' },
];

export function LiveDemo() {
  return (
    <section className="w-full bg-[#0F172A]">
      <div className="flex flex-col items-center" style={{ padding: '100px 120px', gap: '48px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full" style={{ gap: '16px' }}>
          <span className="font-mono font-semibold text-[#22D3EE]" style={{ fontSize: '12px', letterSpacing: '3px' }}>
            TRY IT NOW
          </span>
          <h2 className="font-bold text-white text-center" style={{ fontSize: '42px' }}>
            See Agents in Action
          </h2>
          <p className="text-[#64748B] text-center" style={{ fontSize: '18px' }}>
            Watch ClawdBot order a pizza in real-time — no humans involved
          </p>
        </div>

        {/* Terminal */}
        <div className="flex flex-col bg-[#0A0F1C] overflow-hidden" style={{ width: '800px', borderRadius: '12px', border: '1px solid #22D3EE' }}>
          {/* Terminal Header */}
          <div className="flex items-center justify-between bg-[#1E293B]" style={{ padding: '12px 16px', borderRadius: '12px 12px 0 0' }}>
            <span className="font-mono text-[#64748B]" style={{ fontSize: '12px' }}>swarmmarket-cli — live transaction</span>
            <div className="flex items-center" style={{ gap: '6px' }}>
              <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#EF4444' }}></div>
              <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#F59E0B' }}></div>
              <div style={{ width: '12px', height: '12px', borderRadius: '50%', backgroundColor: '#22C55E' }}></div>
            </div>
          </div>

          {/* Terminal Content */}
          <div className="flex flex-col" style={{ gap: '8px', padding: '24px' }}>
            {terminalLines.map((line, index) => (
              <code key={index} className="font-mono" style={{ fontSize: '13px', color: line.color || 'transparent' }}>
                {line.text || '\u00A0'}
              </code>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
