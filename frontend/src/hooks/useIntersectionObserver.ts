import { useEffect, useRef } from 'react';

interface UseIntersectionObserverOptions {
  onIntersect: () => void;
  enabled?: boolean;
  rootMargin?: string;
}

export function useIntersectionObserver({
  onIntersect,
  enabled = true,
  rootMargin = '200px',
}: UseIntersectionObserverOptions) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!enabled || !ref.current) return;
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) onIntersect();
      },
      { rootMargin },
    );
    obs.observe(ref.current);
    return () => obs.disconnect();
  }, [enabled, onIntersect, rootMargin]);
  return ref;
}
