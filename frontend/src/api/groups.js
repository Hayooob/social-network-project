import { get, post } from "./ftchclient";

// Groups list uses "visible groups" rule (followers + involved)
export async function getVisibleGroups() {
  const data = await get("/api/groups");
  return data || [];
}

export function createGroup({ name, description, is_private }) {
  return post("/api/groups", { name, description, is_private });
}

export function getGroupDetails(id) {
  return get(`/api/groups/${id}`);
}

// Join = request join (public auto accept, private -> pending)
export function requestJoinGroup(id) {
  return post(`/api/groups/${id}/join`);
}

export function leaveGroup(id) {
  return post(`/api/groups/${id}/leave`);
}

// Invitations
export async function getMyGroupInvitations() {
  const data = await get("/api/groups/invitations");
  return data || [];
}

export function acceptGroupInvitation(invId) {
  return post(`/api/groups/invitations/${invId}/accept`);
}

export function declineGroupInvitation(invId) {
  return post(`/api/groups/invitations/${invId}/decline`);
}

// Members / requests
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

// Invite user
export function inviteUserToGroup(groupId, userId) {
  return post(`/api/groups/${groupId}/invite`, { user_id: userId });
}

// Posts (keep)
export async function getGroupPosts(groupId) {
  const data = await get(`/api/groups/${groupId}/posts`);
  return data || [];
}

export function createGroupPost(groupId, content) {
  return post(`/api/groups/${groupId}/posts`, { content });
}
