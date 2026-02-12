import React, { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createEvent } from "../api/events";
import { getMyGroups } from "../api/groups";

export default function CreateEventPage() {
  const navigate = useNavigate();
  const [groups, setGroups] = useState([]);

  const [groupId, setGroupId] = useState(""); // "" means global
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [location, setLocation] = useState("");
  const [eventDate, setEventDate] = useState(""); // "YYYY-MM-DDTHH:mm"
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const g = await getMyGroups();
        setGroups(g);
      } catch (e) {
        console.error(e);
      }
    })();
  }, []);

  const acceptedGroups = useMemo(
    () => groups.filter((g) => g.my_status === "accepted"),
    [groups]
  );

  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    try {
      const payload = {
        group_id: groupId ? Number(groupId) : null,
        title,
        description,
        location,
        // backend accepts RFC3339 OR "YYYY-MM-DDTHH:mm"
        event_date: eventDate,
      };
      const ev = await createEvent(payload);
      navigate(`/events/${ev.id}`);
    } catch (err) {
      console.error(err);
      alert(err.message || "Failed to create event");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="main-content">
      <div className="section-header">
        <span className="section-title">Create Event</span>
        <div className="section-line"></div>
        <span className="section-star star-spin">✦</span>
      </div>

      <div className="card">
        <div className="card-body">
          <form onSubmit={submit} style={{ display: "grid", gap: 12 }}>
            <div>
              <label className="form-label">Event Title</label>
              <input className="form-input" value={title} onChange={(e) => setTitle(e.target.value)} />
            </div>

            <div>
              <label className="form-label">Description</label>
              <textarea className="form-input" rows={4} value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>

            <div>
              <label className="form-label">Location</label>
              <input className="form-input" value={location} onChange={(e) => setLocation(e.target.value)} />
            </div>

            <div>
              <label className="form-label">Date & Time</label>
              <input
                className="form-input"
                type="datetime-local"
                value={eventDate}
                onChange={(e) => setEventDate(e.target.value)}
              />
            </div>

            <div>
              <label className="form-label">Group (optional)</label>
              <select className="form-input" value={groupId} onChange={(e) => setGroupId(e.target.value)}>
                <option value="">No group (global event)</option>
                {acceptedGroups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
              <div style={{ fontSize: 12, opacity: 0.7, marginTop: 6 }}>
                Only groups you are an accepted member of appear here.
              </div>
            </div>

            <button className="btn btn-primary" disabled={busy}>
              {busy ? "Creating..." : "Create Event"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
