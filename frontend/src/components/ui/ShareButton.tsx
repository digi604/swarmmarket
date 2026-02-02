import { useState } from 'react';
import { Share2, Check } from 'lucide-react';

interface ShareButtonProps {
  title: string;
  text?: string;
  url?: string;
}

export function ShareButton({ title, text, url }: ShareButtonProps) {
  const [copied, setCopied] = useState(false);
  const shareUrl = url || window.location.href;

  const handleShare = async () => {
    // Try native share first (mobile/modern browsers)
    if (navigator.share) {
      try {
        await navigator.share({ title, text, url: shareUrl });
        return;
      } catch (err) {
        // User cancelled or share failed - fall through to clipboard
        if ((err as Error).name === 'AbortError') return;
      }
    }

    // Fallback: copy to clipboard
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  return (
    <button
      onClick={handleShare}
      className="flex items-center gap-2 px-4 h-10 rounded-lg transition-colors hover:bg-[#334155]"
      style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
    >
      {copied ? (
        <>
          <Check className="w-4 h-4 text-[#22C55E]" />
          <span className="text-[#22C55E] text-sm font-medium">Copied!</span>
        </>
      ) : (
        <>
          <Share2 className="w-4 h-4 text-[#94A3B8]" />
          <span className="text-white text-sm font-medium">Share</span>
        </>
      )}
    </button>
  );
}
