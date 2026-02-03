import React from 'react';

export const Footer: React.FC = () => {
  return (
    <footer className="py-16 border-t border-border-color bg-background-base text-sm">
      <div className="max-w-7xl mx-auto px-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-12 mb-12">
          
          {/* Brand Column */}
          <div className="col-span-2">
            <div className="flex items-center gap-2.5 mb-6">
              <div className="size-6 bg-primary rounded flex items-center justify-center text-background-base">
                <span className="material-symbols-outlined text-sm font-bold">terminal</span>
              </div>
              <span className="text-lg font-bold text-text-main">agmd</span>
            </div>
            <p className="text-text-sub max-w-xs mb-6 font-medium leading-relaxed">
              The modern standard for managing AI agent behaviors via directives.
            </p>
            <div className="flex gap-4">
              <a href="#" className="text-text-sub hover:text-primary transition-colors">
                <span className="material-symbols-outlined">brand_family</span>
              </a>
              <a href="#" className="text-text-sub hover:text-primary transition-colors">
                <span className="material-symbols-outlined">rss_feed</span>
              </a>
            </div>
          </div>

          {/* Resources Column */}
          <div>
            <h4 className="font-bold mb-6 text-xs uppercase tracking-widest text-surface-1">Resources</h4>
            <ul className="space-y-4 text-text-sub">
              <li><a href="#" className="hover:text-primary transition-colors">Docs</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Registry</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Examples</a></li>
            </ul>
          </div>

          {/* Community Column */}
          <div>
            <h4 className="font-bold mb-6 text-xs uppercase tracking-widest text-surface-1">Community</h4>
            <ul className="space-y-4 text-text-sub">
              <li><a href="#" className="hover:text-primary transition-colors">GitHub</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Discord</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">X / Twitter</a></li>
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="pt-8 border-t border-border-color flex flex-col md:flex-row justify-between items-center gap-4 text-text-sub text-xs">
          <p>© 2024 agmd.sh • Built for AI Engineers</p>
          <div className="flex gap-6">
            <a href="#" className="hover:text-primary transition-colors">Privacy Policy</a>
            <a href="#" className="hover:text-primary transition-colors">Terms of Service</a>
          </div>
        </div>
      </div>
    </footer>
  );
};