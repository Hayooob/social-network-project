import { Link } from "react-router-dom";

export default function Navbar() {
  return (
    <div style={{ padding: 12, borderBottom: "1px solid #ddd" }}>
      <Link to="/login" style={{ marginRight: 12 }}>Login</Link>
      <Link to="/register">Register</Link>
    </div>
  );
}
