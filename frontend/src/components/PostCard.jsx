import React from 'react';
import { Link } from 'react-router-dom';

export default function PostCard({ post }) {
  const formatDate = (timestamp) => {
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
      return 'Recently';
    }
  };

  return (
    <div style={{
      padding: '20px',
      borderBottom: '1px solid rgba(23, 3, 18, 0.08)',
      transition: 'background-color 0.2s ease'
    }}>
      {/* Author Header */}
      <div style={{ 
        display: 'flex', 
        alignItems: 'center', 
        gap: '12px',
        marginBottom: '12px'
      }}>
        {/* Avatar */}
        <Link 
          to={`/users/${post.user_id}`}
          style={{
            width: '40px',
            height: '40px',
            borderRadius: '50%',
            backgroundColor: 'var(--dusk-blue)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--white)',
            fontSize: '16px',
            fontWeight: 600,
            textDecoration: 'none',
            flexShrink: 0
          }}
        >
          {(post.author_name || 'U').charAt(0).toUpperCase()}
        </Link>

        {/* Name and Time */}
        <div style={{ flex: 1 }}>
          <Link 
            to={`/users/${post.user_id}`}
            style={{
              fontWeight: 600,
              fontSize: '14px',
              color: 'var(--coffee-bean)',
              fontFamily: "'Montserrat', sans-serif",
              textDecoration: 'none',
              display: 'block'
            }}
            onMouseEnter={(e) => e.target.style.color = 'var(--dusk-blue)'}
            onMouseLeave={(e) => e.target.style.color = 'var(--coffee-bean)'}
          >
            {post.author_name || 'Unknown'}
          </Link>
          <span style={{
            fontSize: '12px',
            color: 'var(--jet-black)',
            opacity: 0.5,
            fontFamily: "'Montserrat', sans-serif"
          }}>
            {formatDate(post.created_at)}
          </span>
        </div>
      </div>

      {/* Post Content */}
      <p style={{
        margin: 0,
        fontSize: '15px',
        lineHeight: 1.6,
        color: 'var(--jet-black)',
        fontFamily: "'Cormorant Garamond', serif",
        whiteSpace: 'pre-wrap'
      }}>
        {post.content}
      </p>

      {/* Post Image (from their changes) */}
      {post.image_path && (
        <img
          src={
            post.image_path.startsWith("http")
              ? post.image_path
              : `http://localhost:8080${post.image_path}`
          }
          alt="post"
          style={{ marginTop: 12, maxWidth: "100%", borderRadius: 12 }}
        />
      )}
    </div>
  );
}