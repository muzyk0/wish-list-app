import type { ReactNode } from 'react';
import './wishlist-theme.css';

interface PublicWishlistLayoutProps {
  children: ReactNode;
}

export default function PublicWishlistLayout({
  children,
}: PublicWishlistLayoutProps) {
  return <>{children}</>;
}
