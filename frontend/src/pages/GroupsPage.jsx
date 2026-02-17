import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import GroupCard from "../components/GroupCard";
import {
  acceptGroupInvitation,
  declineGroupInvitation,
  getVisibleGroups,
  requestJoinGroup,
  leaveGroup,
} from "../api/groups";

export default function GroupsPage() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busyKey, setBusyKey] = useState("");

  async function load() {
    setLoading(true);
    try {
      const data = await getVisibleGroups();
      setGroups(data);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to load groups");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function doAction(key, fn) {
    setBusyKey(key);
    try {
      await fn();
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Action failed");
    } finally {
      setBusyKey("");
    }
  }

  return (
    <div className="main-content">
      <div className="section-header">
        <span className="section-title">Groups</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12, marginBottom: 12 }}>
        <p style={{ margin: 0, opacity: 0.75 }}>
          Groups from people you follow + groups you’re involved in.
        </p>
        <Link to="/groups/create" className="btn btn-primary">
          Create Group <span>↗</span>
        </Link>
      </div>

      {loading ? (
        <div className="card"><div className="card-body">Loading...</div></div>
      ) : groups.length === 0 ? (
        <div className="card"><div className="card-body">No groups available.</div></div>
      ) : (
        groups.map((g) => (
          <GroupCard
            key={g.id}
            group={g}
            busy={busyKey === `g:${g.id}` || busyKey === `i:${g.invitation_id}`}
            onRequest={(id) => doAction(`g:${id}`, () => requestJoinGroup(id))}
            onLeave={(id) => doAction(`g:${id}`, () => leaveGroup(id))}
            onAcceptInvite={(invId) => doAction(`i:${invId}`, () => acceptGroupInvitation(invId))}
            onDeclineInvite={(invId) => doAction(`i:${invId}`, () => declineGroupInvitation(invId))}
          />
        ))
      )}
    </div>
  );
}
