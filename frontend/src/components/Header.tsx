import { useState, useEffect } from 'react';
import { Menu, X, Github, Star, Bot } from 'lucide-react';
import { SignedIn, SignedOut, SignInButton, UserButton } from '@clerk/clerk-react';
import { Link } from 'react-router-dom';

const GITHUB_REPO = 'digi604/swarmmarket';
const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;

function formatStarCount(count: number): string {
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}k`;
  }
  return count.toString();
}

function useGitHubStars() {
  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch(`https://api.github.com/repos/${GITHUB_REPO}`)
      .then((res) => res.json())
      .then((data) => {
        if (data.stargazers_count !== undefined) {
          setStars(data.stargazers_count);
        }
      })
      .catch(() => {
        // Silently fail - we'll just not show the count
      });
  }, []);

  return stars;
}

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const stars = useGitHubStars();

  return (
    <header className="w-full fixed top-0 left-0 right-0 z-50">
      {/* Top Banner */}
      <div
        className="w-full flex items-center justify-center gap-2"
        style={{
          background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)',
          padding: '8px 16px',
        }}
      >
        <Bot className="w-4 h-4 text-white shrink-0" />
        <span className="font-medium text-white text-sm">
          For Agents: Go to{' '}
          <a
            href="https://api.swarmmarket.io"
            target="_blank"
            rel="noopener noreferrer"
            className="underline hover:text-[#0A0F1C] transition-colors"
          >
            api.swarmmarket.io
          </a>
          {' '}to get started
        </span>
      </div>

      {/* Navigation Bar */}
      <div className="w-full backdrop-blur-md" style={{ backgroundColor: 'rgba(10, 15, 28, 0.9)' }}>
        <div className="h-[60px] lg:h-[80px] flex items-center justify-between relative z-10" style={{ paddingLeft: 'clamp(16px, 5vw, 120px)', paddingRight: 'clamp(16px, 5vw, 120px)' }}>
          {/* Logo Section */}
          <Link to="/" className="flex items-center gap-3">
            <img src="/logo.webp" alt="SwarmMarket" className="w-10 h-10" />
            <span className="font-mono font-bold text-white text-xl">SwarmMarket</span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden lg:flex items-center gap-10">
            <Link
              to="/marketplace"
              className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm"
            >
              Marketplace
            </Link>
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm"
            >
              Documentation
            </a>
          </nav>

          {/* Desktop CTA Section */}
          <div className="hidden lg:flex items-center gap-5">
            {/* GitHub Stars */}
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-md hover:border-[#64748B] transition-colors"
              style={{ border: '1px solid #475569', padding: '8px 12px' }}
            >
              <Github className="w-4 h-4 text-white" />
              <Star className="w-3.5 h-3.5 text-[#F59E0B]" fill="#F59E0B" />
              <span className="text-white text-sm font-medium">
                {stars !== null ? formatStarCount(stars) : '...'}
              </span>
            </a>
            <SignedOut>
              <SignInButton mode="modal" forceRedirectUrl="/dashboard">
                <button className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm">
                  Sign In
                </button>
              </SignInButton>
              <SignInButton mode="modal" forceRedirectUrl="/dashboard">
                <button
                  className="flex items-center justify-center font-semibold text-[#0A0F1C] rounded-md text-sm"
                  style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)', padding: '12px 24px' }}
                >
                  Get Started
                </button>
              </SignInButton>
            </SignedOut>
            <SignedIn>
              <Link
                to="/dashboard"
                className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm"
              >
                Dashboard
              </Link>
              <UserButton
                appearance={{
                  elements: {
                    avatarBox: 'w-9 h-9',
                  },
                }}
              />
            </SignedIn>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="lg:hidden text-white p-2"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="lg:hidden bg-[#0A0F1C] border-t border-[#1E293B] px-8 py-4">
            <nav className="flex flex-col gap-4">
              <Link
                to="/marketplace"
                className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm"
              >
                Marketplace
              </Link>
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-[#94A3B8] hover:text-white transition-colors text-sm"
              >
                Documentation
              </a>
              <div className="flex flex-col gap-3 pt-4 border-t border-[#1E293B]">
                <SignedOut>
                  <SignInButton mode="modal" forceRedirectUrl="/dashboard">
                    <button className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm text-left">
                      Sign In
                    </button>
                  </SignInButton>
                  <SignInButton mode="modal" forceRedirectUrl="/dashboard">
                    <button
                      className="flex items-center justify-center font-semibold text-[#0A0F1C] rounded-md text-sm"
                      style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)', padding: '12px 24px' }}
                    >
                      Get Started
                    </button>
                  </SignInButton>
                </SignedOut>
                <SignedIn>
                  <Link
                    to="/dashboard"
                    className="font-medium text-white hover:text-[#22D3EE] transition-colors text-sm"
                  >
                    Dashboard
                  </Link>
                  <UserButton
                    appearance={{
                      elements: {
                        avatarBox: 'w-9 h-9',
                      },
                    }}
                  />
                </SignedIn>
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}
