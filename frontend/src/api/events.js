import { get, post } from "./ftchclient";

export function createEvent(payload) {
  return post("/api/events", payload);
}

export async function getUpcomingEvents() {
  const data = await get("/api/events");
  return data || [];
}

export function getEventDetails(id) {
  return get(`/api/events/${id}`);
}

export function respondToEvent(id, response) {
  return post(`/api/events/${id}/respond`, { response });
}

export async function getEventResponses(id) {
  const data = await get(`/api/events/${id}/responses`);
  return data || [];
}
