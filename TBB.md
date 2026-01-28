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

Stage 6 – Messaging and Notifications

Backend:

- internal/db/migrations/005createmessages.sql
Create the messages table with id, sender_id, receiver_id, content, is_read, and created_at. Add foreign keys to users table with ON DELETE CASCADE. Add index on (sender_id, receiver_id) for fast conversation lookups.

- internal/db/migrations/006createnotifications.sql
Create the notifications table with id, user_id, type (follow_request, follow_accept, new_message, etc.), reference_id (id of related item), is_read, and created_at. Add foreign key to users table with ON DELETE CASCADE.

- internal/models/messages.go
Define the Message struct that matches the messages table. Include optional SenderName and ReceiverName fields for when you join with users table.

- internal/models/notifications.go
Define the Notification struct that matches the notifications table. Include optional fields for the related user name or content preview.

- internal/db/messagedb.go
Write helper functions: SendMessage to insert a new message, GetConversation to get messages between two users ordered by time, GetConversationList to get list of users you have conversations with (with latest message preview), MarkMessageAsRead to update is_read to true, GetUnreadCount to count unread messages for a user.

- internal/db/notificationdb.go
Write helper functions: CreateNotification to insert a new notification, GetNotifications to get all notifications for a user ordered by newest, MarkNotificationRead to update is_read to true, MarkAllNotificationsRead to mark all as read, GetUnreadNotificationCount to count unread notifications.

- internal/app/messagehandlers.go
Add handlers for POST /api/messages/{userId} to send a message to a user, GET /api/messages/{userId} to get conversation with a user, GET /api/messages to get list of all conversations, POST /api/messages/{messageId}/read to mark a message as read.

- internal/app/notificationhandlers.go
Add handlers for GET /api/notifications to get all notifications, POST /api/notifications/{id}/read to mark one as read, POST /api/notifications/read-all to mark all as read, GET /api/notifications/unread-count to get count of unread.

- internal/app/server.go
Register all the new message and notification routes on the mux.

- Update existing handlers
When a follow request is sent, create a notification for the target user. When a follow request is accepted, create a notification for the requester. When a message is sent, create a notification for the receiver.

Frontend:

- src/api/messages.js
Wrap the message endpoints with helpers like sendMessage, getConversation, getConversationList, and markAsRead using the existing ftchclient.

- src/api/notifications.js
Wrap the notification endpoints with helpers like getNotifications, markNotificationRead, markAllRead, and getUnreadCount using the existing ftchclient.

- src/components/MessageBubble.jsx
A single message bubble showing the content and timestamp. Style differently for sent vs received messages (sent on right, received on left).

- src/components/ConversationCard.jsx
A card showing a conversation preview with the other user's name, last message snippet, timestamp, and unread indicator.

- src/components/NotificationCard.jsx
A card showing a notification with icon based on type, message text, timestamp, and unread styling. Clicking it navigates to the relevant page (profile for follows, messages for new message).

- src/components/NotificationBell.jsx
A bell icon for the navbar that shows unread count badge. Clicking it opens a dropdown or navigates to notifications page.

- src/pages/MessagesPage.jsx
A page with two panels: left panel shows ConversationCard list, right panel shows the active conversation with MessageBubble components and a text input to send new messages.

- src/pages/NotificationsPage.jsx
A page that calls getNotifications on mount and displays a list of NotificationCard components with a "Mark all as read" button at top.

- src/components/Layout.jsx
Update the navbar to include NotificationBell component showing unread count.

- src/App.jsx
Add routes for /messages, /messages/:userId, and /notifications wrapped in Privateroute.

---

Stage 7 – Groups and Events

Backend:

- internal/db/migrations/007creategroups.sql
Create the groups table with id, creator_id, name, description, is_private, and created_at. Add foreign key to users table. Create group_members table with id, group_id, user_id, role (admin, member), status (pending, accepted), and joined_at. Add unique constraint on (group_id, user_id).

- internal/db/migrations/008creategrouposts.sql
Create the group_posts table with id, group_id, user_id, content, and created_at. Add foreign keys to groups and users tables with ON DELETE CASCADE.

- internal/db/migrations/009createevents.sql
Create the events table with id, group_id (nullable for non-group events), creator_id, title, description, location, event_date, and created_at. Add foreign keys to groups and users tables. Create event_responses table with id, event_id, user_id, response (going, not_going, maybe), and created_at. Add unique constraint on (event_id, user_id).

- internal/models/groups.go
Define the Group struct and GroupMember struct that match the tables. Include optional fields like CreatorName, MemberCount for joined data.

- internal/models/groupposts.go
Define the GroupPost struct that matches the group_posts table. Include optional AuthorName field.

- internal/models/events.go
Define the Event struct and EventResponse struct that match the tables. Include optional fields like CreatorName, GroupName, response counts.

- internal/db/groupdb.go
Write helper functions: CreateGroup to insert a new group, GetGroupByID to get group details, GetUserGroups to list groups a user belongs to, AddGroupMember to add a member (pending if private group), UpdateMemberStatus to accept a pending member, RemoveGroupMember to leave or kick from group, GetGroupMembers to list all members, IsGroupMember to check membership.

- internal/db/grouppostdb.go
Write helper functions: CreateGroupPost to insert a post in a group, GetGroupPosts to list posts in a group ordered by newest, DeleteGroupPost to remove a post.

- internal/db/eventdb.go
Write helper functions: CreateEvent to insert a new event, GetEventByID to get event details with response counts, GetUpcomingEvents to list future events for a user (from their groups or public), RespondToEvent to set going/not_going/maybe, GetEventResponses to list who responded and how, GetUserEvents to list events user has responded to.

- internal/app/grouphandlers.go
Add handlers for POST /api/groups to create a group, GET /api/groups to list user's groups, GET /api/groups/{id} to get group details, POST /api/groups/{id}/join to request to join, POST /api/groups/{id}/leave to leave group, GET /api/groups/{id}/members to list members, POST /api/groups/{id}/members/{userId}/accept to accept a pending member (admin only), POST /api/groups/{id}/members/{userId}/remove to kick a member (admin only), GET /api/groups/{id}/posts to get group posts, POST /api/groups/{id}/posts to create a group post.

- internal/app/eventhandlers.go
Add handlers for POST /api/events to create an event, GET /api/events to list upcoming events, GET /api/events/{id} to get event details, POST /api/events/{id}/respond to set your response, GET /api/events/{id}/responses to see who's going.

- internal/app/server.go
Register all the new group and event routes on the mux.

Frontend:

- src/api/groups.js
Wrap the group endpoints with helpers like createGroup, getMyGroups, getGroupDetails, joinGroup, leaveGroup, getGroupMembers, acceptMember, removeMember, getGroupPosts, createGroupPost using the existing ftchclient.

- src/api/events.js
Wrap the event endpoints with helpers like createEvent, getUpcomingEvents, getEventDetails, respondToEvent, getEventResponses using the existing ftchclient.

- src/components/GroupCard.jsx
A card showing group name, member count, and join/leave button based on membership status.

- src/components/GroupMemberCard.jsx
A card showing a group member with their name, role badge (admin/member), and remove button if you're an admin.

- src/components/EventCard.jsx
A card showing event title, date, location, response counts (X going, Y maybe), and your current response buttons.

- src/components/EventResponseButtons.jsx
Three buttons (Going, Maybe, Not Going) that highlight based on current response and call respondToEvent on click.

- src/pages/GroupsPage.jsx
A page listing all your groups with GroupCard components and a "Create Group" button that opens a modal or navigates to create page.

- src/pages/GroupDetailPage.jsx
A page showing group info, member list, and group posts feed. If admin, show pending member requests. Include form to create new group post.

- src/pages/CreateGroupPage.jsx
A form to create a new group with name, description, and public/private toggle.

- src/pages/EventsPage.jsx
A page listing upcoming events with EventCard components and a "Create Event" button.

- src/pages/EventDetailPage.jsx
A page showing full event details, response buttons, and list of who's going/maybe/not going.

- src/pages/CreateEventPage.jsx
A form to create a new event with title, description, location, date picker, and optional group selector.

- src/components/Layout.jsx
Update the navbar to include links to Groups and Events pages.

- src/App.jsx
Add routes for /groups, /groups/:id, /groups/create, /events, /events/:id, /events/create wrapped in Privateroute.