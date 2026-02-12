import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getUpcomingEvents, respondToEvent } from "../api/events";
import EventCard from "../components/EventCard";

export default function EventsPage() {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState(null);

  async function load() {
    setLoading(true);
    try {
      const data = await getUpcomingEvents();
      setEvents(data);
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to load events");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleRespond(eventId, response) {
    setBusyId(eventId);
    try {
      await respondToEvent(eventId, response);
      await load();
    } catch (e) {
      console.error(e);
      alert(e.message || "Failed to respond");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="main-content">
      <div className="section-header">
        <span className="section-title">Events</span>
        <div className="section-line"></div>
        <span className="section-star star-spin-reverse">✳</span>
      </div>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12, marginBottom: 12 }}>
        <p style={{ margin: 0, opacity: 0.75 }}>Upcoming events (global + your groups).</p>
        <Link to="/events/create" className="btn btn-primary">
          Create Event <span>↗</span>
        </Link>
      </div>

      {loading ? (
        <div className="card"><div className="card-body">Loading...</div></div>
      ) : events.length === 0 ? (
        <div className="card"><div className="card-body">No upcoming events.</div></div>
      ) : (
        events.map((ev) => (
          <EventCard key={ev.id} event={ev} busy={busyId === ev.id} onRespond={handleRespond} />
        ))
      )}
    </div>
  );
}
