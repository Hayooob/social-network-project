import React from "react";

export default function EventResponseButtons({ value, onChange, busy }) {
  const btnStyle = (active) => ({
    opacity: active ? 1 : 0.65,
    borderWidth: active ? 2 : 1,
  });

  return (
    <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
      <button
        className="btn btn-primary"
        style={btnStyle(value === "going")}
        disabled={busy}
        onClick={() => onChange?.("going")}
      >
        Going
      </button>
      <button
        className="btn btn-secondary"
        style={btnStyle(value === "maybe")}
        disabled={busy}
        onClick={() => onChange?.("maybe")}
      >
        Maybe
      </button>
      <button
        className="btn btn-outline btn-danger"
        style={btnStyle(value === "not_going")}
        disabled={busy}
        onClick={() => onChange?.("not_going")}
      >
        Not Going
      </button>
    </div>
  );
}
