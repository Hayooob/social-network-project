import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getEventDetails, getEventResponses, respondToEvent } from "../api/events";
import EventResponseButtons from "../components/EventResponseButtons";

export default function EventDetailPage() {
  const { id } = useParams();
  const eventId = Number(id);

  const [event, setEvent] = useState(null);
  const [responses, setResponses] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const [ev, resps] = await Promise.all([
        getEventDetails(eventId),
        getEventResponses(eventId),
      ]);
      setEvent(ev);
      setResponses(resps);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to load event");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, [eventId]);

  async function onChange(resp) {
    setBusy(true);
    try {
      await respondToEvent(eventId, resp);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to respond");
    } finally {
      setBusy(false);
    }
  }

  if (loading) {
    return (
      <div className="main-content">
        <div className="card"><div className="card-body">Loading...</div></div>
      </div>
    );
  }

  if (!event) {
    return (
      <div className="main-content">
        <div className="card"><div className="card-body">Event not found.</div></div>
      </div>
    );
  }

  return (
    <div className="main-content">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12 }}>
        <div>
          <h2 style={{ margin: 0, fontFamily: "'Cormorant Garamond', serif" }}>{event.title}</h2>
          <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
            {event.event_date ? new Date(event.event_date).toLocaleString() : ""} • {event.location}
            {event.group_name ? ` • Group: ${event.group_name}` : ""}
          </div>
        </div>
        <Link to="/events" className="btn btn-secondary">Back</Link>
      </div>

      <div className="card" style={{ marginTop: 12 }}>
        <div className="card-body">
          <p style={{ marginTop: 0 }}>{event.description}</p>

          <div style={{ display: "flex", gap: 10, marginTop: 12, fontSize: 12, opacity: 0.85 }}>
            <span>{event.going_count ?? 0} going</span>
            <span>•</span>
            <span>{event.maybe_count ?? 0} maybe</span>
            <span>•</span>
            <span>{event.not_going_count ?? 0} not going</span>
          </div>

          <div style={{ marginTop: 12 }}>
            <EventResponseButtons value={event.my_response || ""} onChange={onChange} busy={busy} />
          </div>
        </div>
      </div>

      <div className="section-header" style={{ marginTop: 18 }}>
        <span className="section-title">Responses</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      {responses.length === 0 ? (
        <div className="card"><div className="card-body">No responses yet.</div></div>
      ) : (
        responses.map((r) => (
          <div key={r.id} className="card" style={{ marginBottom: 10 }}>
            <div className="card-body" style={{ display: "flex", justifyContent: "space-between" }}>
              <div style={{ fontWeight: 700 }}>{r.user_name}</div>
              <div style={{ textTransform: "capitalize", opacity: 0.85 }}>{r.response.replace("_", " ")}</div>
            </div>
          </div>
        ))
      )}
    </div>
  );
}
