
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