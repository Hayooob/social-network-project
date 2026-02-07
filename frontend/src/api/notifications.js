import { get, post } from "./ftchclient";

// GET /api/notifications - get all notifications
export async function getNotifications(limit = 50, offset = 0) {
  const notifications = await get(`/api/notifications?limit=${limit}&offset=${offset}`);
  return notifications || [];
}

// POST /api/notifications/{id}/read - mark one notification as read
export async function markNotificationRead(notifId) {
  return await post(`/api/notifications/${notifId}/read`);
}

// POST /api/notifications/read-all - mark all notifications as read
export async function markAllNotificationsRead() {
  return await post("/api/notifications/read-all");
}

// GET /api/notifications/unread-count - get unread notification count
export async function getUnreadNotificationCount() {
  try {
    const result = await get("/api/notifications/unread-count");
    return result?.count || 0;
  } catch (err) {
    console.error("getUnreadNotificationCount error:", err);
    return 0;
  }
}