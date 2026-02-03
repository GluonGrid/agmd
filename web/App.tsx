import React from 'react';
import { Header } from './components/Header';
import { Hero } from './components/Hero';
import { ComparisonSection } from './components/ComparisonSection';
import { FeaturesSection } from './components/FeaturesSection';
import { VideoSection } from './components/VideoSection';
import { CTASection } from './components/CTASection';
import { Footer } from './components/Footer';

function App() {
  return (
    <div className="min-h-screen bg-background-base text-text-main flex flex-col font-sans">
      <Header />
      <main className="flex-grow">
        <Hero />
        <ComparisonSection />
        <FeaturesSection />
        <VideoSection />
        <CTASection />
      </main>
      <Footer />
    </div>
  );
}

export default App;