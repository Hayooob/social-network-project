import { get, post } from "./ftchclient";

// GET /api/groups/{groupId}/messages
export async function getGroupMessages(groupId, limit = 50, offset = 0) {
  const data = await get(`/api/groups/${groupId}/messages?limit=${limit}&offset=${offset}`);
  return data || [];
}

// POST /api/groups/{groupId}/messages - with optional image
export async function sendGroupMessage(groupId, content, imageFile = null) {
  if (imageFile) {
    // Use FormData for file upload
    const formData = new FormData();
    formData.append("content", content);
    formData.append("image", imageFile);

    return post(`/api/groups/${groupId}/messages`, formData);
  } else {
    // Use JSON for text-only messages
    return post(`/api/groups/${groupId}/messages`, { content });
  }
}