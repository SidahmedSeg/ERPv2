import { useState, useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';

export function useCompanyCurrency() {
  const { accessToken } = useAuthStore();
  const [currency, setCurrency] = useState<string>('USD');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCurrency = async () => {
      if (!accessToken) {
        setLoading(false);
        return;
      }

      try {
        const response = await fetch('http://localhost:8080/api/settings/company', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = await response.json();
        if (data.success && data.data?.default_currency) {
          setCurrency(data.data.default_currency);
        }
      } catch (error) {
        console.error('Failed to fetch company currency:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchCurrency();
  }, [token]);

  return { currency, loading };
}
