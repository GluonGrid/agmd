import React from 'react';

export const ComparisonSection: React.FC = () => {
  return (
    <section className="py-24 border-y border-border-color bg-background-mantle/30">
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center mb-16">
          <h2 className="text-4xl font-bold tracking-tight mb-4 text-text-main">The Solution</h2>
          <p className="text-text-sub max-w-2xl mx-auto text-lg">
            Separate what you edit from what the AI reads. agmd keeps your source clean and your output complete.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Before: Manual */}
          <div className="flex flex-col group cursor-default">
            <div className="flex items-center gap-2 mb-4 px-2">
              <span className="size-2 rounded-full bg-accent-red shadow-[0_0_8px_rgba(231,130,132,0.5)]"></span>
              <span className="text-sm font-bold text-text-sub uppercase tracking-widest">The Problem: Manual Files</span>
            </div>
            
            <div className="flex-1 rounded-xl border border-surface bg-background-crust overflow-hidden shadow-xl glow-hover-red">
              <div className="bg-background-mantle px-4 py-2 border-b border-surface flex items-center justify-between">
                <span className="text-xs font-mono text-text-sub">AI_RULES.md</span>
                <div className="flex gap-1.5">
                  <div className="size-2.5 rounded-full bg-surface"></div>
                  <div className="size-2.5 rounded-full bg-surface"></div>
                  <div className="size-2.5 rounded-full bg-surface"></div>
                </div>
              </div>
              
              <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto custom-scrollbar">
                <div className="text-text-sub mb-4"># 500 lines of duplicated text...</div>
                
                <div className="text-accent-red">## Code Quality</div>
                <div className="text-text-main">1. Use strict mode</div>
                <div className="text-text-main">2. Avoid `any` type</div>
                <div className="text-text-main mb-4">3. Prefer interfaces...</div>
                
                <div className="text-accent-red">## Workflows</div>
                <div className="text-text-main">1. git commit -m "..."</div>
                <div className="text-text-main mb-4">2. git push origin main</div>
                
                <div className="text-accent-red font-bold opacity-80 select-none border-l-2 border-accent-red pl-2 mt-4 italic text-xs">
                  // Hard to scan, hard to update across 10 repos.
                </div>
              </div>
            </div>
          </div>

          {/* After: Directives */}
          <div className="flex flex-col group cursor-default">
            <div className="flex items-center gap-2 mb-4 px-2">
              <span className="size-2 rounded-full bg-primary shadow-[0_0_8px_rgba(140,170,238,0.5)]"></span>
              <span className="text-sm font-bold text-primary uppercase tracking-widest">The Solution: agmd</span>
            </div>
            
            <div className="flex-1 rounded-xl border border-primary/30 bg-background-crust overflow-hidden shadow-2xl relative glow-hover">
              
              <div className="bg-background-mantle px-4 py-2 border-b border-surface flex items-center justify-between relative z-10">
                <span className="text-xs font-mono text-primary">directives.md</span>
                <div className="flex gap-1.5">
                  <div className="size-2.5 rounded-full bg-primary/20"></div>
                  <div className="size-2.5 rounded-full bg-primary/20"></div>
                  <div className="size-2.5 rounded-full bg-primary/20"></div>
                </div>
              </div>
              
              <div className="p-6 font-mono text-sm leading-relaxed relative z-10 overflow-x-auto custom-scrollbar">
                <div className="text-text-sub"># Project Instructions</div>
                <br/>
                <div className="text-text-sub">## Code Quality</div>
                <div>
                  <span className="text-secondary">:::include</span> <span className="text-text-main">rule:typescript</span>
                </div>
                <div className="mb-4">
                  <span className="text-secondary">:::include</span> <span className="text-text-main">rule:no-any</span>
                </div>
                
                <div className="text-text-sub">## Workflows</div>
                <div>
                    <span className="text-secondary">:::list</span> <span className="text-text-main">workflow</span>
                </div>
                <div className="pl-4 text-text-sub">commit</div>
                <div className="pl-4 text-text-sub">deploy</div>
                <div className="text-secondary">:::end</div>
                
                <div className="text-primary font-bold select-none flex items-center gap-2 mt-6 pt-4 border-t border-surface/50 text-xs">
                  <span className="material-symbols-outlined text-sm">sync_alt</span>
                  agmd sync expands this to full content
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};