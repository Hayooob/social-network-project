import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { logout } from '../api/auth';

// basically this wraps the pages with the bg and  floating stars, navbar, marquee, and the footerrr
export default function Layout({ children }) {
  const { user, setUser } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (err) {
      console.error('Logout failed:', err);
    } finally {
      setUser(null);
      navigate('/login');
    }
  };

  const isActive = (path) => {
    return location.pathname === path;
  };

  return (
    <div className="layout-container">
      {/* Noise texture overlay */}
      <div className="noise-overlay"></div>

      {/* Floating decorative stars */}
      <div className="floating-star floating-star-1 star-float star-spin">✦</div>
      <div className="floating-star floating-star-2 star-float-reverse star-spin-reverse">✳</div>
      <div className="floating-star floating-star-3 star-float-slow star-pulse">✦</div>
      <div className="floating-star floating-star-4 star-float star-spin-reverse">✳</div>
      <div className="floating-star floating-star-5 star-float-slow">✦</div>

      {/* Navbar */}
      <header className="navbar">
        <Link to="/" className="navbar-logo">
          <span className="navbar-logo-text">Social</span>
          <span className="navbar-logo-tagline">connect • share • discover</span>
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
                <span className="notification-badge">3</span>
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
        <span className="footer-copyright">© 2026 SOCIAL</span>
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