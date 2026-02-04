import React, { useEffect, useState } from 'react';

const commands = [
  { text: 'agmd init', output: null, delay: 500 },
  { text: '', output: 'Created directives.md', delay: 800 },
  { text: 'agmd new rule:typescript', output: null, delay: 500 },
  { text: '', output: 'Created ~/.agmd/rule/typescript.md\nOpening editor...', delay: 1000 },
  { text: 'agmd list', output: null, delay: 500 },
  { text: '', output: 'Registry (~/.agmd):\n├── guide\n│   └── agmd.md\n├── profile\n│   └── svelte-kit.md\n└── rule\n    └── typescript.md', delay: 1500 },
  { text: 'agmd symlink', output: null, delay: 500 },
  { text: '', output: 'Symlinked AGENTS.md -> .cursorrules\nSymlinked AGENTS.md -> .windsurfrules', delay: 800 },
  { text: 'agmd sync', output: null, delay: 500 },
  { text: '', output: '✓ Parsed directives.md\n✓ Resolved 1 includes\n✓ Generated AGENTS.md (1.2kb)', delay: 2500 },
];

export const VideoSection: React.FC = () => {
  const [lines, setLines] = useState<{type: 'input' | 'output', content: string}[]>([]);
  const [currentCommandIndex, setCurrentCommandIndex] = useState(0);
  const [currentText, setCurrentText] = useState('');
  const resetAfterOutputIndex = 5;
  
  useEffect(() => {
    let timeout: ReturnType<typeof setTimeout>;

    const processStep = () => {
      const command = commands[currentCommandIndex];
      
      if (!command) {
        // Reset loop after delay
        timeout = setTimeout(() => {
            setLines([]);
            setCurrentCommandIndex(0);
            setCurrentText('');
        }, 5000);
        return;
      }

      if (command.text) {
        // Typing input
        if (currentText.length < command.text.length) {
          timeout = setTimeout(() => {
            setCurrentText(command.text.slice(0, currentText.length + 1));
          }, 30 + Math.random() * 40);
        } else {
          // Finished typing
          timeout = setTimeout(() => {
            setLines(prev => [...prev, { type: 'input', content: command.text }]);
            setCurrentText('');
            setCurrentCommandIndex(prev => prev + 1);
          }, 300);
        }
      } else {
        // Output only step
        if (command.output) {
            setLines(prev => [...prev, { type: 'output', content: command.output }]);
        }
        timeout = setTimeout(() => {
            if (currentCommandIndex === resetAfterOutputIndex) {
              setLines([]);
            }
            setCurrentCommandIndex(prev => prev + 1);
        }, command.delay);
      }
    };

    processStep();

    return () => clearTimeout(timeout);
  }, [currentCommandIndex, currentText]);

  return (
    <section className="py-24 bg-background-crust">
      <div className="max-w-5xl mx-auto px-6">
        <div className="text-center mb-12">
          <h2 className="text-2xl font-bold mb-2 text-text-main">See it in action</h2>
          <p className="text-text-sub mb-4">From a single directive to a fully expanded AI context in seconds.</p>
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-surface/50 border border-surface text-text-sub text-xs shadow-sm">
            <span className="material-symbols-outlined text-sm text-secondary">smart_toy</span>
            <span>Optimized for AI Agents: AI can run <code className="text-primary font-bold">agmd show guide:agmd</code> to read the docs.</span>
          </div>
        </div>

        <div className="rounded-xl overflow-hidden border border-surface shadow-2xl bg-background-base w-full max-w-3xl mx-auto terminal-glow h-[500px] flex flex-col">
          {/* Terminal Window Header */}
          <div className="bg-background-mantle px-4 py-3 flex items-center justify-between border-b border-surface shrink-0">
            <div className="flex gap-2">
              <div className="size-3 rounded-full bg-accent-red/80"></div>
              <div className="size-3 rounded-full bg-accent-peach/80"></div>
              <div className="size-3 rounded-full bg-accent-green/80"></div>
            </div>
            <div className="text-[10px] font-mono font-bold text-text-sub uppercase tracking-widest select-none">
              agmd-shell — zsh
            </div>
            <div className="w-10"></div>
          </div>

          {/* Terminal Content */}
          <div className="p-6 font-mono text-sm flex-1 overflow-y-auto custom-scrollbar bg-background-base flex flex-col justify-start">
            <div className="flex flex-col gap-3">
                {lines.map((line, i) => (
                    <div key={i} className="whitespace-pre-wrap break-words">
                        {line.type === 'input' ? (
                            <div className="flex gap-2">
                                <span className="text-accent-green">➜</span>
                                <span className="text-primary">~</span>
                                <span className="text-text-main">{line.content}</span>
                            </div>
                        ) : (
                            <div className={`text-text-sub leading-relaxed ${line.content.startsWith('Registry') ? '' : 'pl-6 border-l-2 border-surface-1 ml-1'}`}>
                                {line.content}
                            </div>
                        )}
                    </div>
                ))}
                
                {/* Current typing line */}
                {commands[currentCommandIndex]?.text && (
                    <div className="flex gap-2">
                        <span className="text-accent-green">➜</span>
                        <span className="text-primary">~</span>
                        <span className="text-text-main">
                            {currentText}
                            <span className="inline-block w-2 h-4 bg-text-sub align-middle ml-1 animate-blink"></span>
                        </span>
                    </div>
                )}
                 {/* Idle cursor when waiting */}
                {!commands[currentCommandIndex] && (
                   <div className="flex gap-2">
                        <span className="text-accent-green">➜</span>
                        <span className="text-primary">~</span>
                        <span className="text-text-main">
                            <span className="inline-block w-2 h-4 bg-text-sub align-middle ml-1 animate-blink"></span>
                        </span>
                    </div>
                )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};
