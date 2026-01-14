import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getFeed, createPost } from '../api/posts';
import { getFollowCounts } from '../api/followers';
import { useAuth } from '../VerifyAuth';

export default function FeedPage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [posting, setPosting] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    try {
      const feedData = await getFeed();
      const countsData = await getFollowCounts();
      setPosts(feedData);
      setCounts(countsData);
    } catch (err) {
      console.error('Error loading feed:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

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

  return (
    <div className="main-content">
      <div className="page-grid">
        {/* Left Sidebar - Profile Card */}
        <aside>
          <div className="card card-shadow-right profile-card">
            <span className="card-star star-spin">✦</span>
            <div className="avatar">
              {user ? getInitial(user.full_name) : '?'}
              <div className="avatar-status"></div>
            </div>
            <h2 className="profile-name">{user?.full_name || 'Loading...'}</h2>
            <p className="profile-handle">@{user?.email?.split('@')[0] || 'user'}</p>
            <div className="divider"></div>
            <div className="profile-stats">
              <Link to="/followers" style={{ textDecoration: 'none' }}>
                <div className="stat-number">{counts.followers}</div>
                <div className="stat-label">Followers</div>
              </Link>
              <Link to="/followers" style={{ textDecoration: 'none' }}>
                <div className="stat-number">{counts.following}</div>
                <div className="stat-label">Following</div>
              </Link>
            </div>
          </div>

          <div className="quick-links">
            <h3 className="quick-links-title">
              <span className="star-spin text-rose" style={{ fontSize: '12px' }}>✳</span>
              Quick Links
            </h3>
            <Link to="/profile" className="quick-link-item">
              <span className="quick-link-text">Edit Profile</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/follow-requests" className="quick-link-item">
              <span className="quick-link-text">Follow Requests</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <div className="quick-link-item">
              <span className="quick-link-text">Settings</span>
              <span className="quick-link-arrow">→</span>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main>
          {/* Create Post */}
          <div className="card" style={{ marginBottom: '32px' }}>
            <div className="card-header">
              <span className="card-header-title">Create Post</span>
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

          {/* Section Header */}
          <div className="section-header">
            <span className="section-title">Latest Posts</span>
            <div className="section-line"></div>
            <span className="section-star star-spin-reverse">✳</span>
          </div>

          {/* Posts */}
          {loading ? (
            <p>Loading posts...</p>
          ) : posts.length === 0 ? (
            <div className="card">
              <div className="card-body text-center">
                <p>No posts yet. Be the first to share something!</p>
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
        </main>

        {/* Right Sidebar */}
        <aside>
          <div className="card card-shadow-left">
            <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(23, 3, 18, 0.1)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span className="section-title">Suggested</span>
              <span className="text-rose star-float-slow" style={{ fontSize: '12px' }}>✦</span>
            </div>
            <div className="suggestion-item">
              <div className="suggestion-info">
                <div className="avatar avatar-tiny bg-blue">J</div>
                <span className="suggestion-name">Jane Doe</span>
              </div>
              <button className="btn btn-outline">Follow</button>
            </div>
            <div className="suggestion-item">
              <div className="suggestion-info">
                <div className="avatar avatar-tiny bg-rose">T</div>
                <span className="suggestion-name">Tom Brown</span>
              </div>
              <button className="btn btn-outline">Follow</button>
            </div>
            <div className="suggestion-item">
              <div className="suggestion-info">
                <div className="avatar avatar-tiny bg-dark">E</div>
                <span className="suggestion-name">Emma Wilson</span>
              </div>
              <button className="btn btn-outline">Follow</button>
            </div>
          </div>

          <div className="quote-box">
            <div className="quote-star star-float star-spin">✳</div>
            <p className="quote-text">
              "Connection is the energy that exists between people when they feel seen, heard, and valued."
            </p>
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