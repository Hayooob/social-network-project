import { get, post } from "./ftchclient";

// POST /api/login
export async function login(email, password) {
  const user = await post("/api/login", { email, password });
  return user;
}

// POST /api/register
export async function register({
  firstName,
  lastName,
  dateOfBirth,
  email,
  password,
  confirmPassword,
  nickname,
  aboutMe,
  avatarFile,
}) {
  // If there's an avatar, use multipart/form-data.
  if (avatarFile) {
    const fd = new FormData();
    fd.append("first_name", firstName || "");
    fd.append("last_name", lastName || "");
    fd.append("date_of_birth", dateOfBirth || "");
    fd.append("email", email || "");
    fd.append("password", password || "");
    fd.append("confirmPassword", confirmPassword || "");
    if (nickname) fd.append("nickname", nickname);
    if (aboutMe) fd.append("about_me", aboutMe);
    fd.append("avatar", avatarFile);
    return post("/api/register", fd);
  }

  // Otherwise JSON.
  return post("/api/register", {
    first_name: firstName,
    last_name: lastName,
    date_of_birth: dateOfBirth,
    email,
    password,
    confirmPassword,
    nickname,
    about_me: aboutMe,
  });
}

// GET /api/me
export async function getCurrentUser() {
  try {
    const user = await get("/api/me");
    return user;
  } catch (err) {
    if (err.status === 401) {
      return null;
    }
    throw err;
  }
}

// POST /api/logout
export async function logout() {
  try {
    await post("/api/logout");
  } catch (err) {
    if (err.status === 401 || err.status === 404) return;
    throw err;
  }
}

// GET /api/users/suggestions
export async function getSuggestedUsers() {
  try {
    const users = await get("/api/users/suggestions");
    return users || [];
  } catch (err) {
    console.error("getSuggestedUsers error:", err);
    return [];
  }
}