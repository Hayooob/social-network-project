import React from "react";
import { Link } from "react-router-dom";

export default function GroupCard({ group, onJoin, onLeave, busy }) {
  const status = group?.my_status || "";
  const isAccepted = status === "accepted";
  const isPending = status === "pending";

  return (
    <div className="card" style={{ marginBottom: 12 }}>
      <div className="card-body">
        <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Link to={`/groups/${group.id}`} style={{ textDecoration: "none" }}>
              <h3 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>
                {group.name}
              </h3>
            </Link>
            <p style={{ marginTop: 6, opacity: 0.8 }}>{group.description}</p>

            <div style={{ display: "flex", gap: 10, fontSize: 12, opacity: 0.75 }}>
              <span>{group.member_count ?? 0} members</span>
              <span>•</span>
              <span>{group.is_private === 1 ? "Private" : "Public"}</span>
              {status && (
                <>
                  <span>•</span>
                  <span>Status: {status}</span>
                </>
              )}
            </div>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            {!status && (
              <button className="btn btn-primary" onClick={() => onJoin?.(group.id)} disabled={busy}>
                {busy ? "Joining..." : "Join"}
              </button>
            )}

            {isPending && (
              <button className="btn btn-secondary" disabled>
                Pending
              </button>
            )}

            {isAccepted && (
              <button className="btn btn-outline btn-danger" onClick={() => onLeave?.(group.id)} disabled={busy}>
                {busy ? "Leaving..." : "Leave"}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
