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

**Stage 2 – Real auth, sessions, and protected pages

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