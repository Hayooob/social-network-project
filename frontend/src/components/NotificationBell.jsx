import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getUnreadNotificationCount } from '../api/notifications';

export default function NotificationBell() {
  const [count, setCount] = useState(0);

  const fetchCount = async () => {
    const unreadCount = await getUnreadNotificationCount();
    setCount(unreadCount);
  };

  useEffect(() => {
    fetchCount();
    
    // Poll every 30 seconds for new notifications
    const interval = setInterval(fetchCount, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <Link 
      to="/notifications" 
      style={{ 
        position: 'relative', 
        textDecoration: 'none',
        display: 'flex',
        alignItems: 'center',
        padding: '8px'
      }}
    >
      {/* Bell Icon */}
      <span style={{ fontSize: '20px' }}>🔔</span>
      
      {/* Badge */}
      {count > 0 && (
        <span style={{
          position: 'absolute',
          top: '2px',
          right: '2px',
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
          padding: '0 4px'
        }}>
          {count > 99 ? '99+' : count}
        </span>
      )}
    </Link>
  );
}