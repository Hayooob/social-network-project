package app

import (
	"encoding/json"
	"net/http"
	"social-network/internal/db"
	"log"
)

//helpers:
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// register:

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

	// validate required fields
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
	dateOfBirth := "" // placeholder to satisfy NOT NULL TEXT

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
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// login:

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
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// logout:

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		_ = LogoutUser(r.Context(), s.DB, cookie.Value)
		// Clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// Even if there was no cookie treat as success
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// me (current logged in user):

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
		"created_at":    user.CreatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST HANDLERS

// CreatePost handles POST /api/posts - creates a new post for the logged-in user
func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("=== CreatePost called ===")
	log.Println("Method:", r.Method)
	
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get current user from context (set by AuthMiddleware)
	user := CurrentUser(r.Context())
	log.Println("Current user:", user)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
		
	// Parse request body
	var req struct {
		Content string `json:"content"`
		Privacy string `json:"privacy"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	
	// Validate content
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}
	
	// Default privacy to public if not specified
	if req.Privacy == "" {
		req.Privacy = "public"
	}
	
	// Validate privacy value
	if req.Privacy != "public" && req.Privacy != "private" && req.Privacy != "almost-private" {
		writeError(w, http.StatusBadRequest, "invalid privacy setting")
		return
	}
	
	// Create the post
	post, err := db.InsertPost(s.DB, int(user.ID), req.Content, req.Privacy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}
	
	writeJSON(w, http.StatusCreated, post)
}

// GetFeed handles GET /api/feed - returns the public feed
func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request) {
	log.Println("=== GetFeed called ===")
	log.Println("Method:", r.Method)
	log.Println("DB is nil?", s.DB == nil)
	
	if r.Method != http.MethodGet {
		log.Println("Method not allowed")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get recent public posts (limit to 50)
	log.Println("Calling GetPublicFeed...")
	posts, err := db.GetPublicFeed(s.DB, 50)
	if err != nil {
		log.Println("GetPublicFeed ERROR:", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch feed")
		return
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

	// Get current user from context
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	
	// Fetch user's posts
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

	// Return safe user data (no password hashes)
	var suggestions []map[string]any
	for _, u := range users {
		suggestions = append(suggestions, map[string]any{
			"id":        u.ID,
			"uuid":      u.UUID,
			"full_name": u.FullName,
			"email":     u.Email,
			"nickname":  u.Nickname,
		})
	}

	// Return empty array instead of null if no suggestions
	if suggestions == nil {
		suggestions = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, suggestions)
}