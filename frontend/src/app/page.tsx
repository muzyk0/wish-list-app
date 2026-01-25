'use client';

import {
  HeroSection,
  FeaturedListsSection,
  AnnouncementBlock,
} from '@/features/home';
import { Header, Footer } from '@/widgets';

export default function Home() {
  const handleMobileRedirect = () => {
    const appScheme = 'wishlistapp://home';
    const webFallback = 'https://lk.domain.com';

    window.location.href = appScheme;

    setTimeout(() => {
      if (!document.hidden && document.visibilityState !== 'hidden') {
        window.location.href = webFallback;
      }
    }, 1500);
  };

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main>
        <HeroSection onMobileRedirect={handleMobileRedirect} />
        <FeaturedListsSection />
        <AnnouncementBlock />
      </main>

      <Footer />
    </div>
  );
}
