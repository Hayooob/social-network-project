import { get, post } from "./ftchclient";

// GET /api/users/{id} - get another user's profile
export async function getUserProfile(userId) {
  return get(`/api/users/${userId}`);
}

// POST /api/me/privacy - toggle your own privacy setting
export async function togglePrivacy() {
  return post("/api/me/privacy");
}

// GET /api/users/search?q=query - search for users
export async function searchUsers(query) {
  if (!query || query.length < 1) {
    return [];
  }
  try {
    const results = await get(`/api/users/search?q=${encodeURIComponent(query)}`);
    return results || [];
  } catch (err) {
    console.error("searchUsers error:", err);
    return [];
  }
}