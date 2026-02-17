import { get, post } from "./ftchclient";

export async function getGroupEvents(groupId) {
  const data = await get(`/api/groups/${groupId}/events`);
  return data || [];
}

export function createGroupEvent(groupId, payload) {
  return post(`/api/groups/${groupId}/events`, payload);
}

export function respondToGroupEvent(groupId, eventId, response) {
  return post(`/api/groups/${groupId}/events/${eventId}/respond`, { response });
}

export async function getGroupEventResponses(groupId, eventId) {
  const data = await get(`/api/groups/${groupId}/events/${eventId}/responses`);
  return data || [];
}
