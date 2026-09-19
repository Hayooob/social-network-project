import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getMyFollowers, getMyFollowing, unfollowUser } from '../api/followers';

export default function FollowersPage() {
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('followers');
  const [actionLoading, setActionLoading] = useState(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [followersData, followingData] = await Promise.all([
        getMyFollowers(),
        getMyFollowing()
      ]);
      setFollowers(followersData);
      setFollowing(followingData);
    } catch (err) {
      console.error('Error loading followers data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleUnfollow = async (userId) => {
    setActionLoading(userId);
    try {
      await unfollowUser(userId);
      setFollowing(prev => prev.filter(f => f.following_id !== userId));
    } catch (err) {
      console.error('Error unfollowing user:', err);
      alert('Failed to unfollow user.');
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

  return (
    <div className="main-content">
      <div className="followers-page-grid">
        {/* Header Card */}
        <div className="followers-header-section">
          <div className="card card-shadow-right">
            <div className="card-header">
              <span className="card-header-title">My Network</span>
              <span className="card-header-star star-spin">✦</span>
            </div>
            
            <div className="followers-header-content">
              <div className="followers-stats-row">
                <button 
                  className={`followers-stat-button ${activeTab === 'followers' ? 'active' : ''}`}
                  onClick={() => setActiveTab('followers')}
                >
                  <span className="followers-stat-number">{followers.length}</span>
                  <span className="followers-stat-label">Followers</span>
                </button>
                <div className="followers-stat-divider"></div>
                <button 
                  className={`followers-stat-button ${activeTab === 'following' ? 'active' : ''}`}
                  onClick={() => setActiveTab('following')}
                >
                  <span className="followers-stat-number">{following.length}</span>
                  <span className="followers-stat-label">Following</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="followers-content-area">
          {/* Section Header */}
          <div className="section-header">
            <span className="section-title">
              {activeTab === 'followers' ? 'People Following You' : 'People You Follow'}
            </span>
            <div className="section-line"></div>
            <span className="section-star star-spin-reverse">✳</span>
          </div>

          {loading ? (
            <div className="card">
              <div className="card-body text-center">
                <p>Loading...</p>
              </div>
            </div>
          ) : activeTab === 'followers' ? (
            /* Followers List */
            followers.length === 0 ? (
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ padding: '24px 0' }}>
                    <span className="star-float text-rose" style={{ fontSize: '32px', display: 'block', marginBottom: '16px' }}>✦</span>
                    <p style={{ marginBottom: '8px', fontFamily: "'Cormorant Garamond', serif", fontSize: '18px' }}>No followers yet</p>
                    <p style={{ fontSize: '13px', opacity: 0.6 }}>Share your profile to get more followers!</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="followers-list">
                {followers.map((follow, index) => (
                  <div key={follow.id || follow.follower_id} className="card follower-card">
                    <div className="follower-card-accent"></div>
                    <div className="follower-card-content">
                      <div className="follower-info">
                        <div className={`avatar avatar-small ${getAvatarClass(index)}`}>
                          {getInitial(follow.follower_name)}
                        </div>
                        <div className="follower-details">
                          <span className="follower-name">{follow.follower_name || 'Unknown User'}</span>
                          <span className="follower-handle">@{follow.follower_name?.toLowerCase().replace(' ', '') || 'user'}</span>
                        </div>
                      </div>
                      <div className="follower-actions">
                        <span className="follower-date">
                          Followed you
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )
          ) : (
            /* Following List */
            following.length === 0 ? (
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ padding: '24px 0' }}>
                    <span className="star-float text-blue" style={{ fontSize: '32px', display: 'block', marginBottom: '16px' }}>✳</span>
                    <p style={{ marginBottom: '8px', fontFamily: "'Cormorant Garamond', serif", fontSize: '18px' }}>Not following anyone yet</p>
                    <p style={{ fontSize: '13px', opacity: 0.6, marginBottom: '16px' }}>Discover people to follow in your feed!</p>
                    <Link to="/feed" className="btn btn-primary">
                      Go to Feed
                      <span>↗</span>
                    </Link>
                  </div>
                </div>
              </div>
            ) : (
              <div className="followers-list">
                {following.map((follow, index) => (
                  <div key={follow.id || follow.following_id} className="card follower-card">
                    <div className="follower-card-accent follower-card-accent-blue"></div>
                    <div className="follower-card-content">
                      <div className="follower-info">
                        <div className={`avatar avatar-small ${getAvatarClass(index)}`}>
                          {getInitial(follow.following_name)}
                        </div>
                        <div className="follower-details">
                          <span className="follower-name">{follow.following_name || 'Unknown User'}</span>
                          <span className="follower-handle">@{follow.following_name?.toLowerCase().replace(' ', '') || 'user'}</span>
                        </div>
                      </div>
                      <div className="follower-actions">
                        <button 
                          className="btn btn-outline btn-danger"
                          onClick={() => handleUnfollow(follow.following_id)}
                          disabled={actionLoading === follow.following_id}
                        >
                          {actionLoading === follow.following_id ? 'Unfollowing...' : 'Unfollow'}
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )
          )}
        </div>

        {/* Sidebar */}
        <aside className="followers-sidebar">
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
            <Link to="/follow-requests" className="quick-link-item">
              <span className="quick-link-text">Follow Requests</span>
              <span className="quick-link-arrow">→</span>
            </Link>
          </div>

          <div className="quote-box">
            <div className="quote-star star-float star-spin">✳</div>
            <p className="quote-text">
              "Alone we can do so little; together we can do so much."
            </p>
            <p style={{ fontSize: '11px', opacity: 0.6, marginTop: '8px', fontStyle: 'italic' }}>— Helen Keller</p>
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