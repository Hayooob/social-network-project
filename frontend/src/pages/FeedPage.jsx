import { useAuth } from "../VerifyAuth";

export default function FeedPage() {
  const { user } = useAuth();

  return (
    <div className="page">
      <h1>Feed</h1>
      <p>Welcome{user ? `, ${user.username}` : ""} 👋</p>

      <p>
        Posts should be displayed here from db after auth
      </p>
    </div>
  );
}
