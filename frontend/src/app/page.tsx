'use client';

import {
  HeroSection,
  FeaturedListsSection,
  AnnouncementBlock,
} from '@/features/home';
import { Header, Footer } from '@/widgets';
import {
  DOMAIN_CONSTANTS,
  MOBILE_APP_REDIRECT_PATHS,
} from '@/constants/domains';

export default function Home() {
  const handleMobileRedirect = () => {
    const appScheme = `wishlistapp://${MOBILE_APP_REDIRECT_PATHS.HOME}`;
    const webFallback = DOMAIN_CONSTANTS.MOBILE_APP_BASE_URL;

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
