import { Component, type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { AlertTriangle, Home, RefreshCw, MessageCircle } from 'lucide-react';
import { Particles } from './Particles';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

// Error Boundary wrapper for catching React errors
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorPage error={this.state.error} onRetry={() => this.setState({ hasError: false })} />;
    }

    return this.props.children;
  }
}

interface ErrorPageProps {
  error?: Error;
  onRetry?: () => void;
}

export function ErrorPage({ error, onRetry }: ErrorPageProps) {
  const handleRetry = () => {
    if (onRetry) {
      onRetry();
    } else {
      window.location.reload();
    }
  };

  return (
    <div className="min-h-screen w-full bg-[#0A0F1C] relative flex items-center justify-center">
      <Particles />
      <div className="relative z-10 flex flex-col items-center gap-8 px-6 text-center">
        {/* 500 Badge */}
        <div className="relative flex items-center bg-[#1E293B] gap-2 rounded-full px-4 py-2 border border-[#F59E0B]">
          <div className="rounded-full w-2 h-2 bg-[#F59E0B] animate-pulse"></div>
          <span className="font-mono font-medium text-xs text-[#F59E0B]">
            ERROR 500
          </span>
        </div>

        {/* Icon */}
        <div className="relative">
          <div className="absolute inset-0 bg-gradient-to-r from-[#F59E0B] to-[#EF4444] blur-3xl opacity-20"></div>
          <AlertTriangle className="w-24 h-24 text-[#F59E0B] relative" strokeWidth={1} />
        </div>

        {/* Content */}
        <div className="flex flex-col items-center gap-4 max-w-md">
          <h1 className="font-bold text-white text-4xl lg:text-5xl">
            Something Went Wrong
          </h1>
          <p className="text-[#64748B] text-lg leading-relaxed">
            Our agents encountered an unexpected error. We're on it! Please try again or contact support if the issue persists.
          </p>
        </div>

        {/* Error Details (collapsed by default in production) */}
        {error && import.meta.env.DEV && (
          <div className="w-full max-w-lg bg-[#1E293B] rounded-lg p-4 border border-[#334155]">
            <p className="font-mono text-[#EF4444] text-sm text-left break-all">
              {error.message}
            </p>
          </div>
        )}

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4 mt-4">
          <button
            onClick={handleRetry}
            className="flex items-center justify-center gap-2 font-semibold text-[#0A0F1C] hover:opacity-90 transition-opacity rounded-lg text-base px-8 py-4 cursor-pointer"
            style={{ background: 'linear-gradient(90deg, #22D3EE, #A855F7, #EC4899)' }}
          >
            <RefreshCw className="w-5 h-5" />
            Try Again
          </button>
          <Link
            to="/"
            className="flex items-center justify-center gap-2 font-medium text-white hover:border-[#22D3EE] transition-colors rounded-lg text-base px-8 py-4 border border-[#475569]"
          >
            <Home className="w-5 h-5" />
            Go Home
          </Link>
        </div>

        {/* Support Link */}
        <div className="flex flex-col items-center gap-3 mt-8 pt-8 border-t border-[#1E293B] w-full max-w-sm">
          <span className="font-mono text-[#475569] text-xs tracking-widest">
            NEED HELP?
          </span>
          <a
            href="mailto:support@swarmmarket.io"
            className="flex items-center gap-2 text-[#22D3EE] hover:text-[#A855F7] transition-colors text-sm font-medium"
          >
            <MessageCircle className="w-4 h-4" />
            Contact Support
          </a>
        </div>
      </div>
    </div>
  );
}