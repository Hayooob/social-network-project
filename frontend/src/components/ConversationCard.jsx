import React from 'react';
import { Link } from 'react-router-dom';

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function ConversationCard({ conversation }) {
  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
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
      if (diffMins < 60) return `${diffMins}m`;
      if (diffHours < 24) return `${diffHours}h`;
      if (diffDays < 7) return `${diffDays}d`;
      return date.toLocaleDateString();
    } catch {
      return '';
    }
  };

  const truncateMessage = (msg, maxLength = 30) => {
    if (!msg) return '';
    return msg.length > maxLength ? msg.substring(0, maxLength) + '...' : msg;
  };

  return (
    <Link
      to={`/messages/${conversation.user_id}`}
      style={{ textDecoration: 'none', color: 'inherit' }}
    >
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        padding: '16px',
        borderBottom: '1px solid rgba(23, 3, 18, 0.08)',
        cursor: 'pointer',
        transition: 'background-color 0.2s ease'
      }}
        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(45, 81, 149, 0.05)'}
        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
      >
        {/* Avatar */}
        {(() => {
          const avatarSrc = conversation.avatar_url
            ? (conversation.avatar_url.startsWith('http')
              ? conversation.avatar_url
              : `${API_BASE}${conversation.avatar_url}`)
            : null;
          return (
            <div style={{
              width: '48px',
              height: '48px',
              borderRadius: '50%',
              backgroundColor: avatarSrc ? 'transparent' : 'var(--dusk-blue)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--white)',
              fontSize: '18px',
              fontWeight: 600,
              flexShrink: 0,
              position: 'relative'
            }}>
              {avatarSrc ? (
                <img
                  src={avatarSrc}
                  alt="avatar"
                  style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                />
              ) : (
                getInitial(conversation.user_name)
              )}

              {/* Unread indicator dot */}
              {conversation.unread_count > 0 && (
                <div style={{
                  position: 'absolute',
                  bottom: '0',
                  right: '0',
                  width: '14px',
                  height: '14px',
                  backgroundColor: 'var(--blush-rose)',
                  borderRadius: '50%',
                  border: '2px solid var(--white)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '8px',
                  color: 'var(--white)',
                  fontWeight: 700
                }}>
                  {conversation.unread_count > 9 ? '9+' : conversation.unread_count}
                </div>
              )}
            </div>
          );
        })()}

        {/* Content */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '4px'
          }}>
            <span style={{
              fontWeight: conversation.unread_count > 0 ? 700 : 500,
              fontSize: '14px',
              color: 'var(--coffee-bean)',
              fontFamily: "'Montserrat', sans-serif"
            }}>
              {conversation.user_name}
            </span>
            <span style={{
              fontSize: '11px',
              color: 'var(--jet-black)',
              opacity: 0.5,
              fontFamily: "'Montserrat', sans-serif"
            }}>
              {formatTime(conversation.last_message_at)}
            </span>
          </div>
          <p style={{
            margin: 0,
            fontSize: '13px',
            color: 'var(--jet-black)',
            opacity: conversation.unread_count > 0 ? 0.9 : 0.6,
            fontWeight: conversation.unread_count > 0 ? 500 : 400,
            fontFamily: "'Cormorant Garamond', serif",
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis'
          }}>
            {truncateMessage(conversation.last_message)}
          </p>
        </div>
      </div>
    </Link>
  );
}