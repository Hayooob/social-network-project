import { useEffect, useState } from "react";
import { useAuth } from "../VerifyAuth";
import { getMyPosts } from "../api/posts";
import PostCard from "../components/PostCard";

// ✅ Stage 5
import PrivacyToggle from "../components/PrivacyToggle";

export default function ProfilePage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [err, setErr] = useState("");

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      setErr("");
      try {
        const res = await getMyPosts();
        if (!cancelled) setPosts(res || []);
      } catch (e) {
        if (!cancelled) setErr(e?.message || "Failed to load posts");
      }
    };

    run();
    return () => {
      cancelled = true;
    };
  }, []);

  if (!user) return null;

  return (
    <div>
      <h2>My Profile</h2>

      <p>
        <strong>Name:</strong> {user.full_name}
      </p>
      <p>
        <strong>Email:</strong> {user.email}
      </p>

      {/* ✅ Stage 5 privacy toggle */}
      <PrivacyToggle />

      <hr />

      <h3>My Posts</h3>
      {err && <p style={{ color: "crimson" }}>{err}</p>}
      {posts.length === 0 ? <p>No posts yet.</p> : posts.map((p) => <PostCard key={p.id} post={p} />)}
    </div>
  );
}
