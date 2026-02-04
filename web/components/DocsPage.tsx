import React from 'react';

const SectionTitle: React.FC<{ title: string; subtitle?: string }> = ({ title, subtitle }) => (
  <div className="flex items-start justify-between gap-6 flex-wrap">
    <div>
      <h2 className="text-3xl md:text-4xl font-bold tracking-tight text-text-main">{title}</h2>
      {subtitle ? <p className="text-text-sub mt-3 max-w-2xl">{subtitle}</p> : null}
    </div>
  </div>
);

const CodeBlock: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <pre className="bg-background-mantle border border-surface-1 rounded-2xl p-5 overflow-x-auto text-sm text-text-main font-mono">
    <code>{children}</code>
  </pre>
);

const CommandExample: React.FC<{ title: string; command: string; output: string; input?: React.ReactNode }> = ({
  title,
  command,
  output,
  input,
}) => {
  return (
    <div className="border-b border-surface/70 pb-6">
      <div className="text-xs font-mono text-text-sub uppercase tracking-[0.2em] mb-2">{title}</div>
      <div className="font-mono text-sm text-text-main overflow-x-auto mb-4">
        <span className="text-secondary">$</span>{' '}
        <span className="text-primary font-semibold">agmd</span>{' '}
        <span>{command.replace(/^agmd(\s+)?/, '')}</span>
      </div>
      {input ? (
        <div className="grid md:grid-cols-2 gap-4">
          <div className="bg-background-base/60 border border-surface-1 rounded-xl p-4 font-mono text-sm text-text-sub whitespace-pre-wrap">
            {input}
          </div>
          <div className="bg-background-base/60 border border-surface-1 rounded-xl p-4 font-mono text-sm text-text-sub whitespace-pre-wrap">
            {output}
          </div>
        </div>
      ) : (
        <div className="bg-background-base/60 border border-surface-1 rounded-xl p-4 font-mono text-sm text-text-sub whitespace-pre-wrap">
          {output}
        </div>
      )}
    </div>
  );
};

const DirectivePanel: React.FC = () => (
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
      <br />
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
);

const Card: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div className="bg-surface/40 border border-border-color rounded-2xl p-6 glow-hover">
    <h3 className="text-lg font-semibold text-text-main mb-3">{title}</h3>
    <div className="text-text-sub text-sm leading-relaxed">{children}</div>
  </div>
);

export const DocsPage: React.FC = () => {
  return (
    <div className="bg-background-base">
      <section className="relative overflow-hidden pt-20 pb-16 md:pt-28 md:pb-20">
        <div
          className="absolute inset-0 z-0 opacity-10 pointer-events-none"
          style={{
            backgroundImage: 'radial-gradient(circle at 2px 2px, #c6d0f5 1px, transparent 0)',
            backgroundSize: '56px 56px',
          }}
        ></div>
        <div className="relative z-10 max-w-6xl mx-auto px-6">
          <div className="flex flex-col gap-6">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-bold tracking-wider uppercase shadow-[0_0_10px_-2px_rgba(140,170,238,0.3)] w-fit">
              Documentation
            </div>
            <h1 className="text-5xl md:text-7xl font-black tracking-tight text-text-main">
              agmd docs
            </h1>
            <p className="text-lg md:text-xl text-text-sub max-w-3xl">
              Install, initialize your registry, and generate clean instruction files with simple directives.
            </p>
          </div>
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-6 pb-16">
        <SectionTitle
          title="Installation"
          subtitle="Get agmd on your machine in seconds."
        />
        <div className="mt-6">
          <CodeBlock>{`# Quick install
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash

# From source
go install github.com/GluonGrid/agmd@latest`}</CodeBlock>
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-6 pb-20">
        <SectionTitle
          title="Quick Start"
          subtitle="Install, initialize your registry, then generate AGENTS.md in any project."
        />
        <div className="mt-6">
          <CodeBlock>{`# Install
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash

# Initialize your personal registry
agmd setup

# In any project
agmd init
agmd edit
agmd sync`}</CodeBlock>
        </div>
      </section>

      <section id="registry" className="max-w-6xl mx-auto px-6 pb-20 scroll-mt-28">
        <SectionTitle
          title="Directive Examples"
          subtitle="Keep directives compact while keeping output full."
        />
        <div className="mt-8">
          <DirectivePanel />
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-6 pb-20">
        <SectionTitle
          title="Commands"
          subtitle="Core commands for building your registry and generating instruction files."
        />
        <div className="mt-8 grid md:grid-cols-2 xl:grid-cols-3 gap-6 text-sm text-text-sub">
          <Card title="agmd setup">Initialize your <code className="text-text-main">~/.agmd/</code> registry.</Card>
          <Card title="agmd init [profile:name]">Create <code className="text-text-main">directives.md</code> in the current project.</Card>
          <Card title="agmd sync">Generate <code className="text-text-main">AGENTS.md</code> from directives.</Card>
          <Card title="agmd edit [type:name]">Edit directives or a registry item.</Card>
          <Card title="agmd new type:name">Create a new registry item.</Card>
          <Card title="agmd show type:name">Display item content for assistants.</Card>
          <Card title="agmd list [type]">List registry items.</Card>
          <Card title="agmd promote">Promote <code className="text-text-main">:::new</code> blocks into your registry.</Card>
          <Card title="agmd migrate &lt;file&gt;">Convert a raw CLAUDE.md/AGENTS.md into directives.</Card>
          <Card title="agmd collect [-f file]">Collect rules from an agmd project into your registry.</Card>
        </div>
        <div className="mt-10 space-y-6">
          <CommandExample
            title="agmd setup"
            command="agmd setup"
            output={`-> Creating: /tmp/agmd-home/.agmd\n\nok Registry ready!\n\nNext steps:\n  agmd init              # Create directives.md in a project\n  agmd new type:name     # Create a reusable item\n  agmd list              # See your registry`}
          />
          <CommandExample
            title="agmd init"
            command="agmd init"
            output={`-> Initializing agmd project...\nok Using default profile\n  Default directives.md template with basic structure\n-> Creating directives.md...\nok Created directives.md\n\nok Project initialized successfully!\n\nCreated:\n  - directives.md - Source file with directives (edit this)\n\nNext steps:\n  - Edit directives.md to add directives\n  - Run 'agmd add rule <name>' to add rules to directives.md\n  - Run 'agmd sync' to create AGENTS.md for AI agents\n  - Run 'agmd new rule <name>' to create custom rules`}
          />
          <CommandExample
            title="agmd new"
            command="agmd new rule:typescript --no-editor"
            input={
              <>
                <span className="text-text-sub"># directives.md</span>
                {'\n'}
                <span className="text-secondary">:::new</span>{' '}
                <span className="text-text-main">rule:typescript</span>
                {'\n'}
                <span className="text-text-main"># TypeScript Standards</span>
                {'\n'}
                <span>Use strict mode. Avoid any.</span>
                {'\n'}
                <span className="text-secondary">:::end</span>
              </>
            }
            output={`ok Created rule:typescript\n-> /tmp/agmd-home/.agmd/rule/typescript.md`}
          />
          <CommandExample
            title="agmd edit"
            command={`agmd edit rule:typescript --content "# TypeScript Standards\\nUse strict mode. Avoid any."`}
            output={`ok Updated rule:typescript`}
          />
          <CommandExample
            title="agmd show"
            command="agmd show rule:typescript"
            output={`# TypeScript Standards\nUse strict mode. Avoid any.`}
          />
          <CommandExample
            title="agmd show --raw"
            command="agmd show rule:typescript --raw"
            output={`---\nname: typescript\ndescription: \"\"\n---\n\n# TypeScript Standards\nUse strict mode. Avoid any.`}
          />
          <CommandExample
            title="agmd list"
            command="agmd list"
            output={`/tmp/agmd-home/.agmd\n\nguide/ (1)\n  agmd - Guide for AI assistants on how to use agmd directives\n\nprofile/ (2)\n  default - Default directives.md template with basic structure\n  starter\n\nrule/ (2)\n  custom-auth\n  typescript`}
          />
          <CommandExample
            title="agmd list rule"
            command="agmd list rule"
            output={`/tmp/agmd-home/.agmd\n\nguide/ (1)\n  agmd - Guide for AI assistants on how to use agmd directives\n\nprofile/ (2)\n  default - Default directives.md template with basic structure\n  starter\n\nrule/ (2)\n  custom-auth\n  typescript`}
          />
          <CommandExample
            title="agmd promote"
            command="agmd promote --all"
            input={
              <>
                <span className="text-text-sub"># directives.md</span>
                {'\n'}
                <span className="text-secondary">:::new</span>{' '}
                <span className="text-text-main">rule:custom-auth</span>
                {'\n'}
                <span className="text-text-main"># Auth Rules</span>
                {'\n'}
                <span>Always rotate tokens weekly.</span>
                {'\n'}
                <span className="text-secondary">:::end</span>
              </>
            }
            output={`-> Found 1 :::new blocks to promote\n\n-> Promoting rule:custom-auth\nok Extracted content from :::new block\nok Created rule at /tmp/agmd-home/.agmd/rule/custom-auth.md\nok Replaced :::new block with :::include rule:custom-auth\n\nok Complete! 1/1 items promoted to registry and directives.md updated.\ni Run 'agmd sync' to update AGENTS.md`}
          />
          <CommandExample
            title="agmd sync"
            command="agmd sync"
            output={`-> Generating AGENTS.md from directives.md...\n-> Loading registry...\n-> Parsing and expanding directives...\n\nok Generated AGENTS.md successfully!\ni Source: directives.md -> Output: AGENTS.md`}
          />
          <CommandExample
            title="agmd migrate"
            command="agmd migrate CLAUDE.md --force"
            output={`-> Migrating CLAUDE.md...\nok Backup created: CLAUDE.md.backup\nok Created directives.md\n\n-> Opening editor...\n\nok Migration complete!\n\nNext steps:\n  1. Wrap sections with :::new markers\n  2. Run 'agmd promote' to save to registry\n  3. Run 'agmd sync' to generate AGENTS.md`}
          />
          <CommandExample
            title="agmd collect"
            command="agmd collect"
            output={`-> Analyzing directives.md and AGENTS.md...\ni No items found to collect`}
          />
          <CommandExample
            title="agmd new profile"
            command={`agmd new profile:starter --content "# Project Instructions"`}
            output={`ok Created profile:starter\n\n-> Use in new project: agmd init profile:starter`}
          />
          <CommandExample
            title="agmd init profile"
            command="agmd init profile:starter"
            output={`-> Initializing agmd project with profile 'starter'...\nok Using profile: starter\n-> Creating directives.md...\nok Created directives.md\n\nok Project initialized successfully!\n\nCreated:\n  - directives.md - Source file with directives (edit this)\n\nNext steps:\n  - Edit directives.md to add directives\n  - Run 'agmd add rule <name>' to add rules to directives.md\n  - Run 'agmd sync' to create AGENTS.md for AI agents\n  - Run 'agmd new rule <name>' to create custom rules`}
          />
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-6 pb-20">
        <SectionTitle
          title="Migrating Existing Projects"
          subtitle="Two paths depending on whether a project already uses agmd."
        />
        <div className="mt-8 grid lg:grid-cols-2 gap-6">
          <Card title="Migrate: raw instructions">
            Use when a project has a freeform CLAUDE.md or AGENTS.md. This creates <code className="text-text-main">directives.md</code> and
            lets you wrap reusable parts in <code className="text-text-main">:::new</code> blocks.
          </Card>
          <Card title="Collect: agmd-compatible">
            Use when a project already has <code className="text-text-main">directives.md</code>. Collects referenced items into your
            registry for reuse elsewhere.
          </Card>
        </div>
        <div className="mt-6 grid lg:grid-cols-2 gap-6">
          <CodeBlock>{`# Migrate raw file
agmd migrate CLAUDE.md
agmd migrate CLAUDE.md --force`}</CodeBlock>
          <CodeBlock>{`# Collect from agmd project
agmd collect
agmd collect -f CLAUDE.md`}</CodeBlock>
        </div>
      </section>

      <section className="max-w-6xl mx-auto px-6 pb-20">
        <SectionTitle title="Profiles" subtitle="Save a project as a reusable template and bootstrap instantly." />
        <div className="mt-6 grid lg:grid-cols-2 gap-6">
          <CodeBlock>{`# Save a template
agmd new profile:svelte-kit

# Use it
agmd init profile:svelte-kit`}</CodeBlock>
          <div className="bg-background-mantle border border-surface-1 rounded-2xl p-6 text-text-sub text-sm leading-relaxed">
            Profiles are complete <code className="text-text-main">directives.md</code> templates, perfect for stack-specific standards.
          </div>
        </div>
      </section>
    </div>
  );
};
