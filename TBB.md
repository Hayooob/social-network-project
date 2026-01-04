Stage 3 – Basic posts, feed, and profile posts

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