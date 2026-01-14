Stage 4 – Followers system (follow, unfollow, requests)

Backend:

- internal/db/migrations/004createfollowers.sql
Create the followers table with id, follower_id, following_id, status (pending or accepted), and created_at. Add unique constraint on (follower_id, following_id) so no duplicate follows. Add foreign keys to users table with ON DELETE CASCADE.

- internal/models/followers.go
Define the Follow struct that matches the followers table. Include optional fields like FollowerName and FollowingName for when you join with users table.

- internal/db/followerdb.go
Write helper functions: CreateFollow to insert a new follow, GetFollowStatus to check if a follow exists between two users, UpdateFollowStatus to change pending to accepted, DeleteFollow to remove a follow, GetFollowers to list who follows a user, GetFollowing to list who a user follows, and GetPendingRequests to list follow requests awaiting approval.

- internal/app/handlers.go
Add handlers for POST /api/users/{id}/follow to follow a user (instant if public profile, pending if private), POST /api/users/{id}/unfollow to remove a follow, GET /api/me/followers to get your followers, GET /api/me/following to get who you follow, GET /api/me/follow-requests to get pending requests, POST /api/follow-requests/{id}/accept to accept a request, and POST /api/follow-requests/{id}/decline to decline a request.

- internal/app/server.go
Register the new follower routes on the mux. Since Go's default mux doesn't support URL params, you'll need to either parse the URL path manually in a handler or use a router like gorilla/mux.

Frontend:

- src/api/followers.js
Wrap the follower endpoints with helpers like followUser, unfollowUser, getMyFollowers, getMyFollowing, getPendingRequests, acceptRequest, and declineRequest, all using the existing ftchclient.

- src/components/FollowButton.jsx
A button that shows "Follow" when not following, "Requested" when pending, or "Following" when accepted. Clicking it calls the appropriate API function and updates its state.

- src/components/FollowRequestCard.jsx
A card showing a pending follow request with the requester's name and Accept/Decline buttons that call the API and remove the card on success.

- src/components/UserCard.jsx
A small card showing a user's name and a FollowButton, used in follower/following lists.

- src/pages/ProfilePage.jsx
Update to show follower and following counts, and if viewing your own profile show a link to pending follow requests. Add a FollowButton if viewing someone else's profile.

- src/pages/FollowRequestsPage.jsx
A new page that calls getPendingRequests on mount and displays a list of FollowRequestCard components. Should be wrapped in Privateroute.

- src/App.jsx
Add routes for /follow-requests and later /users/:id for viewing other profiles.