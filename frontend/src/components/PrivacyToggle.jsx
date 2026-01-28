import React, { useState } from "react";
import { togglePrivacy } from "../api/users";
import { useAuth } from "../VerifyAuth";

export default function PrivacyToggle() {
  const { user, setUser } = useAuth();
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");

  if (!user) return null;

  const label = user.is_private ? "Private" : "Public";

  const onToggle = async () => {
    setErr("");
    setLoading(true);
    try {
      const res = await togglePrivacy(); // { is_private: true/false }
      setUser({ ...user, is_private: res.is_private });
    } catch (e) {
      setErr(e?.message || "Failed to toggle privacy");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ marginTop: "10px" }}>
      <div style={{ display: "flex", alignItems: "center", gap: "10px", flexWrap: "wrap" }}>
        <span style={{ fontSize: "12px", opacity: 0.8 }}>
          <strong>Profile:</strong> {label}
        </span>

        <button className="btn btn-outline" onClick={onToggle} disabled={loading}>
          {loading ? "Updating..." : "Toggle Privacy"}
          <span>↗</span>
        </button>
      </div>

      {err && <p style={{ color: "crimson", marginTop: "8px" }}>{err}</p>}
    </div>
  );
}
