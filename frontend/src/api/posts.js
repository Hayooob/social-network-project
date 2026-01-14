import { post, get } from './ftchclient';

export async function getFeed() {
  try {
    const data = await get('/api/feed');
    return data || [];
  } catch (err) {
    console.error('getFeed error:', err);
    return [];
  }
}

export async function createPost(content, privacy = 'public') {
  return post('/api/posts', { content, privacy });
}

export async function getMyPosts() {
  try {
    const data = await get('/api/me/posts');
    return data || [];
  } catch (err) {
    console.error('getMyPosts error:', err);
    return [];
  }
}