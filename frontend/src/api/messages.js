import { get, post } from "./ftchclient";

// GET /api/messages - get list of all conversations
export async function getConversations() {
  const conversations = await get("/api/messages");
  return conversations || [];
}

// GET /api/messages/{userId} - get messages with a specific user
export async function getConversation(userId, limit = 50, offset = 0) {
  const messages = await get(`/api/messages/${userId}?limit=${limit}&offset=${offset}`);
  return messages || [];
}

// POST /api/messages/{userId} - send a message (HTTP fallback)
export async function sendMessage(userId, content) {
  const message = await post(`/api/messages/${userId}`, { content });
  return message;
}

// GET /api/messages/unread-count - get unread message count
export async function getUnreadMessageCount() {
  try {
    const result = await get("/api/messages/unread-count");
    return result?.count || 0;
  } catch (err) {
    console.error("getUnreadMessageCount error:", err);
    return 0;
  }
}