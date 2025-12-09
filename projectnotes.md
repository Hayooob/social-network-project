Quick Overview

BACKEND
- All the backend code is Go, it will use a SQLite db.

- internal/app will have everything related to handling requests:
    - login handler.
    - register handler.
    - create post handler.

- internal/db will have everything related to the database:
    - opening the SQLite file, running queries &helper functions.

- FRONTEND
    - (currently TBD) React + vite, using react router for pages. 

        






Main funcs to tackle: (auth, profiles, followers, posts, groups, chat (websockets), notifications.)
