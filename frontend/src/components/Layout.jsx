import React, { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { logout } from '../api/auth';
import { getPendingRequests } from '../api/followers';

export default function Layout({ children }) {
  const { user, setUser } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [requestCount, setRequestCount] = useState(0);

  const isActive = (path) => location.pathname === path;

  // Fetch pending request count when user is logged in
  useEffect(() => {
    async function fetchRequestCount() {
      if (user) {
        try {
          const requests = await getPendingRequests();
          setRequestCount(requests.length);
        } catch (err) {
          console.error('Error fetching request count:', err);
          setRequestCount(0);
        }
      }
    }
    fetchRequestCount();
  }, [user, location.pathname]); // Re-fetch when page changes

  async function handleLogout() {
    try {
      await logout();
    } catch (err) {
      console.error('Logout failed:', err);
    } finally {
      setUser(null);
      navigate('/login');
    }
  }

  return (
    <div className="app-container">
      {/* Background Elements */}
      <div className="gradient-bg"></div>
      <div className="noise-overlay"></div>
      
      {/* Floating Stars */}
      <div className="floating-stars">
        <span className="star star-1 star-float star-spin">✦</span>
        <span className="star star-2 star-float-reverse star-spin-reverse">✳</span>
        <span className="star star-3 star-float-slow star-pulse">✦</span>
        <span className="star star-4 star-float">✳</span>
        <span className="star star-5 star-float-reverse star-spin">✦</span>
      </div>

      {/* Navbar */}
      <header className="navbar">
        <Link to="/" className="navbar-logo-link">
          <div className="navbar-logo">
            <span className="navbar-logo-text">Social</span>
            <span className="navbar-logo-tagline">connect • share • discover</span>
          </div>
        </Link>

        <nav className="navbar-links">
          {!user && (
            <>
              <Link to="/login" className={`nav-link ${isActive('/login') ? 'active' : ''}`}>
                Login
                <span className="nav-link-arrow">↗</span>
              </Link>
              <Link to="/register" className={`nav-link ${isActive('/register') ? 'active' : ''}`}>
                Register
                <span className="nav-link-arrow">↗</span>
              </Link>
            </>
          )}

          {user && (
            <>
              <Link to="/feed" className={`nav-link ${isActive('/feed') ? 'active' : ''}`}>
                Feed
                <span className="nav-link-arrow">↗</span>
              </Link>
              <Link to="/profile" className={`nav-link ${isActive('/profile') ? 'active' : ''}`}>
                Profile
                <span className="nav-link-arrow">↗</span>
              </Link>
              <Link to="/followers" className={`nav-link ${isActive('/followers') ? 'active' : ''}`}>
                Followers
                <span className="nav-link-arrow">↗</span>
              </Link>
              <Link to="/follow-requests" className={`nav-link ${isActive('/follow-requests') ? 'active' : ''}`}>
                Requests
                {requestCount > 0 && (
                  <span className="notification-badge">{requestCount}</span>
                )}
                <span className="nav-link-arrow">↗</span>
              </Link>
              <button onClick={handleLogout} className="btn btn-secondary">
                Logout
              </button>
            </>
          )}
        </nav>
      </header>

      {/* Marquee Banner */}
      <div className="marquee-container">
        <div className="marquee-content marquee-track">
          {[...Array(16)].map((_, i) => (
            <span key={i} className={`marquee-text ${i % 2 === 0 ? 'marquee-text-medium' : 'marquee-text-light'}`}>
              feed • connect • share • discover • create
            </span>
          ))}
        </div>
      </div>

      {/* Page Content */}
      {children}

      {/* Footer */}
      <footer className="footer">
        <span className="footer-copyright">© 2026 SOCIAL TEAM @ REBOOT01 ˚ʚ♡ɞ˚</span>
        <div className="footer-links">
          <span className="footer-link">About</span>
          <span className="footer-link">Privacy</span>
          <span className="footer-link">Terms</span>
        </div>
        <span className="footer-star star-spin">✦</span>
      </footer>
    </div>
  );
}