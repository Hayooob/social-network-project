import { get, post } from "./ftchclient";

export function createGroup({ name, description, is_private }) {
  return post("/api/groups", { name, description, is_private });
}

export async function getMyGroups() {
  const data = await get("/api/groups");
  return data || [];
}

export function getGroupDetails(id) {
  return get(`/api/groups/${id}`);
}

export function joinGroup(id) {
  return post(`/api/groups/${id}/join`);
}

export function leaveGroup(id) {
  return post(`/api/groups/${id}/leave`);
}

export async function getGroupMembers(id) {
  const data = await get(`/api/groups/${id}/members`);
  return data || [];
}

export function acceptMember(groupId, userId) {
  return post(`/api/groups/${groupId}/members/${userId}/accept`);
}

export function removeMember(groupId, userId) {
  return post(`/api/groups/${groupId}/members/${userId}/remove`);
}

export async function getGroupPosts(groupId) {
  const data = await get(`/api/groups/${groupId}/posts`);
  return data || [];
}

export function createGroupPost(groupId, content) {
  return post(`/api/groups/${groupId}/posts`, { content });
}
