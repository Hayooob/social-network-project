import React, { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  acceptMember,
  createGroupPost,
  getGroupDetails,
  getGroupMembers,
  getGroupPosts,
  inviteUserToGroup,
  removeMember,
  requestJoinGroup,
  leaveGroup,
} from "../api/groups";
import { searchUsers } from "../api/users";
import { createGroupEvent, getGroupEvents, getGroupEventResponses, respondToGroupEvent } from "../api/events";
import GroupMemberCard from "../components/GroupMemberCard";
import EventCard from "../components/EventCard";
import GroupChat from "../components/GroupChat";
export default function GroupDetailPage() {
  const { id } = useParams();
  const groupId = Number(id);

  const [group, setGroup] = useState(null);
  const [members, setMembers] = useState([]);
  const [posts, setPosts] = useState([]);
  const [events, setEvents] = useState([]);

  const [postText, setPostText] = useState("");

  // Invite UI
  const [q, setQ] = useState("");
  const [searchRes, setSearchRes] = useState([]);
  const [inviteBusy, setInviteBusy] = useState(false);

  // Events create UI
  const [evTitle, setEvTitle] = useState("");
  const [evDesc, setEvDesc] = useState("");
  const [evLoc, setEvLoc] = useState("");
  const [evDate, setEvDate] = useState(""); // datetime-local
  const [eventBusyId, setEventBusyId] = useState(null);

  // Response viewer
  const [openRespEventId, setOpenRespEventId] = useState(null);
  const [openResponses, setOpenResponses] = useState([]);

  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [memberBusy, setMemberBusy] = useState(null);

  const isAdmin = useMemo(
    () => group?.my_role === "admin" && group?.my_status === "accepted",
    [group]
  );

  const isAcceptedMember = group?.my_status === "accepted";

  async function load() {
    setLoading(true);
    try {
      const g = await getGroupDetails(groupId);
      setGroup(g);

      // members/posts/events only if member (backend enforces too)
      const [m, p, e] = await Promise.all([
        getGroupMembers(groupId).catch(() => []),
        getGroupPosts(groupId).catch(() => []),
        getGroupEvents(groupId).catch(() => []),
      ]);
      setMembers(m);
      setPosts(p);
      setEvents(e);
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

  async function handleRequestJoin() {
    setBusy(true);
    try {
      await requestJoinGroup(groupId);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to request join");
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
      setPosts(await getGroupPosts(groupId));
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to create post");
    } finally {
      setBusy(false);
    }
  }

  const pendingRequests = useMemo(() => members.filter((m) => m.status === "pending"), [members]);
  const acceptedMembers = useMemo(() => members.filter((m) => m.status === "accepted"), [members]);

  async function handleAccept(userId) {
    setMemberBusy(userId);
    try {
      await acceptMember(groupId, userId);
      setMembers(await getGroupMembers(groupId));
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to accept");
    } finally {
      setMemberBusy(null);
    }
  }

  async function handleRemove(userId) {
    setMemberBusy(userId);
    try {
      await removeMember(groupId, userId);
      setMembers(await getGroupMembers(groupId));
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to remove");
    } finally {
      setMemberBusy(null);
    }
  }

  // Invite search
  async function doSearch() {
    if (!q.trim()) return;
    try {
      const res = await searchUsers(q.trim());
      setSearchRes(res);
    } catch (e) {
      console.error(e);
      alert(e.message || "Search failed");
    }
  }

  async function doInvite(userId) {
    setInviteBusy(true);
    try {
      await inviteUserToGroup(groupId, userId);
      alert("Invitation sent");
      setQ("");
      setSearchRes([]);
    } catch (e) {
      console.error(e);
      alert(e.message || "Invite failed");
    } finally {
      setInviteBusy(false);
    }
  }

  // Events
  async function createEvent(e) {
    e.preventDefault();
    if (!evTitle.trim() || !evDesc.trim() || !evLoc.trim() || !evDate) {
      alert("Fill all event fields");
      return;
    }
    setBusy(true);
    try {
      await createGroupEvent(groupId, {
        title: evTitle.trim(),
        description: evDesc.trim(),
        location: evLoc.trim(),
        event_date: evDate, // backend accepts "YYYY-MM-DDTHH:mm"
      });
      setEvTitle("");
      setEvDesc("");
      setEvLoc("");
      setEvDate("");
      setEvents(await getGroupEvents(groupId));
    } catch (e) {
      console.error(e);
      alert(e.message || "Create event failed");
    } finally {
      setBusy(false);
    }
  }

  async function respond(eventId, response) {
    setEventBusyId(eventId);
    try {
      await respondToGroupEvent(groupId, eventId, response);
      setEvents(await getGroupEvents(groupId));
      if (openRespEventId === eventId) {
        setOpenResponses(await getGroupEventResponses(groupId, eventId));
      }
    } catch (e) {
      console.error(e);
      alert(e.message || "Respond failed");
    } finally {
      setEventBusyId(null);
    }
  }

  async function openResponsesFor(eventId) {
    setOpenRespEventId(eventId);
    try {
      setOpenResponses(await getGroupEventResponses(groupId, eventId));
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to load responses");
    }
  }

  if (loading) {
    return <div className="main-content"><div className="card"><div className="card-body">Loading...</div></div></div>;
  }
  if (!group) {
    return <div className="main-content"><div className="card"><div className="card-body">Group not found.</div></div></div>;
  }

  const invited = (group.invitation_id || 0) > 0;
  const isPending = group.my_status === "pending";

  return (
    <div className="main-content">
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap" }}>
        <div>
          <h2 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>{group.name}</h2>
          <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
            {group.is_private === 1 ? "Private" : "Public"} • {group.member_count ?? 0} members
          </div>
        </div>

        <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
          <Link to="/groups" className="btn btn-secondary">Back</Link>

          {!group.my_status && !invited && (
            <button className="btn btn-primary" disabled={busy} onClick={handleRequestJoin}>
              {busy ? "..." : "Request"}
            </button>
          )}

          {isPending && <button className="btn btn-secondary" disabled>Requested</button>}

          {isAcceptedMember && (
            <button className="btn btn-outline btn-danger" disabled={busy} onClick={handleLeave}>
              {busy ? "..." : "Leave"}
            </button>
          )}
        </div>
      </div>

      <div className="card" style={{ marginTop: 12 }}>
        <div className="card-body">
          <p style={{ margin: 0 }}>{group.description}</p>
        </div>
      </div>

      {/* Invite users (members only) */}
      {isAcceptedMember && (
        <div className="card" style={{ marginTop: 12 }}>
          <div className="card-body">
            <h4 style={{ marginTop: 0 }}>Invite people</h4>

            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              <input
                className="form-input"
                placeholder="Search users..."
                value={q}
                onChange={(e) => setQ(e.target.value)}
                style={{ flex: 1, minWidth: 220 }}
              />
              <button className="btn btn-secondary" onClick={doSearch} disabled={inviteBusy}>
                Search
              </button>
            </div>

            {searchRes.length > 0 && (
              <div style={{ marginTop: 10 }}>
                {searchRes.map((u) => (
                  <div key={u.id} className="card" style={{ marginBottom: 10 }}>
                    <div className="card-body" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                      <div>
                        <div style={{ fontWeight: 700 }}>{u.full_name || u.nickname || u.email}</div>
                        <div style={{ fontSize: 12, opacity: 0.75 }}>{u.email}</div>
                      </div>
                      <button className="btn btn-primary" disabled={inviteBusy} onClick={() => doInvite(u.id)}>
                        Invite
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Members + pending requests */}
      <div className="section-header" style={{ marginTop: 18 }}>
        <span className="section-title">Members</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      {isAdmin && pendingRequests.length > 0 && (
        <div className="card" style={{ marginBottom: 12 }}>
          <div className="card-body">
            <h4 style={{ marginTop: 0 }}>Join requests</h4>
            {pendingRequests.map((m) => (
              <div key={m.id} className="card" style={{ marginBottom: 10 }}>
                <div className="card-body" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                  <div>
                    <div style={{ fontWeight: 700 }}>{m.user_name}</div>
                    <div style={{ fontSize: 12, opacity: 0.75 }}>{m.role} • {m.status}</div>
                  </div>

                  <div style={{ display: "flex", gap: 8 }}>
                    <button className="btn btn-primary" disabled={memberBusy === m.user_id} onClick={() => handleAccept(m.user_id)}>
                      {memberBusy === m.user_id ? "..." : "Accept"}
                    </button>
                    <button className="btn btn-outline btn-danger" disabled={memberBusy === m.user_id} onClick={() => handleRemove(m.user_id)}>
                      {memberBusy === m.user_id ? "..." : "Decline"}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {acceptedMembers.length === 0 ? (
        <div className="card"><div className="card-body">No members yet.</div></div>
      ) : (
        acceptedMembers.map((m) => (
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
                {busy ? "..." : "Post"}
              </button>
            </form>
          </div>
        </div>
      ) : (
        <div className="card" style={{ marginBottom: 12 }}>
          <div className="card-body" style={{ opacity: 0.8 }}>Join the group to see and create posts.</div>
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

      {/* Events inside group */}
      <div className="section-header" style={{ marginTop: 18 }}>
        <span className="section-title">Events</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      {isAcceptedMember ? (
        <>
          <div className="card" style={{ marginBottom: 12 }}>
            <div className="card-body">
              <h4 style={{ marginTop: 0 }}>Create event</h4>
              <form onSubmit={createEvent} style={{ display: "grid", gap: 10 }}>
                <input className="form-input" placeholder="Title" value={evTitle} onChange={(e) => setEvTitle(e.target.value)} />
                <textarea className="form-input" rows={3} placeholder="Description" value={evDesc} onChange={(e) => setEvDesc(e.target.value)} />
                <input className="form-input" placeholder="Location" value={evLoc} onChange={(e) => setEvLoc(e.target.value)} />
                <input className="form-input" type="datetime-local" value={evDate} onChange={(e) => setEvDate(e.target.value)} />
                <button className="btn btn-primary" disabled={busy}>{busy ? "..." : "Create Event"}</button>
              </form>
            </div>
          </div>

          {events.length === 0 ? (
            <div className="card"><div className="card-body">No events yet.</div></div>
          ) : (
            events.map((ev) => (
              <EventCard
                key={ev.id}
                event={ev}
                busy={eventBusyId === ev.id}
                onRespond={(eventId, response) => respond(eventId, response)}
                onOpenResponses={(eventId) => openResponsesFor(eventId)}
              />
            ))
          )}

          {openRespEventId && (
            <div className="card" style={{ marginTop: 12 }}>
              <div className="card-body">
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12 }}>
                  <h4 style={{ margin: 0 }}>Responses</h4>
                  <button className="btn btn-secondary" onClick={() => { setOpenRespEventId(null); setOpenResponses([]); }}>
                    Close
                  </button>
                </div>

                {openResponses.length === 0 ? (
                  <div style={{ marginTop: 10, opacity: 0.8 }}>No responses yet.</div>
                ) : (
                  <div style={{ marginTop: 10 }}>
                    {openResponses.map((r) => (
                      <div key={r.id} className="card" style={{ marginBottom: 10 }}>
                        <div className="card-body" style={{ display: "flex", justifyContent: "space-between" }}>
                          <div style={{ fontWeight: 700 }}>{r.user_name}</div>
                          <div style={{ opacity: 0.85 }}>{r.response.replace("_", " ")}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </>
      ) : (
        <div className="card"><div className="card-body">Join the group to see events.</div></div>
      )}

      {/* Group Chat - only for accepted members */}
      {isAcceptedMember && (
        <div style={{ marginTop: '24px' }}>
          <GroupChat groupId={groupId} />
        </div>
      )}
    </div>
  );
}
