***Stage 1 (DONE)

Backend:

- cmd/main.go 
Opens the SQLite database, runs the migrations, builds a Server, and starts the HTTP server.

- internal/app/server.go 
Holds the shared DB and Mux on the Server struct, wires up routes like /health, and exposes a Listen method to run the app.

- internal/db/db.go 
Provides OpenDb and RunMigrations so the app can connect to the database and apply all sql migration files on startup.

- internal/db/migrations/001createusers.sql 
Creates the users table with basic user info and timestamps.

- internal/db/migrations/002createsessions.sql 
Creates the sessions table to store login sessions tied to a user.

- internal/db/migrations/003createposts.sql 
Creates the posts table so we have somewhere to store feed posts later.

- internal/models/users.go 
Defines the User struct that mirrors the users table.

- internal/db/userdbreq.go 
Adds simple helpers like GetUserByID to fetch users from the database.

Frontend: 

- src/main.jsx 
Boots the React app, wraps it in BrowserRouter and AuthProvider, and renders into the root element.

- src/App.jsx 
Sets up the main routes for /login, /register, /feed, and /profile, and uses Privateroute for the protected ones.

- src/components/NavBar.jsx 
Renders the top navigation bar with links to the main pages and later handles logout.

- src/components/Privateroute.jsx 
A wrapper that will either show the requested page or send you back to /login depending on auth state.

- src/context/VerifyAuth.jsx 
Sets up a React context where we keep the current user and a useAuth hook to read or update it.

- src/pages/Login.jsx 
Shows the login form UI for email and password and was initially wired only to local state.

- src/pages/Register.jsx 
Shows the registration form UI for username, email, password, and confirm password.

- src/pages/FeedPage.jsx 
Simple placeholder page used to confirm that routing works after login.

- src/pages/ProfilePage.jsx 
Placeholder profile page that will later display the logged in users information.

**Stage 2 – Real auth, sessions, and protected pages (DONE)

Backend:

- internal/app/password.go 
Handles hashing passwords and checking them using bcrypt so plain text passwords are never stored.

- internal/app/auth.go 
Contains the core auth logic for registering a user, logging in and creating a session, logging out, and looking up a user from a session token.

- internal/db/sessions.go 
Inserts, fetches, and deletes rows in the sessions table so we can manage active logins.

- internal/app/middleware.go 
AuthMiddleware reads the session_id cookie and loads the current user into the request context, and CORSMiddleware adds the headers needed so the React dev server on localhost:5173 can talk to the API on localhost:8080 with cookies.

- internal/app/handlers.go 
Implements the HTTP handlers for POST /api/register, POST /api/login, POST /api/logout, and GET /api/me including validation and JSON responses.

- internal/app/server.go 
Registers the auth routes on the mux and in Listen wraps the mux with AuthMiddleware and CORSMiddleware before passing it to http.ListenAndServe.

Frontend:

- src/api/ftchclient.js 
A small fetch wrapper that targets VITE_API_URL or http://localhost:8080
, always sends JSON, includes credentials, and throws a helpful error when something fails.

- src/api/auth.js 
Uses the fetch client to talk to the backend with register, login, logout, and getCurrentUser so components do not have to care about raw URLs.

- src/context/VerifyAuth.jsx 
On mount calls getCurrentUser to hydrate the user from /api/me and provides user and setUser through context and the useAuth hook.

- src/components/Privateroute.jsx 
Checks user from useAuth and either renders the child route or redirects to /login when there is no session.

- src/components/NavBar.jsx 
Shows links to /feed and /profile and when logged in shows a Logout button that calls logout, clears the user in context, and sends you back to /login.

- src/pages/Register.jsx 
Submits the registration form to POST /api/register via auth.register, shows any error message, and then sends you to /login when it succeeds.

- src/pages/Login.jsx 
Sends credentials to POST /api/login via auth.login, stores the returned user in context, and redirects you to /feed.

- src/pages/FeedPage.jsx 
Now actually protected by Privateroute, so it only renders when you are logged in and can rely on useAuth to know who you are.

- src/pages/ProfilePage.jsx 
Also protected by Privateroute and shows the current users basic profile info from useAuth.

Stage 3 – Basic posts, feed, and profile posts (DONE)

Backend:

- internal/models/posts.go 
Define the Post struct that matches the posts table (also make the posts table in migrations folder if not done)

- internal/db/postdb.go 
write helper functions to insert a new post, list recent posts for the global feed, and list posts by a specific user id ordered by newest first

- internal/app/handlers.go 
Add handlers for POST /api/posts to create a new post for the logged in user, GET /api/feed to return the public feed, and GET /api/me/posts to return posts by the current user

- internal/app/server.go 
Register the new post routes on the mux so the frontend can call /api/posts, /api/feed, and /api/me/posts next to the existing routes

Frontend:

- src/api/posts.js 
Wrap the post endpoints with small helpers like getFeed, createPost, and getMyPosts, all using the existing ftchclient 

- src/components/PostForm.jsx 
a form with a textarea and submit button that lets a logged in user create a new post and then refreshes the feed.

- src/components/PostCard.jsx 
create a single post card with the author name, timestamp, and content 

- src/pages/FeedPage.jsx 
change it so it now calls getFeed on mount instead of the filler txt. (shows a list of PostCard components & include PostForm at the top so users can add new posts directly from the feedpage)

- src/pages/ProfilePage.jsx 
it should useAuth to show the current users basic info & calls getMyPosts

---

Stage 4 – Followers system (DONE)

Backend:

- internal/db/migrations/004createfollowers.sql
Create the followers table with id, follower_id, following_id, status (pending or accepted), and created_at. Add unique constraint on (follower_id, following_id) so no duplicate follows. Add foreign keys to users table with ON DELETE CASCADE.

- internal/models/follows.go
Define the Follow struct that matches the followers table. Include optional fields like FollowerName and FollowingName for when you join with users table.

- internal/db/followdb.go
Write helper functions: CreateFollow to insert a new follow, GetFollowStatus to check if a follow exists between two users, UpdateFollowStatus to change pending to accepted, DeleteFollow to remove a follow, GetFollowers to list who follows a user, GetFollowing to list who a user follows, and GetPendingFollowRequests to list follow requests awaiting approval.

- internal/app/followhandlers.go
Add handlers for POST /api/users/{id}/follow to follow a user (instant if public profile, pending if private), POST /api/users/{id}/unfollow to remove a follow, GET /api/me/followers to get your followers, GET /api/me/following to get who you follow, GET /api/me/follow-requests to get pending requests, POST /api/follow-requests/{id}/accept to accept a request, and POST /api/follow-requests/{id}/decline to decline a request.

- internal/app/server.go
Register the new follower routes on the mux.

Frontend:

- src/api/followers.js
Wrap the follower endpoints with helpers like followUser, unfollowUser, getMyFollowers, getMyFollowing, getPendingRequests, acceptFollowRequest, and declineFollowRequest, all using the existing ftchclient.

- src/components/FollowButton.jsx
A button that shows "Follow" when not following, "Requested" when pending, or "Following" when accepted. Clicking it calls the appropriate API function and updates its state.

- src/components/FollowRequestCard.jsx
A card showing a pending follow request with the requester's name and Accept/Decline buttons that call the API and remove the card on success.

- src/components/UserCard.jsx
A small card showing a user's name and a FollowButton, used in follower/following lists.

- src/pages/ProfilePage.jsx
Update to show follower and following counts, and if viewing your own profile show a link to pending follow requests.

- src/pages/FollowRequestsPage.jsx
A new page that calls getPendingRequests on mount and displays a list of FollowRequestCard components. Should be wrapped in Privateroute.

- src/pages/FollowersPage.jsx
A new page with tabs for Followers and Following, displays UserCard components for each list.

- src/App.jsx
Add routes for /follow-requests and /followers wrapped in Privateroute.

---

Stage 5 – Profile privacy and viewing other users (DONE)

Backend:

- internal/app/handlers.go
Add handler for GET /api/users/{id} to get another user's public profile info. Should check if profile is private and if requester is a follower before returning full info.

- internal/app/handlers.go
Add handler for POST /api/me/privacy to toggle the current user's is_private setting between true and false.

- internal/db/userdbreq.go
Add UpdateUserPrivacy function to update the is_private field for a user.

- internal/db/followdb.go
Add IsFollowing function that returns true/false to quickly check if one user follows another.

- internal/app/server.go
Register the new routes: GET /api/users/{id} and POST /api/me/privacy.

Frontend:

- src/api/users.js
Create new file with getUserProfile(userId) to fetch another user's profile, and togglePrivacy() to toggle your own privacy setting.

- src/components/PrivacyToggle.jsx
A toggle or button that shows current privacy status (Public/Private) and calls togglePrivacy on click to switch it.

- src/pages/ProfilePage.jsx
Add the PrivacyToggle component so users can switch their profile between public and private.

- src/pages/UserProfilePage.jsx
A new page for viewing other users' profiles. Shows user info, follower/following counts, their posts (if public or you follow them), and a FollowButton.

- src/App.jsx
Add route for /users/:id that renders UserProfilePage wrapped in Privateroute.

- src/components/UserCard.jsx
Update to make the username clickable, linking to /users/{id} so you can view that user's profile.

---