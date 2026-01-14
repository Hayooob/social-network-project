import React from 'react';

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
      <div className="font-bold">{post.author_name || 'Unknown'}</div>
      <div className="text-sm text-gray-500">{formatDate(post.created_at)}</div>
      <p className="mt-2">{post.content}</p>
    </div>
  );
}