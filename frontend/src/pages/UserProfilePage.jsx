import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getUserProfile } from "../api/users";
import FollowButton from "../components/FollowButton";

export default function UserProfilePage() {
  const { id } = useParams();
  const [data, setData] = useState(null); // { user, counts, posts, viewer }
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      setLoading(true);
      setErr("");
      try {
        const res = await getUserProfile(id);
        if (!cancelled) setData(res);
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
  }, [id]);

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

  const { user, counts, posts, viewer } = data;

  return (
    <div className="main-content">
      <div className="card">
        <div className="card-header">
          <span className="card-header-title">{user.full_name || "User"}</span>
          <span className="card-header-star star-spin">✦</span>
        </div>

        <div className="card-body">
          <p><strong>Nickname:</strong> {user.nickname || "-"}</p>
          <p><strong>About:</strong> {user.about_me || "-"}</p>
          <p><strong>Privacy:</strong> {user.is_private ? "Private" : "Public"}</p>

          <p style={{ marginTop: 10 }}>
            <strong>Followers:</strong> {counts?.followers ?? 0}{" "}
            | <strong>Following:</strong> {counts?.following ?? 0}
          </p>

          {!viewer?.is_self && (
            <div style={{ marginTop: 12 }}>
              <FollowButton userId={user.id} isPrivate={user.is_private} />
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
          ) : posts && posts.length > 0 ? (
            posts.map((p) => (
              <div key={p.id} className="border p-3 mb-3 rounded">
                <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 6 }}>
                  {p.author_name || "User"} • {new Date(p.created_at).toLocaleString()}
                </div>
                <div>{p.content}</div>
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

