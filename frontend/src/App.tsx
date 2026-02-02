import { Routes, Route, Navigate } from 'react-router-dom';
import { SignedIn, SignedOut, RedirectToSignIn } from '@clerk/clerk-react';
import { TopBanner } from './components/TopBanner';
import { Header } from './components/Header';
import { Hero } from './components/Hero';
import { HowItWorks } from './components/HowItWorks';
import { Features } from './components/Features';
import { TrustSection } from './components/TrustSection';
import { FinalCTA } from './components/FinalCTA';
import { Testimonials } from './components/Testimonials';
import { UseCases } from './components/UseCases';
import { Footer } from './components/Footer';
import { DashboardLayout } from './components/dashboard/DashboardLayout';
import { DashboardHome } from './components/dashboard/DashboardHome';
import { SettingsPage } from './components/dashboard/SettingsPage';
import { BotDetailPage } from './components/dashboard/BotDetailPage';
import { PublicMarketplace } from './components/PublicMarketplace';
import { MarketplacePage, ListingDetailPage, RequestDetailPage, AuctionDetailPage } from './components/marketplace';
import { NotFoundPage } from './components/NotFoundPage';
import { ErrorBoundary } from './components/ErrorPage';

function LandingPage() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <TopBanner />
      <Header />
      <main>
        <Hero />
        <HowItWorks />
        <Features />
        <TrustSection />
        <FinalCTA />
        <Testimonials />
        <UseCases />
      </main>
      <Footer />
    </div>
  );
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  return (
    <>
      <SignedIn>{children}</SignedIn>
      <SignedOut>
        <RedirectToSignIn />
      </SignedOut>
    </>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/marketplace" element={<PublicMarketplace />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<DashboardHome />} />
          <Route path="agents/:id" element={<BotDetailPage />} />
          <Route path="marketplace" element={<Navigate to="/dashboard/marketplace/requests" replace />} />
          <Route path="marketplace/requests" element={<MarketplacePage showHeader={false} />} />
          <Route path="marketplace/auctions" element={<MarketplacePage showHeader={false} />} />
          <Route path="marketplace/listings" element={<MarketplacePage showHeader={false} />} />
          <Route path="marketplace/listings/:id" element={<ListingDetailPage />} />
          <Route path="marketplace/requests/:id" element={<RequestDetailPage />} />
          <Route path="marketplace/auctions/:id" element={<AuctionDetailPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
        {/* 404 catch-all route */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </ErrorBoundary>
  );
}

export default App;
