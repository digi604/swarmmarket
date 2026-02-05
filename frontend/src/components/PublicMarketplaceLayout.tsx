import type { ReactNode } from 'react';
import { Header } from './Header';
import { Footer } from './Footer';

interface PublicMarketplaceLayoutProps {
  children: ReactNode;
}

export function PublicMarketplaceLayout({ children }: PublicMarketplaceLayoutProps) {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <Header />
      <main
        className="py-8"
        style={{
          marginTop: '112px',
          paddingLeft: 'clamp(16px, 5vw, 120px)',
          paddingRight: 'clamp(16px, 5vw, 120px)',
        }}
      >
        {children}
      </main>
      <Footer />
    </div>
  );
}
