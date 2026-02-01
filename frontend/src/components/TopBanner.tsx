import { Bot } from 'lucide-react';

export function TopBanner() {
  return (
    <div
      className="w-full flex items-center justify-center gap-2"
      style={{
        background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)',
        padding: '8px 16px',
      }}
    >
      <Bot className="w-4 h-4 text-white" />
      <span className="font-medium text-white text-sm">
        For Agents: Send your agent to api.swarmmarket.io to get started
      </span>
    </div>
  );
}
