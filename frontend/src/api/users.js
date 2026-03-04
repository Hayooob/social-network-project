import { get, post, put } from "./ftchclient";

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

// PUT /api/me - update current user's profile info
export async function updateProfile({ full_name, date_of_birth, nickname, about_me, avatarFile }) {
  if (avatarFile) {
    const fd = new FormData();
    if (full_name !== undefined) fd.append("full_name", full_name);
    if (date_of_birth !== undefined) fd.append("date_of_birth", date_of_birth);
    if (nickname !== undefined) fd.append("nickname", nickname);
    if (about_me !== undefined) fd.append("about_me", about_me);
    fd.append("avatar", avatarFile);
    return put("/api/me", fd);
  }
  const body = {};
  if (full_name !== undefined) body.full_name = full_name;
  if (date_of_birth !== undefined) body.date_of_birth = date_of_birth;
  if (nickname !== undefined) body.nickname = nickname;
  if (about_me !== undefined) body.about_me = about_me;
  return put("/api/me", body);
}