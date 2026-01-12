import React, { useEffect, useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { getMyPosts } from '../api/posts';
import PostCard from '../components/PostCard';

export default function ProfilePage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchMyPosts = async () => {
    setLoading(true);
    try {
      const data = await getMyPosts();
      setPosts(data);
    } catch (err) {
      console.error('Error loading user posts:', err);
      alert('Failed to load posts.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user) fetchMyPosts();
  }, [user]);

  if (!user) return <p>Loading user info...</p>;

  return (
    <div className="max-w-2xl mx-auto mt-6">
      <div className="mb-6 p-4 border rounded bg-white shadow-sm">
        <h2 className="text-xl font-semibold">{user.name}</h2>
        <p className="text-gray-600">{user.email}</p>
      </div>

      <h3 className="text-lg font-semibold mb-4">My Posts</h3>
      {loading ? (
        <p>Loading posts...</p>
      ) : posts.length === 0 ? (
        <p>No posts yet.</p>
      ) : (
        posts.map((post) => <PostCard key={post.id} post={post} />)
      )}
    </div>
  );
}
