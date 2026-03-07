import React from "react";
import { useEffect, useState } from "react";
import { useParams, Link, useLocation } from "react-router-dom";
import { getUserProfile } from "../api/users";
import { checkMutual, followUser, unfollowUser } from "../api/followers";
import { MessageCircle } from "lucide-react";
import { listComments, createComment, toggleLike } from "../api/posts";

export default function UserProfilePage() {
  const { id } = useParams();
  const location = useLocation();
  const [data, setData] = useState(null);
  const [isMutual, setIsMutual] = useState(false);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [followLoading, setFollowLoading] = useState(false);
  const [postsState, setPostsState] = useState([]);
  const [openCommentsPostId, setOpenCommentsPostId] = useState(null);
  const [commentsByPostId, setCommentsByPostId] = useState({});
  const [commentDraftByPostId, setCommentDraftByPostId] = useState({});
  const [likesByPostId, setLikesByPostId] = useState({});

  const fetchProfile = async () => {
    try {
      const [profileRes, mutualRes] = await Promise.all([
        getUserProfile(id),
        checkMutual(id)
      ]);
      setData(profileRes);
      setPostsState(profileRes?.posts || []);
      setIsMutual(mutualRes);
    } catch (e) {
      setErr(e?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    setErr("");
    fetchProfile();
  }, [id, location.key]);

  const handleFollow = async () => {
    setFollowLoading(true);
    try {
      const isCurrentlyFollowing = data?.viewer?.is_following;

      if (!isCurrentlyFollowing) {
        await followUser(id);
      } else {
        await unfollowUser(id);
      }

      await fetchProfile();
    } catch (err) {
      if (err.message?.includes("already following")) {
        // Just refresh state silently
        await fetchProfile();
      } else {
        alert(err.message || "Something went wrong");
      }
    } finally {
      setFollowLoading(false);
    }
  };

  const getButtonText = () => {
    if (followLoading) return "...";
    if (!data?.viewer?.is_following && !data?.viewer?.is_pending) return "Follow";
    if (data?.viewer?.is_pending) return "Requested";
    if (data?.viewer?.is_following) return "Following";
    return "Follow";
  };

  const getButtonStyle = () => {
    if (data?.viewer?.is_following) {
      return {
        backgroundColor: "var(--soft-linen)",
        color: "var(--coffee-bean)",
        border: "1px solid var(--coffee-bean)"
      };
    }
    if (data?.viewer?.is_pending) {
      return {
        backgroundColor: "var(--blush-rose)",
        color: "var(--white)",
        border: "1px solid var(--blush-rose)"
      };
    }
    return {
      backgroundColor: "var(--dusk-blue)",
      color: "var(--white)",
      border: "1px solid var(--dusk-blue)"
    };
  };

  const formatDate = (timestamp) => {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffMins < 1) return "Just now";
      if (diffMins < 60) return `${diffMins}m ago`;
      if (diffHours < 24) return `${diffHours}h ago`;
      if (diffDays < 7) return `${diffDays}d ago`;
      return date.toLocaleDateString();
    } catch {
      return "Recently";
    }
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
      setPostsState((prev) =>
        prev.map((p) => (p.id === postId ? { ...p, like_count: res.like_count } : p))
      );
    } catch (e) {
      console.error("toggleLike error:", e);
      alert(e.message || "Failed to like");
    }
  };

  if (loading) {
    return (
      <div className="main-content page-center">
        <p>Loading profile...</p>
      </div>
    );
  }

  if (err) {
    return (
      <div className="main-content page-center">
        <p style={{ color: "crimson" }}>{err}</p>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="main-content page-center">
        <p>Profile not found.</p>
      </div>
    );
  }

  const { user, counts, viewer } = data;
  const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const avatarSrc = user?.avatar_url
    ? (user.avatar_url.startsWith("http") ? user.avatar_url : `${API_BASE}${user.avatar_url}`)
    : null;

  return (
    <div className="main-content" style={{ maxWidth: "800px", margin: "0 auto", padding: "24px" }}>
      {/* Profile Card */}
      <div className="card" style={{ marginBottom: "24px" }}>
        <div className="card-header">
          <span className="card-header-title">{user.full_name || "User"}</span>
          <span className="card-header-star star-spin">✦</span>
        </div>

        <div className="card-body" style={{ padding: "24px" }}>
          {/* Avatar and Basic Info */}
          <div style={{ display: "flex", alignItems: "center", gap: "20px", marginBottom: "20px" }}>
            {avatarSrc ? (
              <img
                src={avatarSrc}
                alt="avatar"
                style={{ width: 80, height: 80, borderRadius: "50%", objectFit: "cover" }}
              />
            ) : (
              <div style={{
                width: 80,
                height: 80,
                borderRadius: "50%",
                backgroundColor: "var(--dusk-blue)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "var(--white)",
                fontSize: "28px",
                fontWeight: 600
              }}>
                {(user.full_name || "U").charAt(0).toUpperCase()}
              </div>
            )}
            <div>
              <h2 style={{ margin: 0, fontSize: "20px", fontWeight: 600, color: "var(--coffee-bean)" }}>
                {user.full_name}
              </h2>
              {user.nickname && (
                <p style={{ margin: "4px 0 0", fontSize: "14px", color: "var(--jet-black)", opacity: 0.6 }}>
                  @{user.nickname}
                </p>
              )}
              <p style={{ margin: "4px 0 0", fontSize: "13px", color: "var(--jet-black)", opacity: 0.5 }}>
                {user.is_private ? "🔒 Private" : "🌐 Public"}
              </p>
            </div>
          </div>

          {/* About Me */}
          {user.about_me && (
            <p style={{
              margin: "0 0 20px",
              fontSize: "15px",
              lineHeight: 1.6,
              color: "var(--jet-black)",
              fontFamily: "'Cormorant Garamond', serif"
            }}>
              {user.about_me}
            </p>
          )}

          {/* Stats */}
          <div style={{
            display: "flex",
            gap: "40px",
            padding: "16px 0",
            borderTop: "1px solid rgba(23, 3, 18, 0.08)",
            borderBottom: "1px solid rgba(23, 3, 18, 0.08)",
            marginBottom: "20px"
          }}>
            <div style={{ textAlign: "center" }}>
              <div style={{ fontSize: "24px", fontWeight: 700, color: "var(--coffee-bean)" }}>
                {counts?.followers ?? 0}
              </div>
              <div style={{ fontSize: "12px", color: "var(--jet-black)", opacity: 0.6, textTransform: "uppercase", letterSpacing: "1px" }}>
                Followers
              </div>
            </div>
            <div style={{ textAlign: "center" }}>
              <div style={{ fontSize: "24px", fontWeight: 700, color: "var(--coffee-bean)" }}>
                {counts?.following ?? 0}
              </div>
              <div style={{ fontSize: "12px", color: "var(--jet-black)", opacity: 0.6, textTransform: "uppercase", letterSpacing: "1px" }}>
                Following
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          {!viewer?.is_self && (
            <div style={{ display: "flex", gap: "12px", alignItems: "center" }}>
              <button
                onClick={handleFollow}
                disabled={followLoading}
                style={{
                  ...getButtonStyle(),
                  padding: "10px 24px",
                  borderRadius: "8px",
                  fontSize: "14px",
                  fontWeight: 600,
                  fontFamily: "'Montserrat', sans-serif",
                  cursor: followLoading ? "not-allowed" : "pointer",
                  opacity: followLoading ? 0.7 : 1,
                  transition: "all 0.2s ease"
                }}
              >
                {getButtonText()}
              </button>

              {(viewer?.is_following || isMutual) && (
                <Link
                  to={`/messages/${user.id}`}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "8px",
                    padding: "10px 20px",
                    backgroundColor: "var(--soft-linen)",
                    color: "var(--dusk-blue)",
                    border: "1px solid var(--dusk-blue)",
                    borderRadius: "8px",
                    fontSize: "14px",
                    fontWeight: 600,
                    fontFamily: "'Montserrat', sans-serif",
                    textDecoration: "none",
                    transition: "all 0.2s ease"
                  }}
                >
                  <MessageCircle size={16} />
                  Message
                </Link>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Posts Section */}
      <div className="card">
        <div className="card-header">
          <span className="card-header-title">Posts</span>
          <span className="card-header-star star-spin">✦</span>
        </div>

        <div style={{ padding: "0" }}>
          {!viewer?.can_view && user.is_private ? (
            <div style={{ padding: "40px 24px", textAlign: "center" }}>
              <p style={{
                color: "var(--jet-black)",
                opacity: 0.6,
                fontFamily: "'Cormorant Garamond', serif",
                fontSize: "16px"
              }}>
                This profile is private. Follow to view posts.
              </p>
            </div>
          ) : postsState && postsState.length > 0 ? (
            postsState.map((p, index) => (
              <div
                key={p.id}
                style={{
                  padding: "20px 24px",
                  borderBottom: index < postsState.length - 1 ? "1px solid rgba(23, 3, 18, 0.08)" : "none"
                }}
              >
                {/* Post Header */}
                <div style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  marginBottom: "12px"
                }}>
                  {(() => {
                    const avatarSrc = p.author_avatar_url
                      ? (p.author_avatar_url.startsWith("http")
                        ? p.author_avatar_url
                        : `${API_BASE}${p.author_avatar_url}`)
                      : null;
                    const color = index % 3 === 0 ? "var(--dusk-blue)" : index % 3 === 1 ? "var(--blush-rose)" : "var(--jet-black)";
                    return (
                      <div style={{
                        width: "36px",
                        height: "36px",
                        borderRadius: "50%",
                        backgroundColor: avatarSrc ? "transparent" : color,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        color: "var(--white)",
                        fontSize: "14px",
                        fontWeight: 600
                      }}>
                        {avatarSrc ? (
                          <img
                            src={avatarSrc}
                            alt="avatar"
                            style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                          />
                        ) : (
                          (p.author_name || "U").charAt(0).toUpperCase()
                        )}
                      </div>
                    );
                  })()}
                  <div style={{ flex: 1 }}>
                    <div style={{
                      fontWeight: 600,
                      fontSize: "14px",
                      color: "var(--coffee-bean)",
                      fontFamily: "'Montserrat', sans-serif"
                    }}>
                      {p.author_name || "User"}
                    </div>
                    <div style={{
                      fontSize: "12px",
                      color: "var(--jet-black)",
                      opacity: 0.5,
                      fontFamily: "'Montserrat', sans-serif"
                    }}>
                      {formatDate(p.created_at)}
                    </div>
                  </div>
                </div>

                {/* Post Content */}
                <p style={{
                  margin: 0,
                  fontSize: "15px",
                  lineHeight: 1.6,
                  color: "var(--jet-black)",
                  fontFamily: "'Cormorant Garamond', serif",
                  whiteSpace: "pre-wrap"
                }}>
                  {p.content}
                </p>

                {/* Post Image */}
                {p.image_path && (
                  <div style={{ marginTop: 12 }}>
                    <img
                      src={p.image_path.startsWith("http") ? p.image_path : `${API_BASE}${p.image_path}`}
                      alt="post"
                      style={{ maxWidth: "100%", borderRadius: 12 }}
                    />
                  </div>
                )}

                {/* Post Actions */}
                <div className="post-actions" style={{ marginTop: 12 }}>
                  <button type="button" className="post-action" onClick={() => onToggleLike(p.id)}>
                    ♥ Like {likesByPostId[p.id]?.like_count ?? p.like_count ?? 0}
                  </button>
                  <button type="button" className="post-action" onClick={() => onToggleComments(p.id)}>
                    ↩ Reply {commentsByPostId[p.id]?.length ? `(${commentsByPostId[p.id].length})` : ""}
                  </button>
                </div>

                {/* Comments Section */}
                {openCommentsPostId === p.id && (
                  <div style={{ marginTop: 12, borderTop: "1px solid rgba(0,0,0,0.08)", paddingTop: 12 }}>
                    <div style={{ display: "grid", gap: 10, marginBottom: 10 }}>
                      {(commentsByPostId[p.id] || []).map((c) => (
                        <div key={c.id} style={{ fontSize: 14 }}>
                          <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 2 }}>
                            {c.author_name || "Unknown"} •{" "}
                            {c.created_at ? new Date(c.created_at).toLocaleString() : ""}
                          </div>
                          <div>{c.content}</div>
                        </div>
                      ))}
                      {(commentsByPostId[p.id] || []).length === 0 && (
                        <div style={{ fontSize: 13, opacity: 0.6 }}>No comments yet.</div>
                      )}
                    </div>

                    <div style={{ display: "flex", gap: 8 }}>
                      <input
                        className="form-input"
                        placeholder="Write a comment…"
                        value={commentDraftByPostId[p.id] || ""}
                        onChange={(e) =>
                          setCommentDraftByPostId((prev) => ({ ...prev, [p.id]: e.target.value }))
                        }
                      />
                      <button className="btn btn-primary" type="button" onClick={() => onSubmitComment(p.id)}>
                        Send ↗
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))
          ) : (
            <div style={{ padding: "40px 24px", textAlign: "center" }}>
              <p style={{
                color: "var(--jet-black)",
                opacity: 0.6,
                fontFamily: "'Cormorant Garamond', serif",
                fontSize: "16px"
              }}>
                No posts yet.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}