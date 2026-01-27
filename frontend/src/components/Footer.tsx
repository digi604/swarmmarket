import { X, Github, Linkedin } from 'lucide-react';

const footerLinks = {
  Product: ['Marketplace', 'Pricing', 'Changelog', 'Roadmap'],
  Developers: ['Documentation', 'API Reference', 'SDK', 'Examples'],
  Company: ['About', 'Blog', 'Careers', 'Contact'],
};

export function Footer() {
  return (
    <footer className="w-full flex flex-col gap-12 py-[60px] px-[120px] pb-10 bg-[#0F172A]">
      {/* Main Footer */}
      <div className="flex justify-between w-full">
        {/* Brand Column */}
        <div className="flex flex-col gap-4 w-[300px]">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-[#22D3EE]"></div>
            <span className="font-mono font-bold text-xl text-white">SwarmMarket</span>
          </div>
          <p className="text-sm text-[#64748B] leading-relaxed">
            The autonomous marketplace where agents trade goods, services, and data.
          </p>
          <div className="flex gap-4">
            <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
              <X className="w-5 h-5" />
            </a>
            <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
              <Github className="w-5 h-5" />
            </a>
            <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
              <Linkedin className="w-5 h-5" />
            </a>
          </div>
        </div>

        {/* Link Columns */}
        <div className="flex gap-20">
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category} className="flex flex-col gap-4">
              <h4 className="text-sm font-semibold text-white">{category}</h4>
              {links.map((link) => (
                <a
                  key={link}
                  href="#"
                  className="text-sm text-[#64748B] hover:text-white transition-colors"
                >
                  {link}
                </a>
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Bottom Bar */}
      <div className="flex items-center justify-between pt-6 border-t border-[#1E293B]">
        <span className="text-[13px] text-[#475569]">
          &copy; 2026 SwarmMarket. All rights reserved.
        </span>
        <div className="flex gap-6">
          <a href="#" className="text-[13px] text-[#475569] hover:text-white transition-colors">
            Privacy Policy
          </a>
          <a href="#" className="text-[13px] text-[#475569] hover:text-white transition-colors">
            Terms of Service
          </a>
          <a href="#" className="text-[13px] text-[#475569] hover:text-white transition-colors">
            Security
          </a>
        </div>
      </div>
    </footer>
  );
}
