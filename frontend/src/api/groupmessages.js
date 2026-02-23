import { get, post } from "./ftchclient";

// GET /api/groups/{groupId}/messages
export async function getGroupMessages(groupId, limit = 50, offset = 0) {
  const data = await get(`/api/groups/${groupId}/messages?limit=${limit}&offset=${offset}`);
  return data || [];
}

// POST /api/groups/{groupId}/messages
export async function sendGroupMessage(groupId, content) {
  return post(`/api/groups/${groupId}/messages`, { content });
}