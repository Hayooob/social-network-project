import { get, post } from "./ftchclient";

export async function getFeed() {
  const data = await get("/api/feed");
  return data || [];
}

export async function createPost(
  content,
  privacy = "public",
  imageFile = null,
  allowedViewers = []
) {
  if (imageFile) {
    const fd = new FormData();
    fd.append("content", content);
    fd.append("privacy", privacy);
    fd.append("image", imageFile); 
    if (allowedViewers && allowedViewers.length > 0) {
      fd.append("allowed_viewers", JSON.stringify(allowedViewers));
    }
    return post("/api/posts", fd);
  }

  return post("/api/posts", {
    content,
    privacy,
    allowed_viewers: allowedViewers,
  });
}

export async function getMyPosts() {
  const data = await get("/api/me/posts");
  return data || [];
}

export async function listComments(postId) {
  return get(`/api/posts/${postId}/comments`);
}

export async function createComment(postId, content) {
  return post(`/api/posts/${postId}/comments`, { content });
}

export async function toggleLike(postId) {
  return post(`/api/posts/${postId}/like`);
}
