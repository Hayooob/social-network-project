import { post, get } from './ftchclient';

export async function getFeed() {
  return get('/api/feed');
}

export async function createPost(content) {
  return post('/api/posts', { content });
}

export async function getMyPosts() {
  return get('/api/me/posts');
}
