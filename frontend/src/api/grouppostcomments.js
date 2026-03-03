import { get, post } from "./ftchclient";

// GET /api/groups/{groupId}/posts/{postId}/comments
export async function getGroupPostComments(groupId, postId) {
    return get(`/api/groups/${groupId}/posts/${postId}/comments`);
}

// POST /api/groups/{groupId}/posts/{postId}/comments
export async function addGroupPostComment(groupId, postId, content, imagePath = "") {
    return post(`/api/groups/${groupId}/posts/${postId}/comments`, {
        content,
        image_path: imagePath,
    });
}
