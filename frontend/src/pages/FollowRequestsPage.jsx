import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { getPendingRequests, acceptFollowRequest, declineFollowRequest } from '../api/followers';

export default function FollowRequestsPage() {
  const { user } = useAuth();
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(null);

  const fetchRequests = async () => {
    setLoading(true);
    try {
      const data = await getPendingRequests();
      setRequests(data);
    } catch (err) {
      console.error('Error loading follow requests:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequests();
  }, []);

  const handleAccept = async (userId) => {
    setActionLoading(`accept-${userId}`);
    try {
      await acceptFollowRequest(userId);
      setRequests(prev => prev.filter(r => r.follower_id !== userId));
    } catch (err) {
      console.error('Error accepting request:', err);
      alert('Failed to accept request.');
    } finally {
      setActionLoading(null);
    }
  };

  const handleDecline = async (userId) => {
    setActionLoading(`decline-${userId}`);
    try {
      await declineFollowRequest(userId);
      setRequests(prev => prev.filter(r => r.follower_id !== userId));
    } catch (err) {
      console.error('Error declining request:', err);
      alert('Failed to decline request.');
    } finally {
      setActionLoading(null);
    }
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  const getAvatarClass = (index) => {
    const classes = ['bg-blue', 'bg-rose', 'bg-dark'];
    return classes[index % 3];
  };

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
    <div className="main-content">
      <div className="requests-page-grid">
        {/* Header Card */}
        <div className="requests-header-section">
          <div className="card card-shadow-right">
            <div className="card-header">
              <span className="card-header-title">Follow Requests</span>
              <span className="card-header-star star-spin">✦</span>
            </div>
            
            <div className="requests-header-content">
              <div className="requests-count-box">
                <span className="requests-count-number">{requests.length}</span>
                <span className="requests-count-label">
                  {requests.length === 1 ? 'Pending Request' : 'Pending Requests'}
                </span>
              </div>
              {requests.length > 0 && (
                <p className="requests-helper-text">
                  These people want to follow you. Accept to let them see your posts.
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="requests-content-area">
          {/* Section Header */}
          <div className="section-header">
            <span className="section-title">Pending Requests</span>
            <div className="section-line"></div>
            <span className="section-star star-spin-reverse">✳</span>
          </div>

          {loading ? (
            <div className="card">
              <div className="card-body text-center">
                <p>Loading requests...</p>
              </div>
            </div>
          ) : requests.length === 0 ? (
            <div className="card">
              <div className="card-body text-center">
                <div style={{ padding: '48px 24px' }}>
                  <span className="star-float text-rose" style={{ fontSize: '48px', display: 'block', marginBottom: '24px' }}>✦</span>
                  <p style={{ marginBottom: '8px', fontFamily: "'Cormorant Garamond', serif", fontSize: '24px' }}>All caught up!</p>
                  <p style={{ fontSize: '14px', opacity: 0.6, marginBottom: '24px' }}>You don't have any pending follow requests.</p>
                  <Link to="/feed" className="btn btn-primary">
                    Back to Feed
                    <span>↗</span>
                  </Link>
                </div>
              </div>
            </div>
          ) : (
            <div className="requests-list">
              {requests.map((request, index) => (
                <div key={request.id || request.follower_id} className="card request-card">
                  <div className="request-card-accent"></div>
                  <div className="request-card-content">
                    <div className="request-info">
                      <div className={`avatar avatar-medium ${getAvatarClass(index)}`}>
                        {getInitial(request.follower_name)}
                      </div>
                      <div className="request-details">
                        <span className="request-name">{request.follower_name || 'Unknown User'}</span>
                        <span className="request-handle">@{request.follower_name?.toLowerCase().replace(' ', '') || 'user'}</span>
                        <span className="request-time">Requested {formatDate(request.created_at)}</span>
                      </div>
                    </div>
                    <div className="request-actions">
                      <button 
                        className="btn btn-primary"
                        onClick={() => handleAccept(request.follower_id)}
                        disabled={actionLoading === `accept-${request.follower_id}`}
                      >
                        {actionLoading === `accept-${request.follower_id}` ? 'Accepting...' : 'Accept'}
                        <span>✓</span>
                      </button>
                      <button 
                        className="btn btn-outline btn-danger"
                        onClick={() => handleDecline(request.follower_id)}
                        disabled={actionLoading === `decline-${request.follower_id}`}
                      >
                        {actionLoading === `decline-${request.follower_id}` ? 'Declining...' : 'Decline'}
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Sidebar */}
        <aside className="requests-sidebar">
          <div className="quick-links">
            <h3 className="quick-links-title">
              <span className="star-spin text-rose" style={{ fontSize: '12px' }}>✳</span>
              Quick Links
            </h3>
            <Link to="/feed" className="quick-link-item">
              <span className="quick-link-text">Back to Feed</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/profile" className="quick-link-item">
              <span className="quick-link-text">My Profile</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/followers" className="quick-link-item">
              <span className="quick-link-text">My Followers</span>
              <span className="quick-link-arrow">→</span>
            </Link>
          </div>

          <div className="info-box">
            <h4 className="info-box-title">
              <span style={{ marginRight: '8px' }}>ℹ</span>
              About Requests
            </h4>
            <p className="info-box-text">
              When your profile is private, people need your approval to follow you. 
              Accepting a request lets them see your posts and activity.
            </p>
          </div>

          <div className="quote-box">
            <div className="quote-star star-float star-spin">✳</div>
            <p className="quote-text">
              "The greatest gift is not being afraid to question."
            </p>
            <p style={{ fontSize: '11px', opacity: 0.6, marginTop: '8px', fontStyle: 'italic' }}>— Ruby Dee</p>
          </div>

          <div style={{ marginTop: '24px', display: 'flex', justifyContent: 'center', gap: '16px' }}>
            <span className="star-float text-rose" style={{ opacity: 0.4, fontSize: '20px' }}>✦</span>
            <span className="star-float-reverse text-blue" style={{ opacity: 0.3, fontSize: '16px' }}>✳</span>
            <span className="star-float text-dark" style={{ opacity: 0.2, fontSize: '24px' }}>✦</span>
          </div>
        </aside>
      </div>
    </div>
  );
}