'use client';

import { useRouter, usePathname } from 'next/navigation';
import { useState, useEffect, useRef } from 'react';

export function useOptimisticNavigation() {
  const router = useRouter();
  const pathname = usePathname();
  const [pendingPath, setPendingPath] = useState<string | null>(null);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Reset pendingPath when pathname changes (navigation completes)
  useEffect(() => {
    if (pendingPath && pathname === pendingPath) {
      setPendingPath(null);
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    }
  }, [pathname, pendingPath]);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const navigate = (href: string) => {
    setPendingPath(href);
    router.push(href);

    // Fallback: reset after 2 seconds if pathname didn't change
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      setPendingPath(null);
      timeoutRef.current = null;
    }, 2000);
  };

  const reset = () => {
    setPendingPath(null);
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  };

  return {
    navigate,
    isPending: !!pendingPath,
    pendingPath,
    reset,
  };
}
