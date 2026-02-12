import React, { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  acceptMember,
  createGroupPost,
  getGroupDetails,
  getGroupMembers,
  getGroupPosts,
  joinGroup,
  leaveGroup,
  removeMember,
} from "../api/groups";
import GroupMemberCard from "../components/GroupMemberCard";

export default function GroupDetailPage() {
  const { id } = useParams();
  const groupId = Number(id);

  const [group, setGroup] = useState(null);
  const [members, setMembers] = useState([]);
  const [posts, setPosts] = useState([]);
  const [postText, setPostText] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [memberBusy, setMemberBusy] = useState(null);

  const isAdmin = useMemo(() => group?.my_role === "admin" && group?.my_status === "accepted", [group]);

  async function load() {
    setLoading(true);
    try {
      const g = await getGroupDetails(groupId);
      setGroup(g);

      // only fetch members/posts if allowed (backend will forbid otherwise)
      const [m, p] = await Promise.all([
        getGroupMembers(groupId).catch(() => []),
        getGroupPosts(groupId).catch(() => []),
      ]);
      setMembers(m);
      setPosts(p);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to load group");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, [groupId]);

  async function handleJoin() {
    setBusy(true);
    try {
      await joinGroup(groupId);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to join group");
    } finally {
      setBusy(false);
    }
  }

  async function handleLeave() {
    setBusy(true);
    try {
      await leaveGroup(groupId);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to leave group");
    } finally {
      setBusy(false);
    }
  }

  async function handleCreatePost(e) {
    e.preventDefault();
    if (!postText.trim()) return;
    setBusy(true);
    try {
      await createGroupPost(groupId, postText.trim());
      setPostText("");
      const p = await getGroupPosts(groupId);
      setPosts(p);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to create post");
    } finally {
      setBusy(false);
    }
  }

  const pending = members.filter((m) => m.status === "pending");
  const accepted = members.filter((m) => m.status === "accepted");

  async function handleAccept(userId) {
    setMemberBusy(userId);
    try {
      await acceptMember(groupId, userId);
      const m = await getGroupMembers(groupId);
      setMembers(m);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to accept member");
    } finally {
      setMemberBusy(null);
    }
  }

  async function handleRemove(userId) {
    setMemberBusy(userId);
    try {
      await removeMember(groupId, userId);
      const m = await getGroupMembers(groupId);
      setMembers(m);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to remove member");
    } finally {
      setMemberBusy(null);
    }
  }

  if (loading) {
    return (
      <div className="main-content">
        <div className="card"><div className="card-body">Loading...</div></div>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="main-content">
        <div className="card"><div className="card-body">Group not found.</div></div>
      </div>
    );
  }

  const myStatus = group.my_status || "";
  const isAcceptedMember = myStatus === "accepted";
  const isPendingMember = myStatus === "pending";

  return (
    <div className="main-content">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12 }}>
        <div>
          <h2 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>{group.name}</h2>
          <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
            {group.is_private === 1 ? "Private" : "Public"} • {group.member_count ?? 0} members
          </div>
        </div>

        <div style={{ display: "flex", gap: 8 }}>
          <Link to="/groups" className="btn btn-secondary">Back</Link>

          {!myStatus && (
            <button className="btn btn-primary" onClick={handleJoin} disabled={busy}>
              {busy ? "Joining..." : "Join"}
            </button>
          )}
          {isPendingMember && <button className="btn btn-secondary" disabled>Pending</button>}
          {isAcceptedMember && (
            <button className="btn btn-outline btn-danger" onClick={handleLeave} disabled={busy}>
              {busy ? "Leaving..." : "Leave"}
            </button>
          )}
        </div>
      </div>

      <div className="card" style={{ marginTop: 12 }}>
        <div className="card-body">
          <p style={{ margin: 0 }}>{group.description}</p>
        </div>
      </div>

      {/* Members */}
      <div className="section-header" style={{ marginTop: 18 }}>
        <span className="section-title">Members</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      {isAdmin && pending.length > 0 && (
        <div className="card" style={{ marginBottom: 12 }}>
          <div className="card-body">
            <h4 style={{ marginTop: 0 }}>Pending requests</h4>
            {pending.map((m) => (
              <div key={m.id} className="card" style={{ marginBottom: 10 }}>
                <div className="card-body" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                  <div>
                    <div style={{ fontWeight: 700 }}>{m.user_name}</div>
                    <div style={{ fontSize: 12, opacity: 0.75 }}>{m.role} • {m.status}</div>
                  </div>
                  <button className="btn btn-primary" onClick={() => handleAccept(m.user_id)} disabled={memberBusy === m.user_id}>
                    {memberBusy === m.user_id ? "Accepting..." : "Accept"}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {accepted.length === 0 ? (
        <div className="card"><div className="card-body">No members yet.</div></div>
      ) : (
        accepted.map((m) => (
          <GroupMemberCard
            key={m.id}
            member={m}
            isAdmin={isAdmin}
            busy={memberBusy === m.user_id}
            onRemove={handleRemove}
          />
        ))
      )}

      {/* Posts */}
      <div className="section-header" style={{ marginTop: 18 }}>
        <span className="section-title">Group Posts</span>
        <div className="section-line"></div>
        <span className="section-star star-spin">✦</span>
      </div>

      {isAcceptedMember ? (
        <div className="card" style={{ marginBottom: 12 }}>
          <div className="card-body">
            <form onSubmit={handleCreatePost} style={{ display: "grid", gap: 10 }}>
              <textarea
                className="form-input"
                rows={3}
                placeholder="Write something..."
                value={postText}
                onChange={(e) => setPostText(e.target.value)}
              />
              <button className="btn btn-primary" disabled={busy}>
                {busy ? "Posting..." : "Post"}
              </button>
            </form>
          </div>
        </div>
      ) : (
        <div className="card" style={{ marginBottom: 12 }}>
          <div className="card-body" style={{ opacity: 0.8 }}>
            Join the group to see and create posts.
          </div>
        </div>
      )}

      {posts.length === 0 ? (
        <div className="card"><div className="card-body">No posts yet.</div></div>
      ) : (
        posts.map((p) => (
          <div key={p.id} className="card" style={{ marginBottom: 12 }}>
            <div className="card-body">
              <div style={{ fontSize: 12, opacity: 0.75, marginBottom: 6 }}>
                {p.author_name} • {p.created_at ? new Date(p.created_at).toLocaleString() : ""}
              </div>
              <div>{p.content}</div>
            </div>
          </div>
        ))
      )}
    </div>
  );
}
