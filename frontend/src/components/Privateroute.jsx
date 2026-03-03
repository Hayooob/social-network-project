import React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../VerifyAuth";

export default function Privateroute({ children }) {
  const { user, loading } = useAuth();

  if (loading) {
    // checking /api/me
    return (
      <div className="page">
        <p>Checking your session...</p>
      </div>
    );
  }

  if (!user) {
    // Not logged -> in send to login
    return <Navigate to="/login" replace />;
  }

  // Logged in -> show thef eed page
  return children;
}
