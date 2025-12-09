TO BE BUILT:

- project skeleton + FS (done)

FRONTEND
- /api:
    - auth.js file (create a function to call go api with fetch):
        - login(email, password)
        - register(formData)
        - getCurrentUser()

- /pages:
    - Loginpage (call auth.js)
    - Resgisterpage
    - Feedpage
    - Profilepage
    - Groupspage
    - Chatpage

- /components:
    - Navbar (on every page)
    - Post Detail (how a single post looks like)
    - Profile Card (how profile info looks like)
    - Chat Window (messages plus input area)

BACKEND
- /cmd/server -> main.go (open db, create server & listen)

- /internal/db: (for DB connection + queries)
    - sessions.go -> create, delete & getuserby session funcs.
    - db.go -> open db & run migrations.
    - userdbreq.go -> create user and getuserbyemail type querries.
    - /migrations -> (create table sql querries) (create session tables etc)

- /internal/app:
    - server.go -> defines server, routes, mux (handle auth routes to mux (mux.handlefunc(/register,...)))
    - handlers.go -> handle register(post), login(post), logout(post), currentuser(get) 
    - middleware.go -> file to read cookie load user from session...


- /internal/models: (Define all relavent structs)
    - users.go -> define user struct
    - session, post group msgs etc... 