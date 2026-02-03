import React from 'react';

export const CTASection: React.FC = () => {
  return (
    <section className="py-24 bg-primary relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden opacity-10 pointer-events-none">
         <div className="absolute -top-[50%] -left-[10%] w-[120%] h-[200%] bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20"></div>
      </div>

      <div className="max-w-7xl mx-auto px-6 text-center relative z-10">
        <h2 className="text-background-crust text-4xl md:text-5xl font-black mb-8 tracking-tighter">
          Ready to sync your agents?
        </h2>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <button className="w-full sm:w-auto px-10 py-4 bg-background-crust text-text-main font-bold rounded-lg hover:opacity-90 hover:scale-105 transition-all flex items-center justify-center gap-2 shadow-xl border border-transparent">
            Get Started
            <span className="material-symbols-outlined">arrow_forward</span>
          </button>
          <button className="w-full sm:w-auto px-10 py-4 border-2 border-background-crust text-background-crust font-bold rounded-lg hover:bg-background-crust/10 transition-colors">
            View GitHub
          </button>
        </div>
      </div>
    </section>
  );
};