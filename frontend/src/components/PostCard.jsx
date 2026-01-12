import React from 'react';

export default function PostCard({ post }) {
  const { authorName, content, createdAt } = post;

  const formattedDate = new Date(createdAt).toLocaleString();

  return (
    <div className="border rounded p-4 mb-4 bg-white shadow-sm">
      <div className="flex justify-between items-center mb-2">
        <span className="font-semibold">{authorName}</span>
        <span className="text-sm text-gray-500">{formattedDate}</span>
      </div>
      <div className="text-gray-800">{content}</div>
    </div>
  );
}
