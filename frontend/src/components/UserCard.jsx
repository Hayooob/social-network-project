import React from "react";
import { Link } from "react-router-dom";
import FollowButton from "./FollowButton";

export default function UserCard({ user, showFollowButton, currentUserId }) {
  const isOwnProfile = currentUserId && currentUserId === user.follower_id;
  const displayName = user.follower_name || user.following_name || "Unknown";
  const targetUserId = user.follower_id || user.following_id;

  return (
    <div className="border p-4 mb-4 rounded flex justify-between items-center">
      <Link to={`/users/${targetUserId}`} className="font-bold">
        {displayName}
      </Link>

      {showFollowButton && !isOwnProfile && (
        <FollowButton userId={targetUserId} initialStatus="none" />
      )}
    </div>
  );
}

