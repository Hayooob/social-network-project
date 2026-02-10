import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { getMyPosts, listComments, createComment, toggleLike } from '../api/posts';
import { getFollowCounts } from '../api/followers';
import { togglePrivacy } from '../api/users';

export default function ProfilePage() {
  const { user, setUser } = useAuth();
  const [posts, setPosts] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);
  const [privacyLoading, setPrivacyLoading] = useState(false);
  const [error, setError] = useState('');
  const [openCommentsPostId, setOpenCommentsPostId] = useState(null);
const [commentsByPostId, setCommentsByPostId] = useState({});
const [commentDraftByPostId, setCommentDraftByPostId] = useState({});
const [likesByPostId, setLikesByPostId] = useState({});


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
    setError('Failed to load profile data');
  } finally {
    setLoading(false);
  }
};

  useEffect(() => {
    fetchData();
  }, []);

  const handleTogglePrivacy = async () => {
    setPrivacyLoading(true);
    try {
      const res = await togglePrivacy();
      setUser({ ...user, is_private: res.is_private });
    } catch (err) {
      console.error('Error toggling privacy:', err);
    } finally {
      setPrivacyLoading(false);
    }
  };

  const getInitial = (name) => {
    return name ? name.charAt(0).toUpperCase() : '?';
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

  const getAccentClass = (index) => {
    const classes = ['post-accent', 'post-accent-rose', 'post-accent-dark'];
    return classes[index % 3];
  };
const loadComments = async (postId) => {
  const data = await listComments(postId);
  setCommentsByPostId((prev) => ({ ...prev, [postId]: data || [] }));
};

const onToggleComments = async (postId) => {
  if (openCommentsPostId === postId) {
    setOpenCommentsPostId(null);
    return;
  }
  setOpenCommentsPostId(postId);
  if (!commentsByPostId[postId]) {
    try {
      await loadComments(postId);
    } catch (e) {
      console.error("loadComments error:", e);
      alert(e.message || "Failed to load comments");
    }
  }
};

const onSubmitComment = async (postId) => {
  const text = (commentDraftByPostId[postId] || "").trim();
  if (!text) return;

  try {
    await createComment(postId, text);
    setCommentDraftByPostId((prev) => ({ ...prev, [postId]: "" }));
    await loadComments(postId);
  } catch (e) {
    console.error("createComment error:", e);
    alert(e.message || "Failed to create comment");
  }
};

const onToggleLike = async (postId) => {
  try {
    const res = await toggleLike(postId);

    setLikesByPostId((prev) => ({ ...prev, [postId]: res }));

    setPosts((prev) =>
      prev.map((p) =>
        p.id === postId ? { ...p, like_count: res.like_count } : p
      )
    );
  } catch (e) {
    console.error("toggleLike error:", e);
    alert(e.message || "Failed to like");
  }
};

  if (!user) return null;

  return (
    <div className="main-content">
      <div className="page-grid-two-col">
        {/* Main Content - Profile Info & Posts */}
        <main>
          {/* Profile Header Card */}
          <div className="card card-shadow-right" style={{ marginBottom: '32px' }}>
            <div className="card-header">
              <span className="card-header-title">My Profile</span>
              <span className="card-header-star star-spin">✦</span>
            </div>
            <div className="card-body">
              <div style={{ display: 'flex', gap: '24px', alignItems: 'flex-start' }}>
                {/* Avatar */}
                <div className="avatar" style={{ flexShrink: 0 }}>
                  {getInitial(user.full_name)}
                  <div className="avatar-status"></div>
                </div>

                {/* Info */}
                <div style={{ flex: 1 }}>
                  <h2 className="profile-name" style={{ fontSize: '18px', marginBottom: '8px' }}>
                    {user.full_name}
                  </h2>
                  <p style={{ fontSize: '14px', color: 'var(--blush-rose)', marginBottom: '4px', fontFamily: "'Cormorant Garamond', serif" }}>
                    @{user.email?.split('@')[0] || 'user'}
                  </p>
                  <p style={{ fontSize: '13px', opacity: 0.7, fontFamily: "'Cormorant Garamond', serif" }}>
                    {user.email}
                  </p>

                  {/* Privacy Toggle */}
                  <div style={{ marginTop: '16px', display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <span style={{ fontSize: '11px', letterSpacing: '2px', textTransform: 'uppercase', opacity: 0.7 }}>
                      Profile: {user.is_private ? 'Private' : 'Public'}
                    </span>
                    <button 
                      className="btn btn-outline" 
                      onClick={handleTogglePrivacy}
                      disabled={privacyLoading}
                      style={{ padding: '6px 12px' }}
                    >
                      {privacyLoading ? 'Updating...' : 'Toggle'}
                      <span>↗</span>
                    </button>
                  </div>
                </div>

                {/* Stats */}
                <div style={{ display: 'flex', gap: '24px', textAlign: 'center' }}>
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
            </div>
          </div>

          {/* Section Header */}
          <div className="section-header">
            <span className="section-title">My Posts</span>
            <div className="section-line"></div>
            <span className="section-star star-spin-reverse">✳</span>
          </div>

          {/* Posts */}
          {error && (
            <div className="form-error">{error}</div>
          )}

          {loading ? (
            <p style={{ fontFamily: "'Cormorant Garamond', serif" }}>Loading posts...</p>
          ) : posts.length === 0 ? (
            <div className="card">
              <div className="card-body text-center">
                <p style={{ fontFamily: "'Cormorant Garamond', serif", opacity: 0.7 }}>
                  You haven't posted anything yet. Go to the <Link to="/feed">Feed</Link> to create your first post!
                </p>
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
                        {getInitial(user.full_name)}
                      </div>
                      <div>
                        <div className="post-author-name">{user.full_name}</div>
                        <div className="post-author-handle">@{user.email?.split('@')[0]}</div>
                      </div>
                    </div>
                    <span className="post-time">{formatDate(post.created_at)}</span>
                  </div>
                 <p className="post-text">{post.content}</p>

{post.image_path && (
  <div style={{ marginTop: 12 }}>
    <img
      src={`http://localhost:8080${post.image_path}`}
      alt="post"
      style={{ maxWidth: "100%", borderRadius: 12 }}
    />
  </div>
)}

<div className="post-actions">
  <button type="button" className="post-action" onClick={() => onToggleLike(post.id)}>
    ♥ Like {likesByPostId[post.id]?.like_count ?? post.like_count ?? 0}
  </button>

  <button type="button" className="post-action" onClick={() => onToggleComments(post.id)}>
    ↩ Reply {commentsByPostId[post.id]?.length ? `(${commentsByPostId[post.id].length})` : ""}
  </button>

  <span className="post-action">⋯</span>
</div>

{openCommentsPostId === post.id && (
  <div style={{ marginTop: 12, borderTop: "1px solid rgba(0,0,0,0.08)", paddingTop: 12 }}>
    <div style={{ display: "grid", gap: 10, marginBottom: 10 }}>
      {(commentsByPostId[post.id] || []).map((c) => (
        <div key={c.id} style={{ fontSize: 14 }}>
          <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 2 }}>
            {c.author_name || "Unknown"} •{" "}
            {c.created_at ? new Date(c.created_at).toLocaleString() : ""}
          </div>
          <div>{c.content}</div>
        </div>
      ))}
      {(commentsByPostId[post.id] || []).length === 0 && (
        <div style={{ fontSize: 13, opacity: 0.6 }}>No comments yet.</div>
      )}
    </div>

    <div style={{ display: "flex", gap: 8 }}>
      <input
        className="form-input"
        placeholder="Write a comment…"
        value={commentDraftByPostId[post.id] || ""}
        onChange={(e) =>
          setCommentDraftByPostId((prev) => ({ ...prev, [post.id]: e.target.value }))
        }
      />
      <button className="btn btn-primary" type="button" onClick={() => onSubmitComment(post.id)}>
        Send ↗
      </button>
    </div>
  </div>
)}

                </div>
              </div>
            ))
          )}
        </main>

        {/* Right Sidebar */}
        <aside>
          {/* Quick Links */}
          <div className="card card-shadow-left">
            <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(23, 3, 18, 0.1)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span className="section-title">Quick Links</span>
              <span className="text-rose star-spin" style={{ fontSize: '12px' }}>✳</span>
            </div>
            
            <Link to="/followers" className="quick-link-item" style={{ padding: '16px 24px', display: 'flex', justifyContent: 'space-between', textDecoration: 'none' }}>
              <span className="quick-link-text">My Followers</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/follow-requests" className="quick-link-item" style={{ padding: '16px 24px', display: 'flex', justifyContent: 'space-between', textDecoration: 'none' }}>
              <span className="quick-link-text">Follow Requests</span>
              <span className="quick-link-arrow">→</span>
            </Link>
            <Link to="/feed" className="quick-link-item" style={{ padding: '16px 24px', display: 'flex', justifyContent: 'space-between', textDecoration: 'none', borderBottom: 'none' }}>
              <span className="quick-link-text">Back to Feed</span>
              <span className="quick-link-arrow">→</span>
            </Link>
          </div>

          {/* Decorative Quote */}
          <div className="quote-box">
            <div className="quote-star star-float star-spin">✳</div>
            <p className="quote-text">
              "Be yourself; everyone else is already taken."
            </p>
          </div>

          {/* Floating Stars */}
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