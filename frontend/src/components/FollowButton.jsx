import React, { useState } from 'react';
import { followUser, unfollowUser } from '../api/followers';

export default function FollowButton({ userId, initialStatus, onStatusChange }) {
  const [status, setStatus] = useState(initialStatus || 'none');
  const [loading, setLoading] = useState(false);

  const handleClick = async () => {
    setLoading(true);
    try {
      if (status === 'none') {
        const response = await followUser(userId);
        setStatus(response.status);
        if (onStatusChange) onStatusChange(response.status);
      } else if (status === 'accepted' || status === 'pending') {
        await unfollowUser(userId);
        setStatus('none');
        if (onStatusChange) onStatusChange('none');
      }
    } catch (err) {
      console.error('Follow action failed:', err);
      alert('Failed to update follow status.');
    } finally {
      setLoading(false);
    }
  };

  const getButtonText = () => {
    if (loading) return 'Loading...';
    if (status === 'none') return 'Follow';
    if (status === 'pending') return 'Requested';
    if (status === 'accepted') return 'Following';
    return 'Follow';
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={loading}
      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
    >
      {getButtonText()}
    </button>
  );
}