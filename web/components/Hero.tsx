import React, { useState } from 'react';

export const Hero: React.FC = () => {
  const [copied, setCopied] = useState(false);
  const installCmd = 'curl -fsSL https://gluongrid.dev/agmd/install.sh | bash';

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className="relative overflow-hidden pt-24 pb-32 md:pt-40 md:pb-52 bg-background-base">
      {/* Background Dot Pattern */}
      <div 
        className="absolute inset-0 z-0 opacity-10 pointer-events-none" 
        style={{ 
          backgroundImage: 'radial-gradient(circle at 2px 2px, #c6d0f5 1px, transparent 0)', 
          backgroundSize: '48px 48px' 
        }}
      ></div>

      <div className="relative z-10 max-w-5xl mx-auto px-6 text-center flex flex-col items-center">
        {/* Beta Badge */}
        <div className="flex justify-center w-full mb-8">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-bold tracking-wider uppercase shadow-[0_0_10px_-2px_rgba(140,170,238,0.3)]">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
            </span>
            v0.2.0 Available
          </div>
        </div>

        {/* Main Title */}
        <div className="relative inline-block mb-6">
          <div
            className="absolute -inset-6 md:-inset-10 rounded-full blur-3xl opacity-40 pointer-events-none"
            style={{
              background:
                'radial-gradient(circle, rgba(214,224,255,0.55) 0%, rgba(140,170,238,0.25) 45%, rgba(48,52,70,0) 70%)',
            }}
          ></div>
          <h1 className="relative text-7xl md:text-9xl font-black tracking-tighter bg-gradient-to-b from-text-main to-text-sub bg-clip-text text-transparent font-mono select-none">
            agmd
          </h1>
        </div>

        {/* Subtitle */}
        <p className="text-xl md:text-2xl text-text-sub max-w-2xl mx-auto mb-12 font-medium leading-relaxed">
          Your AI instructions, organized. The CLI that solves "copy-paste hell" for <code className="bg-surface px-1.5 py-0.5 rounded text-secondary text-sm font-bold">AGENTS.md</code> and <code className="bg-surface px-1.5 py-0.5 rounded text-secondary text-sm font-bold">AI_RULES.md</code>.
        </p>

        {/* Install Terminal */}
        <div className="max-w-3xl mx-auto terminal-glow mb-10">
          <div className="bg-background-mantle rounded-2xl border border-surface-1 p-2 flex items-center shadow-lg transition-transform hover:scale-[1.01] duration-300">
            <div className="flex-1 flex items-center gap-4 px-5 py-4 font-mono text-sm md:text-base text-left overflow-x-auto custom-scrollbar">
              <span className="text-secondary select-none">$</span>
              <code className="text-text-main flex flex-nowrap gap-x-2 whitespace-nowrap">
                <span className="text-primary">curl</span>
                <span className="text-text-sub">-fsSL</span>
                <span>https://gluongrid.dev/agmd/install.sh</span>
                <span className="text-text-sub">|</span>
                <span className="text-secondary">bash</span>
              </code>
            </div>
            <button 
              onClick={handleCopy}
              className="bg-primary hover:bg-primary/90 text-background-crust font-bold py-4 px-6 rounded-xl transition-all flex items-center gap-2 group shrink-0 active:scale-95"
            >
              <span className="material-symbols-outlined text-xl">
                {copied ? 'check' : 'content_copy'}
              </span>
              <span className="hidden sm:inline">{copied ? 'Copied' : 'Copy'}</span>
            </button>
          </div>
        </div>
        
        <p className="text-xs text-text-sub font-mono tracking-wide">
            Runs on <span className="text-text-main">macOS</span>, <span className="text-text-main">Linux</span>, and <span className="text-text-main">WSL</span>
        </p>
      </div>
    </section>
  );
};
