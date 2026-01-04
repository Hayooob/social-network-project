import { Routes, Route, Navigate } from "react-router-dom";
import NavBar from "./components/NavBar";
import Login from "./pages/Login";
import Register from "./pages/Register";
import FeedPage from "./pages/FeedPage";
import ProfilePage from "./pages/ProfilePage";
import Privateroute from "./components/Privateroute";

export default function App() {
  return (
    <>
      <NavBar />
      <Routes>
        {/* Public pages */}
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        {/* pages only shown on login */}
        <Route
          path="/feed"
          element={
            <Privateroute>
              <FeedPage />
            </Privateroute>
          }
        />
        <Route
          path="/profile"
          element={
            <Privateroute>
              <ProfilePage />
            </Privateroute>
          }
        />

        {/* default */}
        <Route path="/" element={<Navigate to="/feed" replace />} />
        <Route path="*" element={<div style={{ padding: 16 }}>Not found</div>} />
      </Routes>
    </>
  );
}
