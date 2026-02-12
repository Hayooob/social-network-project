import React from "react";
import { Link } from "react-router-dom";
import EventResponseButtons from "./EventResponseButtons";

export default function EventCard({ event, onRespond, busy }) {
  const dateStr = event?.event_date ? new Date(event.event_date).toLocaleString() : "";

  return (
    <div className="card" style={{ marginBottom: 12 }}>
      <div className="card-body">
        <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Link to={`/events/${event.id}`} style={{ textDecoration: "none" }}>
              <h3 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>
                {event.title}
              </h3>
            </Link>
            <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
              <div>{dateStr}</div>
              <div>{event.location}</div>
              {event.group_name && <div>Group: {event.group_name}</div>}
            </div>

            <div style={{ display: "flex", gap: 10, marginTop: 10, fontSize: 12, opacity: 0.8 }}>
              <span>{event.going_count ?? 0} going</span>
              <span>•</span>
              <span>{event.maybe_count ?? 0} maybe</span>
              <span>•</span>
              <span>{event.not_going_count ?? 0} not going</span>
            </div>
          </div>

          <div style={{ minWidth: 220 }}>
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
