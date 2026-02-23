import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getUserProfile } from "../api/users";
import { checkMutual } from "../api/followers";
import FollowButton from "../components/FollowButton";
import { MessageCircle } from "lucide-react";
import { useLocation } from "react-router-dom";
import { listComments, createComment, toggleLike } from "../api/posts";

export default function UserProfilePage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState(null);
  const [isMutual, setIsMutual] = useState(false);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
const location = useLocation();
const [postsState, setPostsState] = useState([]);
const [openCommentsPostId, setOpenCommentsPostId] = useState(null);
const [commentsByPostId, setCommentsByPostId] = useState({});
const [commentDraftByPostId, setCommentDraftByPostId] = useState({});
const [likesByPostId, setLikesByPostId] = useState({});



  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      setLoading(true);
      setErr("");
      try {
        const [profileRes, mutualRes] = await Promise.all([
          getUserProfile(id),
          checkMutual(id)
        ]);
        if (!cancelled) {
          setData(profileRes);
          setPostsState(profileRes?.posts || []);
          setIsMutual(mutualRes);
        }
      } catch (e) {
        if (!cancelled) setErr(e?.message || "Failed to load profile");
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    run();
    return () => {
      cancelled = true;
    };
}, [id, location.key]);

  const handleMessage = () => {
    navigate(`/messages/${id}`);
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

    // update visible count immediately
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
    <div className="main-content">
      <div className="card">
        <div className="card-header">
          <span className="card-header-title">{user.full_name || "User"}</span>
          <span className="card-header-star star-spin">✦</span>
        </div>

        <div className="card-body">
          {avatarSrc && (
            <div style={{ marginBottom: 12 }}>
              <img
                src={avatarSrc}
                alt="avatar"
                style={{ width: 72, height: 72, borderRadius: 999, objectFit: "cover" }}
              />
            </div>
          )}

          {user.email && <p><strong>Email:</strong> {user.email}</p>}
          {user.date_of_birth && <p><strong>Date of Birth:</strong> {user.date_of_birth}</p>}
          <p><strong>Nickname:</strong> {user.nickname || "-"}</p>
          <p><strong>About:</strong> {user.about_me || "-"}</p>
          <p><strong>Privacy:</strong> {user.is_private ? "Private" : "Public"}</p>

          <p style={{ marginTop: 10 }}>
            <strong>Followers:</strong> {counts?.followers ?? 0}{" "}
            | <strong>Following:</strong> {counts?.following ?? 0}
          </p>

          {!viewer?.is_self && (
            <div style={{ marginTop: 12, display: 'flex', gap: '12px', alignItems: 'center' }}>
              <FollowButton userId={user.id} isPrivate={user.is_private} />
              
              {/* Message Button - only show if mutual friends */}
              {isMutual && (
                <button 
                  className="btn btn-outline"
                  onClick={handleMessage}
                  style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                >
                  <MessageCircle size={16} />
                  Message
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      <div style={{ marginTop: 18 }} className="card">
        <div className="card-header">
          <span className="card-header-title">Posts</span>
          <span className="card-header-star star-spin">✦</span>
        </div>

        <div className="card-body">
          {!viewer?.can_view && user.is_private ? (
            <p>This profile is private. Follow to view posts.</p>
          ) : postsState && postsState.length > 0 ? (
       postsState.map((p) => (
  <div key={p.id} className="border p-3 mb-3 rounded">
    <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 6 }}>
      {p.author_name || "User"} • {new Date(p.created_at).toLocaleString()}
    </div>

    <div>{p.content}</div>

    {p.image_path && (
      <div style={{ marginTop: 12 }}>
        <img
          src={p.image_path.startsWith("http") ? p.image_path : `${API_BASE}${p.image_path}`}
          alt="post"
          style={{ maxWidth: "100%", borderRadius: 12 }}
        />
      </div>
    )}

    <div className="post-actions" style={{ marginTop: 10 }}>
      <button type="button" className="post-action" onClick={() => onToggleLike(p.id)}>
        ♥ Like {likesByPostId[p.id]?.like_count ?? p.like_count ?? 0}
      </button>

      <button type="button" className="post-action" onClick={() => onToggleComments(p.id)}>
        ↩ Reply {commentsByPostId[p.id]?.length ? `(${commentsByPostId[p.id].length})` : ""}
      </button>
    </div>

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
            <p>No posts yet.</p>
          )}
        </div>
      </div>
    </div>
  );
}