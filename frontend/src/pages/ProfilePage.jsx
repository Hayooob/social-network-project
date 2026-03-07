import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../VerifyAuth';
import { getMyPosts, listComments, createComment, toggleLike } from '../api/posts';
import { getFollowCounts } from '../api/followers';
import { togglePrivacy, updateProfile } from '../api/users';

export default function ProfilePage() {
  const { user, setUser } = useAuth();
  const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const [posts, setPosts] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);
  const [privacyLoading, setPrivacyLoading] = useState(false);
  const [error, setError] = useState('');
  const [openCommentsPostId, setOpenCommentsPostId] = useState(null);
  const [commentsByPostId, setCommentsByPostId] = useState({});
  const [commentDraftByPostId, setCommentDraftByPostId] = useState({});
  const [likesByPostId, setLikesByPostId] = useState({});

  // editable profile state
  const [editing, setEditing] = useState(false);
  const [editFullName, setEditFullName] = useState('');
  const [editDateOfBirth, setEditDateOfBirth] = useState('');
  const [editNickname, setEditNickname] = useState('');
  const [editAboutMe, setEditAboutMe] = useState('');
  const [avatarFile, setAvatarFile] = useState(null);
  const [savingProfile, setSavingProfile] = useState(false);
  const [editError, setEditError] = useState('');

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

  useEffect(() => {
    if (user) {
      setEditFullName(user.full_name || '');
      setEditDateOfBirth(user.date_of_birth || '');
      setEditNickname(user.nickname || '');
      setEditAboutMe(user.about_me || '');
      setAvatarFile(null);
    }
  }, [user]);

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

  const handleSaveProfile = async (e) => {
    e.preventDefault();
    setSavingProfile(true);
    setEditError('');
    try {
      const payload = {
        full_name: editFullName,
        date_of_birth: editDateOfBirth,
        nickname: editNickname || undefined,
        about_me: editAboutMe || undefined,
        avatarFile,
      };
      const updated = await updateProfile(payload);
      setUser(updated);
      setEditing(false);
    } catch (err) {
      console.error('updateProfile error', err);
      setEditError(err.message || 'Failed to update profile');
    } finally {
      setSavingProfile(false);
    }
  };

  const handleCancelEdit = () => {
    if (user) {
      setEditFullName(user.full_name || '');
      setEditDateOfBirth(user.date_of_birth || '');
      setEditNickname(user.nickname || '');
      setEditAboutMe(user.about_me || '');
    }
    setAvatarFile(null);
    setEditError('');
    setEditing(false);
  };

  const getInitial = (name) => name ? name.charAt(0).toUpperCase() : '?';

  const avatarSrc = user?.avatar_url
    ? (user.avatar_url.startsWith("http") ? user.avatar_url : `${API_BASE}${user.avatar_url}`)
    : null;

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
            <div className="card-header" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span className="card-header-title">My Profile</span>
              {!editing && (
                <button className="btn btn-outline" onClick={() => setEditing(true)} style={{ padding: '4px 8px' }}>
                  Edit Profile
                </button>
              )}
              <span className="card-header-star star-spin">✦</span>
            </div>

            <div className="card-body">
              {editing ? (
                <form onSubmit={handleSaveProfile}>
                  <div style={{ display: 'flex', gap: '24px', alignItems: 'flex-start' }}>
                    {/* Avatar + upload */}
                    <div>
                      <div className="avatar" style={{ flexShrink: 0 }}>
                        {avatarSrc ? (
                          <img
                            src={avatarSrc}
                            alt="avatar"
                            style={{ width: 56, height: 56, borderRadius: 999, objectFit: "cover" }}
                          />
                        ) : (
                          getInitial(user.full_name)
                        )}
                        <div className="avatar-status"></div>
                      </div>
                      <input
                        type="file"
                        accept="image/*"
                        style={{ marginTop: 8 }}
                        onChange={(e) => setAvatarFile(e.target.files?.[0] || null)}
                      />
                    </div>

                    {/* Info inputs */}
                    <div style={{ flex: 1, display: 'grid', gap: 8 }}>
                      <input
                        className="form-input"
                        value={editFullName}
                        onChange={(e) => setEditFullName(e.target.value)}
                        placeholder="Full name"
                      />
                      <input
                        className="form-input"
                        type="date"
                        value={editDateOfBirth}
                        onChange={(e) => setEditDateOfBirth(e.target.value)}
                        placeholder="Date of birth"
                      />
                      <input
                        className="form-input"
                        value={editNickname}
                        onChange={(e) => setEditNickname(e.target.value)}
                        placeholder="Nickname (optional)"
                      />
                      <textarea
                        className="form-input"
                        value={editAboutMe}
                        onChange={(e) => setEditAboutMe(e.target.value)}
                        placeholder="About me (optional)"
                        rows={3}
                      />
                    </div>
                  </div>

                  {editError && <div className="form-error" style={{ marginTop: 8 }}>{editError}</div>}

                  <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
                    <button className="btn btn-primary" type="submit" disabled={savingProfile}>
                      {savingProfile ? 'Saving...' : 'Save'}
                    </button>
                    <button className="btn btn-outline" type="button" onClick={handleCancelEdit} disabled={savingProfile}>
                      Cancel
                    </button>
                  </div>
                </form>
              ) : (
                <div style={{ display: 'flex', gap: '24px', alignItems: 'flex-start' }}>
                  {/* Avatar */}
                  <div className="avatar" style={{ flexShrink: 0 }}>
                    {avatarSrc ? (
                      <img
                        src={avatarSrc}
                        alt="avatar"
                        style={{ width: 56, height: 56, borderRadius: 999, objectFit: "cover" }}
                      />
                    ) : (
                      getInitial(user.full_name)
                    )}
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

                    <div style={{ marginTop: 10, display: 'grid', gap: 6 }}>
                      <div style={{ fontSize: 13, opacity: 0.85, fontFamily: "'Cormorant Garamond', serif" }}>
                        <strong>Date of Birth:</strong> {user.date_of_birth || "-"}
                      </div>
                      {user.nickname && (
                        <div style={{ fontSize: 13, opacity: 0.85, fontFamily: "'Cormorant Garamond', serif" }}>
                          <strong>Nickname:</strong> {user.nickname}
                        </div>
                      )}
                      {user.about_me && (
                        <div style={{ fontSize: 13, opacity: 0.85, fontFamily: "'Cormorant Garamond', serif" }}>
                          <strong>About Me:</strong> {user.about_me}
                        </div>
                      )}
                    </div>

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
              )}
            </div>{/* ← fixed: was missing closing of card-body for non-editing view */}
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

          <div className="quote-box">
            <div className="quote-star star-float star-spin">✳</div>
            <p className="quote-text">
              "Be yourself; everyone else is already taken."
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
