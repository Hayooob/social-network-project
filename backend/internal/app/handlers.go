package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/db"
	"social-network/internal/models"
)

// helpers:
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// -------------------------
// AUTH: register / login / logout / me
// -------------------------

type registerRequest struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.ConfirmPassword == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	if req.Password != req.ConfirmPassword {
		writeError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	// map frontend fields to DB fields
	fullName := req.Username
	dateOfBirth := "" // placeholder to satisfy NOT NULL TEXT (your earlier stages)

	user, err := RegisterUser(r.Context(), s.DB, fullName, dateOfBirth, req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrEmailAlreadyInUse:
			writeError(w, http.StatusBadRequest, "email already in use")
			return
		default:
			writeError(w, http.StatusInternalServerError, "could not register user")
			return
		}
	}

	resp := map[string]any{
		"id":            user.ID,
		"uuid":          user.UUID,
		"email":         user.Email,
		"full_name":     user.FullName,
		"date_of_birth": user.DateOfBirth,
		"is_private":    user.IsPrivate,
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "missing email or password")
		return
	}

	user, token, expiresAt, err := LoginUser(r.Context(), s.DB, req.Email, req.Password, 0)
	if err != nil {
		if err == ErrInvalidCredentials {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		// Secure: true, // enable when using HTTPS
	})

	resp := map[string]any{
		"id":            user.ID,
		"uuid":          user.UUID,
		"email":         user.Email,
		"full_name":     user.FullName,
		"date_of_birth": user.DateOfBirth,
		"is_private":    user.IsPrivate,
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		_ = LogoutUser(r.Context(), s.DB, cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp := map[string]any{
		"id":            user.ID,
		"uuid":          user.UUID,
		"email":         user.Email,
		"full_name":     user.FullName,
		"date_of_birth": user.DateOfBirth,
		"is_private":    user.IsPrivate,
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// -------------------------
// STAGE 5: toggle privacy
// POST /api/me/privacy
// -------------------------
func (s *Server) handleTogglePrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	newVal := !user.IsPrivate
	if err := db.UpdateUserPrivacy(s.DB, user.ID, newVal); err != nil {
		log.Println("UpdateUserPrivacy error:", err)
		writeError(w, http.StatusInternalServerError, "failed to update privacy")
		return
	}

	// return new value (frontend will update state)
	writeJSON(w, http.StatusOK, map[string]any{
		"is_private": newVal,
	})
}

// -------------------------
// STAGE 5: view other user profile
// GET /api/users/{id}
// -------------------------
func (s *Server) handleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUser := CurrentUser(r.Context())
	if currentUser == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// ID comes from path: /api/users/{id}
	// routing is done in followhandlers.go
	idStr := r.URL.Query().Get("_id") // not used normally
	_ = idStr                         // keep silent; actual parse done in followhandlers.go

	// Extract id from URL path here (safer in case followhandlers passes directly too)
	trimmed := r.URL.Path[len("/api/users/"):]
	// if anything after id exists, it's not this handler's job
	if trimmed == "" || containsSlash(trimmed) {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	targetID64, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	targetUser, err := db.GetUserByID(s.DB, targetID64)
	if err != nil {
		log.Println("GetUserByID error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if targetUser == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// If viewing yourself: full info + all your posts
	if targetUser.ID == currentUser.ID {
		posts, err := db.GetPostsByUserID(s.DB, int(currentUser.ID))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch posts")
			return
		}

		followerCount, _ := db.GetFollowerCount(s.DB, int(currentUser.ID))
		followingCount, _ := db.GetFollowingCount(s.DB, int(currentUser.ID))

		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":            targetUser.ID,
				"uuid":          targetUser.UUID,
				"email":         targetUser.Email,
				"full_name":     targetUser.FullName,
				"date_of_birth": targetUser.DateOfBirth,
				"avatar_url":    targetUser.AvatarURL,
				"nickname":      targetUser.Nickname,
				"about_me":      targetUser.AboutMe,
				"is_private":    targetUser.IsPrivate,
				"created_at":    targetUser.CreatedAt,
			},
			"counts": map[string]any{
				"followers": followerCount,
				"following": followingCount,
			},
			"posts": posts,
			"viewer": map[string]any{
				"is_following": false,
				"is_self":      true,
				"can_view":     true,
			},
		})
		return
	}

	// Determine if current user follows target (accepted)
	isFollowing, err := db.IsFollowing(s.DB, int(currentUser.ID), int(targetUser.ID))
	if err != nil {
		log.Println("IsFollowing error:", err)
		writeError(w, http.StatusInternalServerError, "failed to check follow status")
		return
	}

	canViewFull := !targetUser.IsPrivate || isFollowing

	followerCount, _ := db.GetFollowerCount(s.DB, int(targetUser.ID))
	followingCount, _ := db.GetFollowingCount(s.DB, int(targetUser.ID))

	// Private profile and not a follower => limited response
	if !canViewFull {
		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":         targetUser.ID,
				"uuid":       targetUser.UUID,
				"full_name":  targetUser.FullName,
				"avatar_url": targetUser.AvatarURL,
				"nickname":   targetUser.Nickname,
				"is_private": targetUser.IsPrivate,
			},
			"counts": map[string]any{
				"followers": followerCount,
				"following": followingCount,
			},
			"posts": []any{}, // hidden
			"viewer": map[string]any{
				"is_following": false,
				"is_self":      false,
				"can_view":     false,
			},
		})
		return
	}

	// Public profile OR follower => show profile + visible posts (public + almost-private)
	posts, err := db.GetPostsByUserVisible(s.DB, int(targetUser.ID), true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":            targetUser.ID,
			"uuid":          targetUser.UUID,
			"email":         targetUser.Email,
			"full_name":     targetUser.FullName,
			"date_of_birth": targetUser.DateOfBirth,
			"avatar_url":    targetUser.AvatarURL,
			"nickname":      targetUser.Nickname,
			"about_me":      targetUser.AboutMe,
			"is_private":    targetUser.IsPrivate,
			"created_at":    targetUser.CreatedAt,
		},
		"counts": map[string]any{
			"followers": followerCount,
			"following": followingCount,
		},
		"posts": posts,
		"viewer": map[string]any{
			"is_following": isFollowing,
			"is_self":      false,
			"can_view":     true,
		},
	})
}

// tiny helper
func containsSlash(s string) bool {
	for _, ch := range s {
		if ch == '/' {
			return true
		}
	}
	return false
}

// -------------------------
// POSTS (existing stage 3)
// -------------------------

// CreatePost handles POST /api/posts - creates a new post for the logged-in user
func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("=== CreatePost called ===")
	log.Println("Method:", r.Method)

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	log.Println("Current user:", user)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Content string `json:"content"`
		Privacy string `json:"privacy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}

	if req.Privacy == "" {
		req.Privacy = "public"
	}

	if req.Privacy != "public" && req.Privacy != "private" && req.Privacy != "almost-private" {
		writeError(w, http.StatusBadRequest, "invalid privacy setting")
		return
	}

	post, err := db.InsertPost(s.DB, int(user.ID), req.Content, req.Privacy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

// GetFeed handles GET /api/feed - returns the personalized feed for logged-in users
func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request) {
	log.Println("=== GetFeed called ===")
	log.Println("Method:", r.Method)

	if r.Method != http.MethodGet {
		log.Println("Method not allowed")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())

	var posts []models.Post
	var err error

	if user != nil {
		// Logged in: show personalized feed
		log.Println("Calling GetPersonalizedFeed for user:", user.ID)
		posts, err = db.GetPersonalizedFeed(s.DB, int(user.ID), 50)
	} else {
		// Not logged in: show only public posts from public users
		log.Println("Calling GetPublicFeed (no user)")
		posts, err = db.GetPublicFeed(s.DB, 50)
	}

	if err != nil {
		log.Println("GetFeed ERROR:", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch feed")
		return
	}

	if posts == nil {
		posts = []models.Post{}
	}

	log.Println("Found", len(posts), "posts")
	writeJSON(w, http.StatusOK, posts)
}

// GetMyPosts handles GET /api/me/posts - returns posts by the current user
func (s *Server) GetMyPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	posts, err := db.GetPostsByUserID(s.DB, int(user.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

// GetSuggestedUsers handles GET /api/users/suggestions
func (s *Server) GetSuggestedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := db.GetSuggestedUsers(s.DB, user.ID, 5)
	if err != nil {
		log.Println("GetSuggestedUsers error:", err)
		writeError(w, http.StatusInternalServerError, "failed to get suggestions")
		return
	}

	var suggestions []map[string]any
	for _, u := range users {
		suggestions = append(suggestions, map[string]any{
			"id":         u.ID,
			"uuid":       u.UUID,
			"full_name":  u.FullName,
			"email":      u.Email,
			"nickname":   u.Nickname,
			"is_private": u.IsPrivate,
		})
	}

	if suggestions == nil {
		suggestions = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, suggestions)
}

// SearchUsers handles GET /api/users/search?q=searchterm
func (s *Server) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	// Minimum 1 character to search
	if len(query) < 1 {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	users, err := db.SearchUsers(s.DB, query, int(user.ID), 20)
	if err != nil {
		log.Println("SearchUsers error:", err)
		writeError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	if users == nil {
		users = []models.User{}
	}

	// Return safe user data (no email for privacy)
	var results []map[string]any
	for _, u := range users {
		results = append(results, map[string]any{
			"id":         u.ID,
			"uuid":       u.UUID,
			"full_name":  u.FullName,
			"nickname":   u.Nickname,
			"avatar_url": u.AvatarURL,
			"is_private": u.IsPrivate,
		})
	}

	if results == nil {
		results = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, results)
}