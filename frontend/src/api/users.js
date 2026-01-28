import { get, post } from "./ftchclient";

// GET /api/users/{id}
export async function getUserProfile(userId) {
  return get(`/api/users/${userId}`);
}

// POST /api/me/privacy
export async function togglePrivacy() {
  return post(`/api/me/privacy`);
}

