import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { getMyPosts, createPost } from '../api/posts';
import { getFollowCounts } from '../api/followers';

export default function ProfilePage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [posting, setPosting] = useState(false);
  const [activeTab, setActiveTab] = useState('posts');

  const fetchData = async () => {
    setLoading(true);
    try {
      const [postsData, countsData] = await Promise.all([
        getMyPosts(),
        getFollowCounts()
      ]);
      setPosts(postsData);
      setCounts(countsData);
    } catch (err) {
      console.error('Error loading profile data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user) {
      fetchData();
    }
  }, [user]);

  const handlePostSubmit = async (e) => {
    e.preventDefault();
    if (!content.trim()) return;

    setPosting(true);
    try {
      await createPost(content.trim());
      setContent('');
      fetchData();
    } catch (err) {
      console.error('Error creating post:', err);
      alert('Failed to create post.');
    } finally {
      setPosting(false);
    }
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
      if (diffMins < 60) return `${diffMins} min ago`;
      if (diffHours < 24) return `${diffHours} hours ago`;
      if (diffDays < 7) return `${diffDays} days ago`;
      return date.toLocaleDateString();
    } catch {
      return 'Recently';
    }
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
  };

  const getAccentClass = (index) => {
    const classes = ['post-accent', 'post-accent-rose', 'post-accent-dark'];
    return classes[index % 3];
  };

  if (!user) {
    return (
      <div className="main-content page-center">
        <p>Loading user info...</p>
      </div>
    );
  }

  return (
    <div className="main-content">
      <div className="profile-page-grid">
        {/* Profile Header Section */}
        <div className="profile-header-section">
          <div className="card card-shadow-right">
            <div className="card-header">
              <span className="card-header-title">My Profile</span>
              <span className="card-header-star star-spin">✦</span>
            </div>
            
            <div className="profile-header-content">
              <div className="profile-header-left">
                <div className="avatar avatar-large">
                  {getInitial(user.full_name)}
                  <div className="avatar-status"></div>
                </div>
              </div>
              
              <div className="profile-header-info">
                <h1 className="profile-display-name">{user.full_name}</h1>
                <p className="profile-display-handle">@{user.email?.split('@')[0]}</p>
                <p className="profile-display-email">{user.email}</p>
                
                <div className="profile-header-stats">
                  <Link to="/followers" className="profile-stat-box">
                    <span className="profile-stat-number">{counts.followers}</span>
                    <span className="profile-stat-label">Followers</span>
                  </Link>
                  <Link to="/followers" className="profile-stat-box">
                    <span className="profile-stat-number">{counts.following}</span>
                    <span className="profile-stat-label">Following</span>
                  </Link>
                  <div className="profile-stat-box">
                    <span className="profile-stat-number">{posts.length}</span>
                    <span className="profile-stat-label">Posts</span>
                  </div>
                </div>
              </div>

              <div className="profile-header-actions">
                <Link to="/follow-requests" className="btn btn-outline">
                  Follow Requests
                  <span>↗</span>
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content Area */}
        <div className="profile-content-area">
          {/* Tabs */}
          <div className="profile-tabs">
            <button 
              className={`profile-tab ${activeTab === 'posts' ? 'active' : ''}`}
              onClick={() => setActiveTab('posts')}
            >
              <span className="star-spin" style={{ fontSize: '10px', marginRight: '8px' }}>✦</span>
              My Posts
            </button>
            <button 
              className={`profile-tab ${activeTab === 'create' ? 'active' : ''}`}
              onClick={() => setActiveTab('create')}
            >
              <span className="star-spin" style={{ fontSize: '10px', marginRight: '8px' }}>✳</span>
              Create Post
            </button>
          </div>

          {/* Create Post Tab */}
          {activeTab === 'create' && (
            <div className="card" style={{ marginBottom: '32px' }}>
              <div className="card-header">
                <span className="card-header-title">New Post</span>
                <span className="card-header-star star-spin">✦</span>
              </div>
              <div className="card-body">
                <form onSubmit={handlePostSubmit}>
                  <textarea
                    className="form-input form-textarea"
                    placeholder="What's on your mind?"
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    disabled={posting}
                    rows={5}
                  />
                  <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '16px' }}>
                    <button type="submit" className="btn btn-primary" disabled={posting}>
                      {posting ? 'Posting...' : 'Post'}
                      <span>↗</span>
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          {/* Posts Tab */}
          {activeTab === 'posts' && (
            <>
              <div className="section-header">
                <span className="section-title">My Posts</span>
                <div className="section-line"></div>
                <span className="section-star star-spin-reverse">✳</span>
              </div>

              {loading ? (
                <div className="card">
                  <div className="card-body text-center">
                    <p>Loading posts...</p>
                  </div>
                </div>
              ) : posts.length === 0 ? (
                <div className="card">
                  <div className="card-body text-center">
                    <p style={{ marginBottom: '16px' }}>You haven't posted anything yet.</p>
                    <button 
                      className="btn btn-primary"
                      onClick={() => setActiveTab('create')}
                    >
                      Create Your First Post
                      <span>↗</span>
                    </button>
                  </div>
                </div>
              ) : (
                posts.map((post, index) => (
                  <div key={post.id} className="card post-card">
                    <div className={getAccentClass(index)}></div>
                    <div className="post-content">
                      <div className="post-header">
                        <div className="post-author-info">
                          <div className={`avatar avatar-small ${index % 3 === 1 ? 'bg-rose' : index % 3 === 2 ? 'bg-dark' : 'bg-blue'}`}>
                            {getInitial(post.author_name)}
                          </div>
                          <div>
                            <div className="post-author-name">{post.author_name || 'Unknown'}</div>
                            <div className="post-author-handle">@{post.author_name?.toLowerCase().replace(' ', '') || 'user'}</div>
                          </div>
                        </div>
                        <span className="post-time">{formatDate(post.created_at)}</span>
                      </div>
                      <p className="post-text">{post.content}</p>
                      <div className="post-actions">
                        <span className="post-action">♥ Like</span>
                        <span className="post-action">↩ Reply</span>
                        <span className="post-action">⋯</span>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </>
          )}
        </div>

        {/* Sidebar */}
        <aside className="profile-sidebar">
          <div className="quick-links">
            <h3 className="quick-links-title">
              <span className="star-spin text-rose" style={{ fontSize: '12px' }}>✳</span>
              Quick Links
            </h3>
            <Link to="/feed" className="quick-link-item">
              <span className="quick-link-text">Back to Feed</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/followers" className="quick-link-item">
              <span className="quick-link-text">My Followers</span>
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
              "Be yourself; everyone else is already taken."
            </p>
            <p style={{ fontSize: '11px', opacity: 0.6, marginTop: '8px', fontStyle: 'italic' }}>— Oscar Wilde</p>
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