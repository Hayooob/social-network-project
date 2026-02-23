import React, { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { getFeed, createPost, listComments, createComment, toggleLike } from '../api/posts';
import { getSuggestedUsers } from '../api/auth';
import { followUser, getFollowCounts, getMyFollowers } from '../api/followers';
import { useAuth } from '../VerifyAuth';
import SearchBar from '../components/SearchBar';


export default function FeedPage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [suggestions, setSuggestions] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [posting, setPosting] = useState(false);
const [openCommentsPostId, setOpenCommentsPostId] = useState(null);
const [commentsByPostId, setCommentsByPostId] = useState({});
const [commentDraftByPostId, setCommentDraftByPostId] = useState({});
const [likesByPostId, setLikesByPostId] = useState({});
const [imageFile, setImageFile] = useState(null);
const [privacy, setPrivacy] = useState("public");
const [followers, setFollowers] = useState([]);
const [allowedViewers, setAllowedViewers] = useState([]);
const location = useLocation();

  const fetchData = async () => {
    setLoading(true);
    try {
      const [feedData, suggestionsData, countsData] = await Promise.all([
        getFeed(),
        getSuggestedUsers(),
        getFollowCounts()
      ]);
      setPosts(feedData);
      setSuggestions(suggestionsData);
      setCounts(countsData);
    } catch (err) {
      console.error('Error loading feed:', err);
    } finally {
      setLoading(false);
    }
  };

useEffect(() => { fetchData(); }, [location.key]);

// Load followers once (used for "private" post audience)
useEffect(() => {
  let cancelled = false;
  (async () => {
    try {
      const f = await getMyFollowers();
      if (!cancelled) setFollowers(f);
    } catch {
      // ignore
    }
  })();
  return () => {
    cancelled = true;
  };
}, []);


  const handlePostSubmit = async (e) => {
    e.preventDefault();
    if (!content.trim()) return;

    if (privacy === "private" && (!allowedViewers || allowedViewers.length === 0)) {
      alert("Choose at least one follower for a private post.");
      return;
    }

    setPosting(true);
    try {
await createPost(content.trim(), privacy, imageFile, allowedViewers);
setImageFile(null);
setAllowedViewers([]);
setPrivacy("public");
      setContent('');
      fetchData();
    } catch (err) {
      console.error('Error creating post:', err);
      alert('Failed to create post.');
    } finally {
      setPosting(false);
    }
  };

  const handleFollow = async (userId) => {
    try {
      await followUser(userId);
      setSuggestions(prev => prev.filter(u => u.id !== userId));
      fetchData();
    } catch (err) {
      console.error('Error following user:', err);
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
    await loadComments(postId);
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
      prev.map((p) => (p.id === postId ? { ...p, like_count: res.like_count } : p))
    );
  } catch (e) {
    console.error("toggleLike error:", e);
    alert(e.message || "Failed to like");
  }
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

        {/* Main Feed */}
        <main>
          {/* Create Post Card */}
          <div className="card card-shadow-left">
            <div className="card-header">
              <span className="card-header-title">Share Something</span>
              <span className="card-header-star star-spin">✦</span>
            </div>
            <div className="card-body">
              <form onSubmit={handlePostSubmit}>
                <textarea
                  className="form-textarea"
                  placeholder="What's on your mind?"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  disabled={posting}
                  rows={3}
                ></textarea>

                <div style={{ marginTop: 12 }}>
                  <label style={{ fontSize: 13, marginRight: 8 }}>Privacy:</label>
                  <select
                    className="form-input"
                    value={privacy}
                    onChange={(e) => setPrivacy(e.target.value)}
                    style={{ width: "auto" }}
                  >
                    <option value="public">Public</option>
                    <option value="almost-private">Almost Private (followers)</option>
                    <option value="private">Private (choose followers)</option>
                  </select>
                </div>

                {privacy === "private" && (
                  <div className="card" style={{ marginTop: 12 }}>
                    <div className="card-body">
                      <p style={{ margin: 0, marginBottom: 10, fontSize: 12, opacity: 0.8 }}>
                        Choose which followers can see this post:
                      </p>

                      {followers.length === 0 ? (
                        <p style={{ margin: 0, fontSize: 12, opacity: 0.8 }}>No followers found.</p>
                      ) : (
                        <div style={{ display: 'grid', gap: 8 }}>
                          {followers.map((f) => (
                            <label key={f.follower_id} style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                              <input
                                type="checkbox"
                                checked={allowedViewers.includes(f.follower_id)}
                                onChange={(e) => {
                                  const id = f.follower_id;
                                  setAllowedViewers((prev) =>
                                    e.target.checked ? [...prev, id] : prev.filter((x) => x !== id)
                                  );
                                }}
                              />
                              <span>{f.follower_name || `User #${f.follower_id}`}</span>
                            </label>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
                <input
  type="file"
  accept="image/*"
  disabled={posting}
  onChange={(e) => setImageFile(e.target.files?.[0] || null)}
  style={{ marginTop: 12 }}
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
                      <Link 
                        to={`/users/${post.user_id}`} 
                        style={{ textDecoration: 'none', display: 'block' }}
                      >
                        <div className={`avatar avatar-small ${index % 3 === 1 ? 'bg-rose' : index % 3 === 2 ? 'bg-dark' : 'bg-blue'}`}>
                          {getInitial(post.author_name)}
                        </div>
                      </Link>
                      <div>
                        <Link 
                          to={`/users/${post.user_id}`} 
                          style={{ 
                            textDecoration: 'none', 
                            color: 'var(--coffee-bean)',
                            fontWeight: 600,
                            fontSize: '14px',
                            fontFamily: "'Montserrat', sans-serif",
                            display: 'block',
                            cursor: 'pointer'
                          }}
                          onMouseEnter={(e) => e.currentTarget.style.color = 'var(--dusk-blue)'}
                          onMouseLeave={(e) => e.currentTarget.style.color = 'var(--coffee-bean)'}
                        >
                          {post.author_name || 'Unknown'}
                        </Link>
                        <div className="post-author-handle">@{post.author_name?.toLowerCase().replace(' ', '') || 'user'}</div>
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
  <button
    type="button"
    className="post-action"
    onClick={() => onToggleLike(post.id)}
  >
    ♥ Like {likesByPostId[post.id]?.like_count ?? post.like_count ?? 0}
  </button>

  <button
    type="button"
    className="post-action"
    onClick={() => onToggleComments(post.id)}
  >
    ↩ Reply {commentsByPostId[post.id]?.length ? `(${commentsByPostId[post.id].length})` : ''}
  </button>

  <span className="post-action">⋯</span>
</div>

{openCommentsPostId === post.id && (
  <div style={{ marginTop: 12, borderTop: '1px solid rgba(0,0,0,0.08)', paddingTop: 12 }}>
    <div style={{ display: 'grid', gap: 10, marginBottom: 10 }}>
      {(commentsByPostId[post.id] || []).map((c) => (
        <div key={c.id} style={{ fontSize: 14 }}>
          <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 2 }}>
            {c.author_name || c.username || 'Unknown'} •{' '}
            {c.created_at ? new Date(c.created_at).toLocaleString() : ''}
          </div>
          <div>{c.content}</div>
        </div>
      ))}

      {(commentsByPostId[post.id] || []).length === 0 && (
        <div style={{ fontSize: 13, opacity: 0.6 }}>No comments yet.</div>
      )}
    </div>

    <div style={{ display: 'flex', gap: 8 }}>
      <input
        className="form-input"
        placeholder="Write a comment…"
        value={commentDraftByPostId[post.id] || ''}
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
          {/* Search Bar */}
          <SearchBar />
          
          {/* Suggested Users */}
          <div className="card card-shadow-left">
            <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(23, 3, 18, 0.1)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span className="section-title">Suggested</span>
              <span className="text-rose star-float-slow" style={{ fontSize: '12px' }}>✦</span>
            </div>
            
            {suggestions.length === 0 ? (
              <div style={{ padding: '24px', textAlign: 'center' }}>
                <p style={{ fontSize: '13px', opacity: 0.6, fontFamily: "'Cormorant Garamond', serif" }}>No suggestions yet</p>
              </div>
            ) : (
              suggestions.map((suggestedUser, index) => (
                <div key={suggestedUser.id} className="suggestion-item">
                  <div className="suggestion-info">
                    <Link to={`/users/${suggestedUser.id}`} style={{ textDecoration: 'none' }}>
                      <div className={`avatar avatar-tiny ${index % 3 === 0 ? 'bg-blue' : index % 3 === 1 ? 'bg-rose' : 'bg-dark'}`}>
                        {suggestedUser.full_name ? suggestedUser.full_name.charAt(0).toUpperCase() : '?'}
                      </div>
                    </Link>
                    <Link to={`/users/${suggestedUser.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
                      <span className="suggestion-name">{suggestedUser.full_name || 'Unknown'}</span>
                    </Link>
                  </div>
                  <button 
                    className="btn btn-outline"
                    onClick={() => handleFollow(suggestedUser.id)}
                  >
                    Follow
                  </button>
                </div>
              ))
            )}
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