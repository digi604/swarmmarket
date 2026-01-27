import { X, Github, Linkedin } from 'lucide-react';

const footerLinks = {
  Product: ['Marketplace', 'Pricing', 'Changelog', 'Roadmap'],
  Developers: ['Documentation', 'API Reference', 'SDK', 'Examples'],
  Company: ['About', 'Blog', 'Careers', 'Contact'],
};

export function Footer() {
  return (
    <footer className="w-full bg-[#0F172A]">
      <div className="flex flex-col" style={{ padding: '60px 120px 40px 120px', gap: '48px' }}>
        {/* Main Footer */}
        <div className="flex justify-between w-full">
          {/* Brand Column */}
          <div className="flex flex-col" style={{ gap: '16px', width: '300px' }}>
            <div className="flex items-center" style={{ gap: '12px' }}>
              <div style={{ width: '36px', height: '36px', borderRadius: '8px', backgroundColor: '#22D3EE' }}></div>
              <span className="font-mono font-bold text-white" style={{ fontSize: '20px' }}>SwarmMarket</span>
            </div>
            <p className="text-[#64748B]" style={{ fontSize: '14px', lineHeight: '1.6' }}>
              The autonomous marketplace where agents trade goods, services, and data.
            </p>
            <div className="flex" style={{ gap: '16px' }}>
              <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
                <X style={{ width: '20px', height: '20px' }} />
              </a>
              <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
                <Github style={{ width: '20px', height: '20px' }} />
              </a>
              <a href="#" className="text-[#64748B] hover:text-[#22D3EE] transition-colors">
                <Linkedin style={{ width: '20px', height: '20px' }} />
              </a>
            </div>
          </div>

          {/* Link Columns */}
          <div className="flex" style={{ gap: '80px' }}>
            {Object.entries(footerLinks).map(([category, links]) => (
              <div key={category} className="flex flex-col" style={{ gap: '16px' }}>
                <h4 className="font-semibold text-white" style={{ fontSize: '14px' }}>{category}</h4>
                {links.map((link) => (
                  <a
                    key={link}
                    href="#"
                    className="text-[#64748B] hover:text-white transition-colors"
                    style={{ fontSize: '14px' }}
                  >
                    {link}
                  </a>
                ))}
              </div>
            ))}
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="flex items-center justify-between w-full" style={{ paddingTop: '24px', borderTop: '1px solid #1E293B' }}>
          <span className="text-[#475569]" style={{ fontSize: '13px' }}>
            &copy; 2026 SwarmMarket. All rights reserved.
          </span>
          <div className="flex" style={{ gap: '24px' }}>
            <a href="#" className="text-[#475569] hover:text-white transition-colors" style={{ fontSize: '13px' }}>
              Privacy Policy
            </a>
            <a href="#" className="text-[#475569] hover:text-white transition-colors" style={{ fontSize: '13px' }}>
              Terms of Service
            </a>
            <a href="#" className="text-[#475569] hover:text-white transition-colors" style={{ fontSize: '13px' }}>
              Security
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
