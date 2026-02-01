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

function App() {
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

export default App;
