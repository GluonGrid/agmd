import React from 'react';

const features = [
  {
    icon: "library_books",
    color: "bg-primary/10 text-primary",
    title: "Personal Registry",
    description: "Store reusable content in `~/.agmd/`. Create your own types like `rule/`, `workflow/`, or `persona/` to match your needs."
  },
  {
    icon: "code",
    color: "bg-secondary/10 text-secondary",
    title: "Simple Syntax",
    description: "Learn just 3 directives: `:::include` to reuse single items, `:::list` for groups, and `:::new` to define inline content."
  },
  {
    icon: "move_up",
    color: "bg-accent-green/10 text-accent-green",
    title: "Easy Migration",
    description: "Got an existing messy `AI_RULES.md`? Run `agmd migrate` to organize it, or `agmd collect` to extract rules into your registry."
  }
];

const renderDescription = (text: string) => {
  const parts = text.split(/(`[^`]+`)/);
  return parts.map((part, i) => {
    if (part.startsWith('`') && part.endsWith('`')) {
      return (
        <code key={i} className="bg-surface px-1.5 py-0.5 rounded text-secondary text-sm font-bold mx-0.5">
          {part.slice(1, -1)}
        </code>
      );
    }
    return part;
  });
};

export const FeaturesSection: React.FC = () => {
  return (
    <section className="py-24 bg-background-base">
      <div className="max-w-7xl mx-auto px-6">
        <div className="mb-16">
          <h2 className="text-3xl font-bold tracking-tight mb-4 text-text-main">How It Works</h2>
          <p className="text-text-sub max-w-2xl text-lg">
            Designed for the "DRY" (Don't Repeat Yourself) principle, applied to AI system prompts.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {features.map((feature, idx) => (
            <div 
              key={idx}
              className="feature-card group p-8 rounded-xl border border-border-color bg-surface/50 transition-all duration-300 hover:-translate-y-1"
            >
              <div className={`size-12 rounded-lg ${feature.color} flex items-center justify-center mb-6 transition-transform group-hover:scale-110 shadow-lg`}>
                <span className="material-symbols-outlined">{feature.icon}</span>
              </div>
              <h3 className="text-xl font-bold mb-3 text-text-main">{feature.title}</h3>
              <p className="text-text-sub leading-relaxed">
                {renderDescription(feature.description)}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};