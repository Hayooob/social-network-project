import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { getMyPosts } from '../api/posts';
import { getFollowCounts } from '../api/followers';
import PostCard from '../components/PostCard';

export default function ProfilePage() {
  const { user } = useAuth();
  const [posts, setPosts] = useState([]);
  const [counts, setCounts] = useState({ followers: 0, following: 0 });
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    setLoading(true);
    try {
      const postsData = await getMyPosts();
      const countsData = await getFollowCounts();
      setPosts(postsData);
      setCounts(countsData);
    } catch (err) {
      console.error('Error loading profile data:', err);
      alert('Failed to load profile data.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user) fetchData();
  }, [user]);

  if (!user) return <p>Loading user info...</p>;

  return (
    <div className="max-w-2xl mx-auto mt-6">
      <div className="mb-6 p-4 border rounded bg-white shadow-sm">
        <h2 className="text-xl font-semibold">{user.full_name}</h2>
        <p className="text-gray-600">{user.email}</p>
        
        <div className="flex space-x-6 mt-4">
          <Link to="/followers" className="text-center">
            <div className="text-xl font-bold">{counts.followers}</div>
            <div className="text-gray-500 text-sm">Followers</div>
          </Link>
          <Link to="/followers" className="text-center">
            <div className="text-xl font-bold">{counts.following}</div>
            <div className="text-gray-500 text-sm">Following</div>
          </Link>
        </div>

        <Link 
          to="/follow-requests" 
          className="block mt-4 text-blue-500 hover:underline"
        >
          View Follow Requests
        </Link>
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