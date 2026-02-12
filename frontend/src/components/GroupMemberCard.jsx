import React from "react";

export default function GroupMemberCard({ member, isAdmin, onRemove, busy }) {
  return (
    <div className="card" style={{ marginBottom: 10 }}>
      <div className="card-body" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <div>
          <div style={{ fontWeight: 700 }}>{member.user_name}</div>
          <div style={{ fontSize: 12, opacity: 0.75 }}>
            {member.role} • {member.status}
          </div>
        </div>

        {isAdmin && member.status === "accepted" && member.role !== "admin" && (
          <button
            className="btn btn-outline btn-danger"
            onClick={() => onRemove?.(member.user_id)}
            disabled={busy}
          >
            {busy ? "Removing..." : "Remove"}
          </button>
        )}
      </div>
    </div>
  );
}
