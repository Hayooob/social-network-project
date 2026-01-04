import { get, post } from "./ftchclient";

// POST /api/login
export async function login(email, password) {
  const user = await post("/api/login", { email, password });
  return user; n
}

// POST /api/register
export async function register({ username, email, password, confirmPassword }) {
  const user = await post("/api/register", {
    username,
    email,
    password,
    confirmPassword,
  });
  return user;
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
