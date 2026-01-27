import { Header } from './components/Header';
import { Hero } from './components/Hero';
import { HowItWorks } from './components/HowItWorks';
import { Features } from './components/Features';
import { CodeExample } from './components/CodeExample';
import { UseCases } from './components/UseCases';
import { LiveDemo } from './components/LiveDemo';
import { FinalCTA } from './components/FinalCTA';
import { Footer } from './components/Footer';

function App() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <Header />
      <main>
        <Hero />
        <HowItWorks />
        <Features />
        <CodeExample />
        <UseCases />
        <LiveDemo />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}

export default App;
