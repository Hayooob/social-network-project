import React from "react";
import { Link } from "react-router-dom";

export default function GroupCard({
  group,
  busy,
  onRequest,
  onLeave,
  onAcceptInvite,
  onDeclineInvite,
}) {
  const invited = (group?.invitation_id || 0) > 0;
  const inviteId = group?.invitation_id || 0;

  const status = group?.my_status || ""; // accepted/pending/""
  const isAccepted = status === "accepted";
  const isPending = status === "pending";

  return (
    <div className="card" style={{ marginBottom: 12 }}>
      <div className="card-body">
        <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Link to={`/groups/${group.id}`} style={{ textDecoration: "none" }}>
              <h3 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>{group.name}</h3>
            </Link>
            <p style={{ marginTop: 6, opacity: 0.8 }}>{group.description}</p>

            <div style={{ display: "flex", gap: 10, fontSize: 12, opacity: 0.75 }}>
              <span>{group.member_count ?? 0} members</span>
              <span>•</span>
              <span>{group.is_private === 1 ? "Private" : "Public"}</span>
              {invited && (
                <>
                  <span>•</span>
                  <span style={{ fontWeight: 700 }}>Invited</span>
                </>
              )}
              {isPending && !invited && (
                <>
                  <span>•</span>
                  <span>Requested</span>
                </>
              )}
            </div>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", justifyContent: "flex-end" }}>
            {invited && (
              <>
                <button className="btn btn-primary" disabled={busy} onClick={() => onAcceptInvite?.(inviteId)}>
                  {busy ? "..." : "Accept"}
                </button>
                <button className="btn btn-outline btn-danger" disabled={busy} onClick={() => onDeclineInvite?.(inviteId)}>
                  {busy ? "..." : "Decline"}
                </button>
              </>
            )}

            {!invited && !status && (
              <button className="btn btn-primary" disabled={busy} onClick={() => onRequest?.(group.id)}>
                {busy ? "..." : "Request"}
              </button>
            )}

            {!invited && isPending && (
              <button className="btn btn-secondary" disabled>
                Requested
              </button>
            )}

            {isAccepted && (
              <>
                <Link to={`/groups/${group.id}`} className="btn btn-secondary">
                  Enter
                </Link>
                <button className="btn btn-outline btn-danger" disabled={busy} onClick={() => onLeave?.(group.id)}>
                  {busy ? "..." : "Leave"}
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
