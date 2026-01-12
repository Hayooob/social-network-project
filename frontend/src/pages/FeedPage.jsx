import React, { useEffect, useState } from 'react';
import { getFeed } from '../api/posts';
import PostForm from '../components/PostForm';
import PostCard from '../components/PostCard';

export default function FeedPage() {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchFeed = async () => {
    setLoading(true);
    try {
      const data = await getFeed();
      setPosts(data);
    } catch (err) {
      console.error('Error loading feed:', err);
      alert('Failed to load feed.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFeed();
  }, []);

  return (
    <div className="max-w-2xl mx-auto mt-6">
      <PostForm onPostCreated={fetchFeed} />
      {loading ? (
        <p>Loading feed...</p>
      ) : posts.length === 0 ? (
        <p>No posts yet.</p>
      ) : (
        posts.map((post) => <PostCard key={post.id} post={post} />)
      )}
    </div>
  );
}
