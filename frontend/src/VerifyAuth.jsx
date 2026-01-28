import React, {
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";
import { getCurrentUser } from "./api/auth";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);      // null = logged out
  const [loading, setLoading] = useState(true); 
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;

    async function initAuth() {
      try {
        const currentUser = await getCurrentUser(); // user or null
        if (!cancelled) {
          setUser(currentUser);
        }
      } catch (err) {
        if (!cancelled) {
          console.error("Failed to fetch current user", err);
          setError(err);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    initAuth();

    return () => {
      cancelled = true;
    };
  }, []);

  const value = {
    user,
    setUser,
    loading,
    error,
  };

  return (
    <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
