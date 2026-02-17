import { get, post } from "./ftchclient";

// GET /api/users/{id}
export async function getUserProfile(userId) {
  return get(`/api/users/${userId}`);
}

// POST /api/me/privacy
export async function togglePrivacy() {
  return post(`/api/me/privacy`);
}

// GET /api/users/search?q=...
export async function searchUsers(q) {
  const data = await get(`/api/users/search?q=${encodeURIComponent(q)}`);
  return data || [];
}
