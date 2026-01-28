import { useState, useEffect } from 'react';
import { get } from '../api/ftchclient'; // ADD THIS LINE

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = async () => {
    try {
      const data = await get('/api/me');
      setUser(data);
    } catch (err) {
      console.error('Failed to fetch user', err);
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  return { user, loading, refetch: fetchUser };
}