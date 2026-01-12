import { useState, useEffect } from 'react';
import { get } from '../api/ftchclient'; // use named import
export function useAuth() {
  const [user, setUser] = useState(null);

  useEffect(() => {
    async function fetchUser() {
      try {
        const data = await ftchclient('/api/me'); // assumes your backend has /api/me returning {id, name, email}
        setUser(data);
      } catch (err) {
        console.error('Failed to fetch user', err);
      }
    }
    fetchUser();
  }, []);

  return { user };
}
