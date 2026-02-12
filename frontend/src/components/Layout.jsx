import React, { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { useWebSocket } from '../WebSocketContext';
import { logout } from '../api/auth';
import { getPendingRequests } from '../api/followers';
import { getUnreadNotificationCount } from '../api/notifications';
import { getUnreadMessageCount } from '../api/messages';
import { Bell } from 'lucide-react';

export default function Layout({ children }) {
  const { user, setUser } = useAuth();
  const { badgeRefreshTrigger, isConnected } = useWebSocket();
  const location = useLocation();
  const navigate = useNavigate();
  const [requestCount, setRequestCount] = useState(0);
  const [notificationCount, setNotificationCount] = useState(0);
  const [messageCount, setMessageCount] = useState(0);

  const isActive = (path) => location.pathname === path || location.pathname.startsWith(path + '/');

  // Fetch counts when user is logged in, page changes, badge refresh is triggered, or WebSocket connects
  useEffect(() => {
    async function fetchCounts() {
      if (user) {
        console.log('Layout: Fetching badge counts... trigger:', badgeRefreshTrigger, 'connected:', isConnected);
        try {
          const [requests, notifs, msgs] = await Promise.all([
            getPendingRequests(),
            getUnreadNotificationCount(),
            getUnreadMessageCount()
          ]);
          console.log('Layout: Badge counts received', { requests: requests.length, notifs, msgs });
          setRequestCount(requests.length);
          setNotificationCount(notifs);
          setMessageCount(msgs);
        } catch (err) {
          console.error('Error fetching counts:', err);
          setRequestCount(0);
          setNotificationCount(0);
          setMessageCount(0);
        }
      }
    }
    fetchCounts();
  }, [user, location.pathname, badgeRefreshTrigger, isConnected]);

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
              
              {/* Messages with badge */}
              <Link 
                to="/messages" 
                className={`nav-link ${isActive('/messages') ? 'active' : ''}`}
                style={{ position: 'relative' }}
              >
                Messages
                {messageCount > 0 && (
                  <span style={{
                    position: 'absolute',
                    top: '-8px',
                    right: '-14px',
                    backgroundColor: 'var(--blush-rose)',
                    color: 'var(--white)',
                    fontSize: '10px',
                    fontWeight: 700,
                    minWidth: '18px',
                    height: '18px',
                    borderRadius: '9px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: "'Montserrat', sans-serif",
                    padding: '0 5px'
                  }}>
                    {messageCount > 99 ? '99+' : messageCount}
                  </span>
                )}
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
              <Link 
                to="/follow-requests" 
                className={`nav-link ${isActive('/follow-requests') ? 'active' : ''}`}
                style={{ position: 'relative' }}
              >
                Requests
                {requestCount > 0 && (
                  <span style={{
                    position: 'absolute',
                    top: '-8px',
                    right: '-14px',
                    backgroundColor: 'var(--blush-rose)',
                    color: 'var(--white)',
                    fontSize: '10px',
                    fontWeight: 700,
                    minWidth: '18px',
                    height: '18px',
                    borderRadius: '9px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: "'Montserrat', sans-serif",
                    padding: '0 5px'
                  }}>
                    {requestCount > 99 ? '99+' : requestCount}
                  </span>
                )}
                <span className="nav-link-arrow">↗</span>
              </Link>

              {/* Notification Bell Icon */}
              <Link 
                to="/notifications" 
                style={{ 
                  position: 'relative',
                  display: 'flex',
                  alignItems: 'center',
                  padding: '8px',
                  marginLeft: '8px'
                }}
              >
                <Bell 
                  size={20} 
                  color={isActive('/notifications') ? 'var(--blush-rose)' : 'var(--coffee-bean)'} 
                  strokeWidth={2}
                />
                {notificationCount > 0 && (
                  <span style={{
                    position: 'absolute',
                    top: '0',
                    right: '0',
                    backgroundColor: 'var(--blush-rose)',
                    color: 'var(--white)',
                    fontSize: '9px',
                    fontWeight: 700,
                    minWidth: '16px',
                    height: '16px',
                    borderRadius: '8px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontFamily: "'Montserrat', sans-serif",
                    padding: '0 4px'
                  }}>
                    {notificationCount > 99 ? '99+' : notificationCount}
                  </span>
                )}
              </Link>
              <Link to="/groups" className={`nav-link ${isActive('/groups') ? 'active' : ''}`}>
                Groups
                <span className="nav-link-arrow">↗</span>
              </Link>

              <Link to="/events" className={`nav-link ${isActive('/events') ? 'active' : ''}`}>
                Events
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