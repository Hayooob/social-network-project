import { useAuth } from "../VerifyAuth";

export default function ProfilePage() {
  const { user } = useAuth();

  if (!user) {
    // this page will be behind PrivateRoute but incase not this is a  check
    return (
      <div className="page">
        <h1>Profile</h1>
        <p>You are not logged in.</p>
      </div>
    );
  }

  return (
    <div className="page">
      <h1>Your Profile</h1>

      <div className="profile-card">
        <p>
          <strong>Username:</strong> {user.username}
        </p>
        <p>
          <strong>Email:</strong> {user.email}
        </p>
        {user.createdAt && (
          <p>
            <strong>Joined:</strong> {new Date(user.createdAt).toLocaleString()}
          </p>
        )}
      </div>

      <p style={{ marginTop: "1rem" }}>
        Here we add (change username, bio, avatar,
        etc.) 
      </p>
    </div>
  );
}
