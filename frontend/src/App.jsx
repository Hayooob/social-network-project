import { Routes, Route, Navigate } from "react-router-dom";
import Layout from "./components/Layout";
import Login from "./pages/Login";
import Register from "./pages/Register";
import FeedPage from "./pages/FeedPage";
import ProfilePage from "./pages/ProfilePage";
import FollowRequestsPage from "./pages/FollowRequestsPage";
import FollowersPage from "./pages/FollowersPage";
import Privateroute from "./components/Privateroute";

//  Stage 5
import UserProfilePage from "./pages/UserProfilePage";

//  Stage 6
import MessagesPage from "./pages/MessagesPage";
import NotificationsPage from "./pages/NotificationsPage";

// Stage 7
import GroupsPage from "./pages/GroupsPage";
import GroupDetailPage from "./pages/GroupDetailPage";
import CreateGroupPage from "./pages/CreateGroupPage";
//import EventsPage from "./pages/EventsPage";
//import EventDetailPage from "./pages/EventDetailPage";
//import CreateEventPage from "./pages/CreateEventPage";

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

        {/*  Stage 5 route */}
        <Route
          path="/users/:id"
          element={
            <Privateroute>
              <UserProfilePage />
            </Privateroute>
          }
        />

        {/*  Stage 6 routes */}
        <Route
          path="/messages"
          element={
            <Privateroute>
              <MessagesPage />
            </Privateroute>
          }
        />
        <Route
          path="/messages/:userId"
          element={
            <Privateroute>
              <MessagesPage />
            </Privateroute>
          }
        />
        <Route
          path="/notifications"
          element={
            <Privateroute>
              <NotificationsPage />
            </Privateroute>
          }
        />
        {/* Stage 7 routes */}
        <Route
          path="/groups"
          element={
            <Privateroute>
              <GroupsPage />
            </Privateroute>
          }
        />
        <Route
          path="/groups/create"
          element={
            <Privateroute>
              <CreateGroupPage />
            </Privateroute>
          }
        />
        <Route
          path="/groups/:id"
          element={
            <Privateroute>
              <GroupDetailPage />
            </Privateroute>
          }
        />

        {/* <Route
          path="/events"
          element={
            <Privateroute>
              <EventsPage />
            </Privateroute>
          }
        />
        <Route
          path="/events/create"
          element={
            <Privateroute>
              <CreateEventPage />
            </Privateroute>
          }
        />
        <Route
          path="/events/:id"
          element={
            <Privateroute>
              <EventDetailPage />
            </Privateroute>
          }
        /> */}

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