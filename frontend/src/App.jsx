import { Routes, Route, Navigate } from "react-router-dom";
import Layout from "./components/Layout";
import Login from "./pages/Login";
import Register from "./pages/Register";
import FeedPage from "./pages/FeedPage";
import ProfilePage from "./pages/ProfilePage";
import FollowRequestsPage from "./pages/FollowRequestsPage";
import FollowersPage from "./pages/FollowersPage";
import Privateroute from "./components/Privateroute";

// ✅ Stage 5
import UserProfilePage from "./pages/UserProfilePage";

export default function App() {
  return (
    <Layout>
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
        <Route
          path="/follow-requests"
          element={
            <Privateroute>
              <FollowRequestsPage />
            </Privateroute>
          }
        />
        <Route
          path="/followers"
          element={
            <Privateroute>
              <FollowersPage />
            </Privateroute>
          }
        />

        {/* ✅ Stage 5 route */}
        <Route
          path="/users/:id"
          element={
            <Privateroute>
              <UserProfilePage />
            </Privateroute>
          }
        />

        {/* default */}
        <Route path="/" element={<Navigate to="/feed" replace />} />
        <Route
          path="*"
          element={
            <div className="main-content page-center text-center">
              <h1>404</h1>
              <p className="mt-2">Page not found</p>
            </div>
          }
        />
      </Routes>
    </Layout>
  );
}
