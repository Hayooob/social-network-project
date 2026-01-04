import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../VerifyAuth";
import { logout } from "../api/auth";

export default function Navbar() {
  const { user, setUser } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    try {
      // tell backend to clear session
      await logout();   
    } catch (err) {
      console.error("Logout failed (ignoring):", err);
      // we still clear local state
    } finally {
      setUser(null);    
      navigate("/login");
    }
  }

  return (
    <nav className="navbar">
      <Link to="/" className="navbar-logo">
        Social
      </Link>

      <div className="navbar-links">
        {!user && (
          <>
            <Link to="/login">Login</Link>
            <Link to="/register">Register</Link>
          </>
        )}

        {user && (
          <>
            <Link to="/feed">Feed</Link>
            <Link to="/profile">Profile</Link>
            <button type="button" onClick={handleLogout}>
              Logout
            </button>
          </>
        )}
      </div>
    </nav>
  );
}
