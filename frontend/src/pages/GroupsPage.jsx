import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import GroupCard from "../components/GroupCard";
import { getMyGroups, joinGroup, leaveGroup } from "../api/groups";

export default function GroupsPage() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState(null);

  async function load() {
    setLoading(true);
    try {
      const data = await getMyGroups();
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

  async function handleJoin(id) {
    setBusyId(id);
    try {
      await joinGroup(id);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to join group");
    } finally {
      setBusyId(null);
    }
  }

  async function handleLeave(id) {
    setBusyId(id);
    try {
      await leaveGroup(id);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to leave group");
    } finally {
      setBusyId(null);
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
        <p style={{ margin: 0, opacity: 0.75 }}>Your groups (accepted & pending).</p>
        <Link to="/groups/create" className="btn btn-primary">
          Create Group <span>↗</span>
        </Link>
      </div>

      {loading ? (
        <div className="card"><div className="card-body">Loading...</div></div>
      ) : groups.length === 0 ? (
        <div className="card">
          <div className="card-body text-center">
            <p style={{ marginBottom: 10 }}>No groups yet.</p>
            <Link to="/groups/create" className="btn btn-primary">Create your first group</Link>
          </div>
        </div>
      ) : (
        groups.map((g) => (
          <GroupCard
            key={g.id}
            group={g}
            busy={busyId === g.id}
            onJoin={handleJoin}
            onLeave={handleLeave}
          />
        ))
      )}
    </div>
  );
}
