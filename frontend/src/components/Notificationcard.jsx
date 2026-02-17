import React from 'react';
import { Link } from 'react-router-dom';
import { UserPlus, UserCheck, Mail, Bell } from 'lucide-react';

export default function NotificationCard({ notification, onMarkRead }) {
  const getIcon = (type) => {
  switch (type) {
    case 'follow_request':
      return <UserPlus size={18} color="var(--white)" />;
    case 'follow_accept':
      return <UserCheck size={18} color="var(--white)" />;
    case 'new_message':
      return <Mail size={18} color="var(--white)" />;

    // Stage 7 (groups/events)
    case 'group_invite':
      return <UserPlus size={18} color="var(--white)" />;
    case 'group_join_request':
      return <UserPlus size={18} color="var(--white)" />;
    case 'group_event_created':
      return <Bell size={18} color="var(--white)" />;

    default:
      return <Bell size={18} color="var(--white)" />;
  }
};

const getMessage = (notification) => {
  const name = notification.from_user_name || 'Someone';
  switch (notification.type) {
    case 'follow_request':
      return `${name} sent you a follow request`;
    case 'follow_accept':
      return `${name} accepted your follow request`;
    case 'new_message':
      return `${name} ${notification.content || 'sent you a message'}`;

    // Stage 7
    case 'group_invite':
      return `${name} invited you to a group`;
    case 'group_join_request':
      return `${name} requested to join your group`;
    case 'group_event_created':
      return `${name} created a new event in your group`;

    default:
      return notification.content || 'New notification';
  }
};

const getLink = (notification) => {
  const ref = notification.reference_id;

  switch (notification.type) {
    case 'follow_request':
      return '/follow-requests';
    case 'follow_accept':
      return notification.from_user_id ? `/users/${notification.from_user_id}` : '/followers';
    case 'new_message':
      return notification.from_user_id ? `/messages/${notification.from_user_id}` : '/messages';

    // Stage 7 routing
    case 'group_invite':
      return '/groups'; // accept/decline happens there
    case 'group_join_request':
      return ref ? `/groups/${ref}` : '/groups'; // creator/admin sees pending requests
    case 'group_event_created':
      return ref ? `/groups/${ref}` : '/groups'; // goes into group -> events section

    default:
      return '#';
  }
};


  const formatTime = (timestamp) => {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffMins < 1) return 'Just now';
      if (diffMins < 60) return `${diffMins}m ago`;
      if (diffHours < 24) return `${diffHours}h ago`;
      if (diffDays < 7) return `${diffDays}d ago`;
      return date.toLocaleDateString();
    } catch {
      return '';
    }
  };

  const handleClick = () => {
    if (!notification.is_read && onMarkRead) {
      onMarkRead(notification.id);
    }
  };

  return (
    <Link 
      to={getLink(notification)} 
      onClick={handleClick}
      style={{ textDecoration: 'none', color: 'inherit' }}
    >
      <div style={{
        display: 'flex',
        alignItems: 'flex-start',
        gap: '12px',
        padding: '16px 20px',
        borderBottom: '1px solid rgba(23, 3, 18, 0.08)',
        backgroundColor: notification.is_read ? 'transparent' : 'rgba(45, 81, 149, 0.04)',
        cursor: 'pointer',
        transition: 'background-color 0.2s ease'
      }}
      onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(45, 81, 149, 0.08)'}
      onMouseLeave={(e) => e.currentTarget.style.backgroundColor = notification.is_read ? 'transparent' : 'rgba(45, 81, 149, 0.04)'}
      >
        {/* Icon */}
        <div style={{
          width: '40px',
          height: '40px',
          borderRadius: '50%',
          backgroundColor: notification.is_read ? 'var(--jet-black)' : 'var(--dusk-blue)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0
        }}>
          {getIcon(notification.type)}
        </div>

        {/* Content */}
        <div style={{ flex: 1 }}>
          <p style={{
            margin: 0,
            fontSize: '14px',
            color: 'var(--jet-black)',
            fontWeight: notification.is_read ? 400 : 600,
            fontFamily: "'Cormorant Garamond', serif",
            lineHeight: 1.4
          }}>
            {getMessage(notification)}
          </p>
          <span style={{
            fontSize: '11px',
            color: 'var(--jet-black)',
            opacity: 0.5,
            fontFamily: "'Montserrat', sans-serif"
          }}>
            {formatTime(notification.created_at)}
          </span>
        </div>

        {/* Unread dot */}
        {!notification.is_read && (
          <div style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            backgroundColor: 'var(--blush-rose)',
            flexShrink: 0,
            marginTop: '6px'
          }} />
        )}
      </div>
    </Link>
  );
}