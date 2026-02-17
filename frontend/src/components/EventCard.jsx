import React from "react";
import EventResponseButtons from "./EventResponseButtons";

export default function EventCard({ event, busy, onRespond, onOpenResponses }) {
  const dateStr = event?.event_date ? new Date(event.event_date).toLocaleString() : "";

  return (
    <div className="card" style={{ marginBottom: 12 }}>
      <div className="card-body">
        <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap" }}>
          <div style={{ flex: 1, minWidth: 240 }}>
            <h4 style={{ margin: 0 }}>{event.title}</h4>
            <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
              <div>{dateStr}</div>
              <div>{event.location}</div>
            </div>

            <div style={{ display: "flex", gap: 10, marginTop: 10, fontSize: 12, opacity: 0.85, flexWrap: "wrap" }}>
              <span>{event.going_count ?? 0} going</span>
              <span>•</span>
              <span>{event.maybe_count ?? 0} maybe</span>
              <span>•</span>
              <span>{event.not_going_count ?? 0} not going</span>
              <button
                className="btn btn-secondary"
                style={{ marginLeft: 10 }}
                onClick={() => onOpenResponses?.(event.id)}
              >
                View
              </button>
            </div>
          </div>

          <div style={{ minWidth: 240 }}>
            <EventResponseButtons
              value={event.my_response || ""}
              busy={busy}
              onChange={(resp) => onRespond?.(event.id, resp)}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
