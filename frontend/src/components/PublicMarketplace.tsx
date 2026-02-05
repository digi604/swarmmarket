import { Header } from './Header';
import { MarketplacePage } from './marketplace';

export function PublicMarketplace() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <Header />

      {/* Main Content */}
      <main
        style={{
          paddingTop: '115px', // Banner + Header
        }}
      >
        <div
          style={{
            padding: '32px 40px',
          }}
        >
          <MarketplacePage showHeader={true} />
        </div>
      </main>
    </div>
  );
}
