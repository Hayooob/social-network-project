import { get, post } from "./ftchclient";

export async function followUser(userId) {
  return post("/api/users/" + userId + "/follow");
}

export async function unfollowUser(userId) {
  return post("/api/users/" + userId + "/unfollow");
}

export async function getMyFollowers() {
  try {
    const data = await get("/api/me/followers");
    return data || [];
  } catch (err) {
    if (err.status === 401) {
      return [];
    }
    throw err;
  }
}

export async function getMyFollowing() {
  try {
    const data = await get("/api/me/following");
    return data || [];
  } catch (err) {
    if (err.status === 401) {
      return [];
    }
    throw err;
  }
}

export async function getPendingRequests() {
  try {
    const data = await get("/api/me/follow-requests");
    return data || [];
  } catch (err) {
    if (err.status === 401) {
      return [];
    }
    throw err;
  }
}

export async function acceptFollowRequest(userId) {
  return post("/api/follow-requests/" + userId + "/accept");
}

export async function declineFollowRequest(userId) {
  return post("/api/follow-requests/" + userId + "/decline");
}

export async function getFollowCounts() {
  try {
    const data = await get("/api/me/follow-counts");
    return data || { followers: 0, following: 0 };
  } catch (err) {
    if (err.status === 401) {
      return { followers: 0, following: 0 };
    }
    throw err;
  }
}