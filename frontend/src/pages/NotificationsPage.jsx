import React, { useEffect, useState } from 'react';
import { getNotifications, markNotificationRead, markAllNotificationsRead } from '../api/notifications';
import NotificationCard from '../components/NotificationCard';
import { Bell } from 'lucide-react';

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchNotifications = async () => {
    setLoading(true);
    try {
      const data = await getNotifications();
      setNotifications(data);
    } catch (err) {
      console.error('Error fetching notifications:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNotifications();
  }, []);

  const handleMarkRead = async (notifId) => {
    try {
      await markNotificationRead(notifId);
      setNotifications(prev => 
        prev.map(n => n.id === notifId ? { ...n, is_read: true } : n)
      );
    } catch (err) {
      console.error('Error marking notification as read:', err);
    }
  };

  const handleMarkAllRead = async () => {
    try {
      await markAllNotificationsRead();
      setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
    } catch (err) {
      console.error('Error marking all notifications as read:', err);
    }
  };

  const unreadCount = notifications.filter(n => !n.is_read).length;

  return (
    <div className="main-content">
      <div style={{ maxWidth: '600px', margin: '0 auto' }}>
        {/* Header Card */}
        <div className="card" style={{ marginBottom: '24px' }}>
          <div className="card-header">
            <span className="card-header-title">Notifications</span>
            <span className="card-header-star star-spin">✦</span>
          </div>
          <div className="card-body" style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            padding: '16px 24px'
          }}>
            <span style={{
              fontSize: '13px',
              color: 'var(--jet-black)',
              opacity: 0.7,
              fontFamily: "'Cormorant Garamond', serif"
            }}>
              {unreadCount > 0 
                ? `You have ${unreadCount} unread notification${unreadCount > 1 ? 's' : ''}`
                : 'All caught up!'
              }
            </span>
            {unreadCount > 0 && (
              <button 
                className="btn btn-outline"
                onClick={handleMarkAllRead}
                style={{ padding: '6px 12px' }}
              >
                Mark all read
              </button>
            )}
          </div>
        </div>

        {/* Notifications List */}
        <div className="card">
          {loading ? (
            <div style={{ padding: '40px', textAlign: 'center' }}>
              <p style={{ opacity: 0.6, fontFamily: "'Cormorant Garamond', serif" }}>
                Loading notifications...
              </p>
            </div>
          ) : notifications.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center' }}>
              <Bell size={48} color="var(--dusk-blue)" strokeWidth={1.5} style={{ marginBottom: '16px' }} />
              <p style={{ 
                opacity: 0.6, 
                fontFamily: "'Cormorant Garamond', serif",
                fontSize: '16px'
              }}>
                No notifications yet
              </p>
            </div>
          ) : (
            notifications.map(notif => (
              <NotificationCard 
                key={notif.id} 
                notification={notif}
                onMarkRead={handleMarkRead}
              />
            ))
          )}
        </div>

        {/* Decorative elements */}
        <div style={{ marginTop: '32px', display: 'flex', justifyContent: 'center', gap: '16px' }}>
          <span className="star-float text-rose" style={{ opacity: 0.4, fontSize: '20px' }}>✦</span>
          <span className="star-float-reverse text-blue" style={{ opacity: 0.3, fontSize: '16px' }}>✳</span>
          <span className="star-float text-dark" style={{ opacity: 0.2, fontSize: '24px' }}>✦</span>
        </div>
      </div>
    </div>
  );
}