'use client';

import { DM_Sans, Playfair_Display } from 'next/font/google';
import type { ReactNode } from 'react';
import { Header } from '@/widgets';
import './wishlist-theme.css';

const playfairDisplay = Playfair_Display({
  variable: '--font-display',
  subsets: ['latin'],
  display: 'swap',
});

const dmSans = DM_Sans({
  variable: '--font-body',
  subsets: ['latin'],
  display: 'swap',
});

interface PublicWishlistLayoutProps {
  children: ReactNode;
}

export default function PublicWishlistLayout({
  children,
}: PublicWishlistLayoutProps) {
  return (
    <div className={`${playfairDisplay.variable} ${dmSans.variable} wl-page`}>
      <Header />
      {children}
    </div>
  );
}
