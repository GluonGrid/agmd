import React, { useEffect, useState } from 'react';
import { Header } from './components/Header';
import { Hero } from './components/Hero';
import { ComparisonSection } from './components/ComparisonSection';
import { FeaturesSection } from './components/FeaturesSection';
import { VideoSection } from './components/VideoSection';
import { CTASection } from './components/CTASection';
import { Footer } from './components/Footer';
import { DocsPage } from './components/DocsPage';

type Route = 'home' | 'docs';

const getRouteFromHash = (): Route => {
  if (typeof window === 'undefined') return 'home';
  const hash = window.location.hash.replace('#', '');
  if (hash.startsWith('/docs')) return 'docs';
  return 'home';
};

function App() {
  const [route, setRoute] = useState<Route>(getRouteFromHash());

  useEffect(() => {
    const onHashChange = () => setRoute(getRouteFromHash());
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, []);

  return (
    <div className="min-h-screen bg-background-base text-text-main flex flex-col font-sans">
      <Header currentRoute={route} />
      <main className="flex-grow">
        {route === 'docs' ? (
          <DocsPage />
        ) : (
          <>
            <Hero />
            <ComparisonSection />
            <FeaturesSection />
            <VideoSection />
            <CTASection />
          </>
        )}
      </main>
      <Footer />
    </div>
  );
}

export default App;
