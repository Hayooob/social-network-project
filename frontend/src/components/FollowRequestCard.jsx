import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { acceptFollowRequest, declineFollowRequest } from '../api/followers';

export default function FollowRequestCard({ request, onHandled }) {
  const [loading, setLoading] = useState(false);

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  const handleAccept = async () => {
    setLoading(true);
    try {
      await acceptFollowRequest(request.id);
      if (onHandled) onHandled(request.id);
    } catch (err) {
      console.error('Error accepting request:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDecline = async () => {
    setLoading(true);
    try {
      await declineFollowRequest(request.id);
      if (onHandled) onHandled(request.id);
    } catch (err) {
      console.error('Error declining request:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      padding: '16px 20px',
      borderBottom: '1px solid rgba(23, 3, 18, 0.08)'
    }}>
      {/* Avatar */}
      <Link to={`/users/${request.follower_id}`} style={{ textDecoration: 'none' }}>
        <div style={{
          width: '48px',
          height: '48px',
          borderRadius: '50%',
          backgroundColor: 'var(--dusk-blue)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--white)',
          fontSize: '18px',
          fontWeight: 600,
          flexShrink: 0
        }}>
          {getInitial(request.follower_name)}
        </div>
      </Link>

      {/* Name */}
      <div style={{ flex: 1 }}>
        <Link 
          to={`/users/${request.follower_id}`}
          style={{
            fontWeight: 600,
            fontSize: '14px',
            color: 'var(--coffee-bean)',
            fontFamily: "'Montserrat', sans-serif",
            textDecoration: 'none'
          }}
        >
          {request.follower_name}
        </Link>
        <p style={{
          margin: '4px 0 0 0',
          fontSize: '12px',
          color: 'var(--jet-black)',
          opacity: 0.6,
          fontFamily: "'Cormorant Garamond', serif"
        }}>
          wants to follow you
        </p>
      </div>

      {/* Buttons */}
      <div style={{ display: 'flex', gap: '8px' }}>
        <button
          className="btn btn-primary"
          onClick={handleAccept}
          disabled={loading}
          style={{ padding: '8px 16px' }}
        >
          Accept
        </button>
        <button
          className="btn btn-outline"
          onClick={handleDecline}
          disabled={loading}
          style={{ padding: '8px 16px' }}
        >
          Decline
        </button>
      </div>
    </div>
  );
}