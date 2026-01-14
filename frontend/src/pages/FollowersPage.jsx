import React, { useEffect, useState } from 'react';
import { getMyFollowers, getMyFollowing } from '../api/followers';
import { useAuth } from '../VerifyAuth';
import UserCard from '../components/UserCard';

export default function FollowersPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('followers');
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    setLoading(true);
    try {
      const followersData = await getMyFollowers();
      const followingData = await getMyFollowing();
      setFollowers(followersData);
      setFollowing(followingData);
    } catch (err) {
      console.error('Error loading follow data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="max-w-2xl mx-auto mt-6">
      <div className="flex space-x-4 mb-6">
        <button
          type="button"
          onClick={() => setActiveTab('followers')}
          className={activeTab === 'followers' 
            ? 'px-4 py-2 bg-blue-500 text-white rounded' 
            : 'px-4 py-2 bg-gray-200 rounded'}
        >
          Followers ({followers.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('following')}
          className={activeTab === 'following' 
            ? 'px-4 py-2 bg-blue-500 text-white rounded' 
            : 'px-4 py-2 bg-gray-200 rounded'}
        >
          Following ({following.length})
        </button>
      </div>

      {loading ? (
        <p>Loading...</p>
      ) : (
        <div>
          {activeTab === 'followers' && (
            followers.length === 0 ? (
              <p>No followers yet.</p>
            ) : (
              followers.map((follow) => (
                <UserCard
                  key={follow.follower_id}
                  user={follow}
                  showFollowButton={true}
                  currentUserId={user ? user.id : null}
                />
              ))
            )
          )}
          {activeTab === 'following' && (
            following.length === 0 ? (
              <p>Not following anyone yet.</p>
            ) : (
              following.map((follow) => (
                <UserCard
                  key={follow.following_id}
                  user={follow}
                  showFollowButton={false}
                  currentUserId={user ? user.id : null}
                />
              ))
            )
          )}
        </div>
      )}
    </div>
  );
}