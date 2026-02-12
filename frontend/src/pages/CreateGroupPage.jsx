import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createGroup } from "../api/groups";

export default function CreateGroupPage() {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const [busy, setBusy] = useState(false);

  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    try {
      const g = await createGroup({ name, description, is_private: isPrivate });
      navigate(`/groups/${g.id}`);
    } catch (err) {
      console.error(err);
      alert(err.message || "Failed to create group");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="main-content">
      <div className="section-header">
        <span className="section-title">Create Group</span>
        <div className="section-line"></div>
        <span className="section-star star-spin">✦</span>
      </div>

      <div className="card">
        <div className="card-body">
          <form onSubmit={submit} style={{ display: "grid", gap: 12 }}>
            <div>
              <label className="form-label">Name</label>
              <input className="form-input" value={name} onChange={(e) => setName(e.target.value)} />
            </div>

            <div>
              <label className="form-label">Description</label>
              <textarea className="form-input" rows={4} value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>

            <label style={{ display: "flex", gap: 10, alignItems: "center" }}>
              <input type="checkbox" checked={isPrivate} onChange={(e) => setIsPrivate(e.target.checked)} />
              Private group (join requests must be accepted)
            </label>

            <button className="btn btn-primary" disabled={busy}>
              {busy ? "Creating..." : "Create"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
