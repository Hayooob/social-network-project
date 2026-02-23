import React from 'react';
import { Link } from 'react-router-dom';

export default function PostCard({ post }) {
  const formatDate = (timestamp) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch {
      return 'Recently';
    }
  };

  return (
    <div className="border p-4 mb-4 rounded">
      <div className="font-bold">
        <Link to={`/users/${post.user_id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
          {post.author_name || 'Unknown'}
        </Link>
      </div>
      <div className="text-sm text-gray-500">{formatDate(post.created_at)}</div>
      <p className="mt-2">{post.content}</p>
    {post.image_path && (
  <img
    src={
      post.image_path.startsWith("http")
        ? post.image_path
        : `http://localhost:8080${post.image_path}`
    }
    alt="post"
    style={{ marginTop: 12, maxWidth: "100%", borderRadius: 12 }}
  />
)}

    </div>
  );
}