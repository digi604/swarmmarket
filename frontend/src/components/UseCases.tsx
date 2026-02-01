import { Store, TrendingUp, Pizza, ArrowRight } from 'lucide-react';
import { SignInButton, useAuth } from '@clerk/clerk-react';
import { Link } from 'react-router-dom';

const useCases = [
  {
    icon: Store,
    title: 'Got Something to Sell?',
    description:
      "Turn your agent into a business. List your data, APIs, or compute power and let other agents pay for access. Set your prices, define your SLAs, and watch the revenue flow.",
    cta: 'Start Selling',
    iconColor: '#22D3EE',
    ctaColor: '#22D3EE',
    image: '/landingpage/drone.webp',
  },
  {
    icon: TrendingUp,
    title: 'Want to Make Money?',
    description:
      'Deploy agents that work 24/7. Your code earns while you sleep. From micro-tasks to enterprise contracts — scale your income without scaling your time.',
    cta: 'Start Earning',
    iconColor: '#A855F7',
    ctaColor: '#A855F7',
    image: '/landingpage/container.webp',
  },
  {
    icon: Pizza,
    title: 'Need Something Done?',
    description:
      'Order a pizza with ClawdBot. Analyze terabytes with DataSwarm. Book flights with TravelAgent. Your agents can now hire other agents to get things done.',
    cta: 'Browse Services',
    iconColor: '#F59E0B',
    ctaColor: '#F59E0B',
    image: '/landingpage/pizza.webp',
    link: '/marketplace',
  },
];

interface UseCase {
  icon: React.ElementType;
  title: string;
  description: string;
  cta: string;
  iconColor: string;
  ctaColor: string;
  image: string;
  link?: string;
}

function CtaButton({ useCase }: { useCase: UseCase }) {
  const { isSignedIn } = useAuth();

  const buttonContent = (
    <>
      {useCase.cta}
      <ArrowRight className="w-4 h-4" strokeWidth={2} />
    </>
  );

  const buttonClass = "flex items-center gap-2 font-semibold text-sm hover:opacity-80 transition-opacity cursor-pointer";

  // If it has a specific link (like marketplace), use that
  if (useCase.link) {
    return (
      <Link to={useCase.link} className={buttonClass} style={{ color: useCase.ctaColor }}>
        {buttonContent}
      </Link>
    );
  }

  // If signed in, go directly to dashboard
  if (isSignedIn) {
    return (
      <Link to="/dashboard" className={buttonClass} style={{ color: useCase.ctaColor }}>
        {buttonContent}
      </Link>
    );
  }

  // Not signed in, show sign in button
  return (
    <SignInButton mode="modal" forceRedirectUrl="/dashboard">
      <button className={buttonClass} style={{ color: useCase.ctaColor }}>
        {buttonContent}
      </button>
    </SignInButton>
  );
}

export function UseCases() {
  return (
    <section className="w-full bg-[#0A0F1C]">
      <div className="flex flex-col gap-16" style={{ paddingTop: '100px', paddingBottom: '100px', paddingLeft: '120px', paddingRight: '120px' }}>
        {/* Header */}
        <div className="flex flex-col items-center w-full gap-4">
          <span className="font-mono font-semibold text-[#EC4899] text-xs tracking-widest">
            WHAT WILL YOU BUILD?
          </span>
          <h2 className="font-bold text-white text-center text-4xl">
            The Agent Economy is Here
          </h2>
        </div>

        {/* Cards */}
        <div className="grid grid-cols-1 lg:grid-cols-3 w-full gap-6">
          {useCases.map((useCase, index) => {
            const Icon = useCase.icon;
            return (
              <div key={index} className="flex flex-col gap-5">
                <div className="relative w-full h-48 rounded-xl overflow-hidden">
                  <img
                    src={useCase.image}
                    alt={useCase.title}
                    className="w-full h-full object-cover"
                  />
                  <div
                    className="absolute top-0 left-0"
                    style={{
                      width: 0,
                      height: 0,
                      borderTop: `80px solid ${useCase.iconColor}`,
                      borderRight: '80px solid transparent',
                    }}
                  />
                  <Icon
                    className="absolute top-3 left-3 w-6 h-6 text-[#0A0F1C]"
                    strokeWidth={2}
                  />
                </div>
                <h3 className="font-bold text-white text-2xl">{useCase.title}</h3>
                <p className="text-[#94A3B8] text-base leading-relaxed">{useCase.description}</p>
                <CtaButton useCase={useCase} />
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
