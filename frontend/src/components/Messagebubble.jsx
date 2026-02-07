import React from 'react';

export default function MessageBubble({ message, isOwn }) {
  const formatTime = (timestamp) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: isOwn ? 'flex-end' : 'flex-start',
      marginBottom: '12px'
    }}>
      <div style={{
        maxWidth: '70%',
        padding: '12px 16px',
        backgroundColor: isOwn ? 'var(--dusk-blue)' : 'var(--white)',
        color: isOwn ? 'var(--white)' : 'var(--jet-black)',
        borderRadius: isOwn ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.06)'
      }}>
        <p style={{
          margin: 0,
          fontSize: '14px',
          lineHeight: 1.5,
          fontFamily: "'Cormorant Garamond', serif"
        }}>
          {message.content}
        </p>
        <span style={{
          display: 'block',
          marginTop: '6px',
          fontSize: '10px',
          opacity: 0.7,
          textAlign: 'right',
          fontFamily: "'Montserrat', sans-serif"
        }}>
          {formatTime(message.created_at)}
        </span>
      </div>
    </div>
  );
}