package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	// Legacy field (older frontend)
	Username string `json:"username"`

	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`

	// Audit-required fields
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Nickname    string `json:"nickname"`
	AboutMe     string `json:"about_me"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var (
		req         registerRequest
		avatarURL   *string
		nickname    *string
		aboutMe     *string
		fullName    string
		dateOfBirth string
	)

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
			writeError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}

		req.Email = strings.TrimSpace(r.FormValue("email"))
		req.Password = r.FormValue("password")
		req.ConfirmPassword = r.FormValue("confirmPassword")
		req.FirstName = strings.TrimSpace(r.FormValue("first_name"))
		req.LastName = strings.TrimSpace(r.FormValue("last_name"))
		req.DateOfBirth = strings.TrimSpace(r.FormValue("date_of_birth"))
		req.Nickname = strings.TrimSpace(r.FormValue("nickname"))
		req.AboutMe = strings.TrimSpace(r.FormValue("about_me"))
		req.Username = strings.TrimSpace(r.FormValue("username")) // legacy

		// avatar file (optional) - field name: "avatar"
		file, header, err := r.FormFile("avatar")
		if err == nil && file != nil && header != nil {
			defer file.Close()

			_ = os.MkdirAll("uploads", 0755)
			ext := strings.ToLower(filepath.Ext(header.Filename))
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif":
				// allowed
			default:
				writeError(w, http.StatusBadRequest, "avatar must be JPG, PNG, or GIF")
				return
			}

			filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
			dstPath := filepath.Join("uploads", filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save avatar")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save avatar")
				return
			}

			p := "/uploads/" + filename
			avatarURL = &p
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
	}

	if req.Email == "" || req.Password == "" || req.ConfirmPassword == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	if req.Password != req.ConfirmPassword {
		writeError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	// Full name + DOB: required in the audit form, but keep backward compatibility.
	if req.FirstName != "" || req.LastName != "" {
		fullName = strings.TrimSpace(strings.TrimSpace(req.FirstName) + " " + strings.TrimSpace(req.LastName))
	} else {
		fullName = strings.TrimSpace(req.Username)
	}
	if fullName == "" {
		writeError(w, http.StatusBadRequest, "first name and last name are required")
		return
	}
	dateOfBirth = strings.TrimSpace(req.DateOfBirth)
	if dateOfBirth == "" {
		writeError(w, http.StatusBadRequest, "date of birth is required")
		return
	}

	if strings.TrimSpace(req.Nickname) != "" {
		v := strings.TrimSpace(req.Nickname)
		nickname = &v
	}
	if strings.TrimSpace(req.AboutMe) != "" {
		v := strings.TrimSpace(req.AboutMe)
		aboutMe = &v
	}

	user, err := RegisterUser(r.Context(), s.DB, fullName, dateOfBirth, req.Email, req.Password, avatarURL, nickname, aboutMe)
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
		"avatar_url":    user.AvatarURL,
		"nickname":      user.Nickname,
		"about_me":      user.AboutMe,
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
		"avatar_url":    user.AvatarURL,
		"nickname":      user.Nickname,
		"about_me":      user.AboutMe,
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
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		resp := map[string]any{
			"id":            user.ID,
			"uuid":          user.UUID,
			"email":         user.Email,
			"full_name":     user.FullName,
			"date_of_birth": user.DateOfBirth,
			"avatar_url":    user.AvatarURL,
			"nickname":      user.Nickname,
			"about_me":      user.AboutMe,
			"is_private":    user.IsPrivate,
			"created_at":    user.CreatedAt,
		}
		writeJSON(w, http.StatusOK, resp)
		return

	case http.MethodPut:
		// update profile information
		var (
			fullName    string
			dateOfBirth string
			nickname    *string
			aboutMe     *string
			avatarURL   *string
		)

		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				writeError(w, http.StatusBadRequest, "invalid multipart form")
				return
			}

			fullName = strings.TrimSpace(r.FormValue("full_name"))
			dateOfBirth = strings.TrimSpace(r.FormValue("date_of_birth"))
			if v := strings.TrimSpace(r.FormValue("nickname")); v != "" {
				nickname = &v
			}
			if v := strings.TrimSpace(r.FormValue("about_me")); v != "" {
				aboutMe = &v
			}

			file, header, err := r.FormFile("avatar")
			if err == nil && file != nil && header != nil {
				defer file.Close()
				_ = os.MkdirAll("uploads", 0755)
				ext := strings.ToLower(filepath.Ext(header.Filename))
				switch ext {
				case ".jpg", ".jpeg", ".png", ".gif":
					// allowed types
				default:
					writeError(w, http.StatusBadRequest, "avatar must be JPG, PNG, or GIF")
					return
				}

				filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
				dstPath := filepath.Join("uploads", filename)
				dst, err := os.Create(dstPath)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "failed to save avatar")
					return
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					writeError(w, http.StatusInternalServerError, "failed to save avatar")
					return
				}

				p := "/uploads/" + filename
				avatarURL = &p
			}
		} else {
			// JSON body
			type reqStruct struct {
				FullName    *string `json:"full_name"`
				DateOfBirth *string `json:"date_of_birth"`
				Nickname    *string `json:"nickname"`
				AboutMe     *string `json:"about_me"`
			}
			var req reqStruct
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			if req.FullName != nil {
				fullName = strings.TrimSpace(*req.FullName)
			}
			if req.DateOfBirth != nil {
				dateOfBirth = strings.TrimSpace(*req.DateOfBirth)
			}
			if req.Nickname != nil {
				n := strings.TrimSpace(*req.Nickname)
				if n != "" {
					nickname = &n
				}
			}
			if req.AboutMe != nil {
				a := strings.TrimSpace(*req.AboutMe)
				if a != "" {
					aboutMe = &a
				}
			}
		}

		// keep existing values where not provided
		if fullName == "" {
			fullName = user.FullName
		}
		if dateOfBirth == "" {
			dateOfBirth = user.DateOfBirth
		}
		if nickname == nil {
			nickname = user.Nickname
		}
		if aboutMe == nil {
			aboutMe = user.AboutMe
		}
		if avatarURL == nil {
			avatarURL = user.AvatarURL
		}

		if err := db.UpdateUserProfile(s.DB, user.ID, fullName, dateOfBirth, avatarURL, nickname, aboutMe); err != nil {
			log.Println("UpdateUserProfile error:", err)
			writeError(w, http.StatusInternalServerError, "failed to update profile")
			return
		}

		updated, err := db.GetUserByID(s.DB, user.ID)
		if err != nil {
			log.Println("GetUserByID after update error:", err)
			writeError(w, http.StatusInternalServerError, "failed to reload user")
			return
		}
		resp := map[string]any{
			"id":            updated.ID,
			"uuid":          updated.UUID,
			"email":         updated.Email,
			"full_name":     updated.FullName,
			"date_of_birth": updated.DateOfBirth,
			"avatar_url":    updated.AvatarURL,
			"nickname":      updated.Nickname,
			"about_me":      updated.AboutMe,
			"is_private":    updated.IsPrivate,
			"created_at":    updated.CreatedAt,
		}
		writeJSON(w, http.StatusOK, resp)
		return

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
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

	// Check if there's a pending follow request from current user to target
	isPending := false
	followStatus, err := db.GetFollowStatus(s.DB, int(currentUser.ID), int(targetUser.ID))
	if err == nil && followStatus != nil && followStatus.Status == "pending" {
		isPending = true
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
				"is_pending":   isPending,
				"is_self":      false,
				"can_view":     false,
			},
		})
		return
	}

	// Public profile OR follower => show profile + visible posts (public + almost-private)
	posts, err := db.GetPostsByUserVisibleForViewer(
		s.DB,
		int(targetUser.ID),  // whose profile
		int(currentUser.ID), // viewer
		isFollowing,         // includeAlmost
	)
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
			"is_pending":   isPending,
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
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var content string
	privacy := "public"
	imagePath := ""
	allowedViewers := []int{}

	ct := r.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
			writeError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}

		content = strings.TrimSpace(r.FormValue("content"))
		if v := strings.TrimSpace(r.FormValue("privacy")); v != "" {
			privacy = v
		}

		if av := strings.TrimSpace(r.FormValue("allowed_viewers")); av != "" {
			_ = json.Unmarshal([]byte(av), &allowedViewers)
		}

		// image file: field name MUST be "image"
		file, header, err := r.FormFile("image")
		if err == nil && file != nil && header != nil {
			defer file.Close()

			_ = os.MkdirAll("uploads", 0755)

			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".png"
			}

			filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
			dstPath := filepath.Join("uploads", filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save image")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to save image")
				return
			}

			imagePath = "/uploads/" + filename
		}

	} else {
		var req struct {
			Content        string `json:"content"`
			Privacy        string `json:"privacy"`
			ImagePath      string `json:"image_path,omitempty"`
			AllowedViewers []int  `json:"allowed_viewers,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		content = strings.TrimSpace(req.Content)
		if strings.TrimSpace(req.Privacy) != "" {
			privacy = req.Privacy
		}
		imagePath = req.ImagePath
		allowedViewers = req.AllowedViewers
	}

	if content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	post, err := db.InsertPostWithExtras(
		s.DB,
		int(user.ID),
		content,
		privacy,
		imagePath,
		allowedViewers,
	)
	if err != nil {
		log.Println("CreatePost error:", err)
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

	posts, err := db.GetPostsByUserVisibleForViewer(s.DB, int(user.ID), int(user.ID), true)
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

// 3 post sroutes
// /api/posts/{id}/comments  (GET, POST)
// /api/posts/{id}/like      (POST)

type commentCreateRequest struct {
	Content   string `json:"content"`
	ImagePath string `json:"image_path,omitempty"`
}

type commentResponse struct {
	ID         int    `json:"id"`
	PostID     int    `json:"post_id"`
	UserID     int64  `json:"user_id"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	ImagePath  string `json:"image_path"`
	CreatedAt  string `json:"created_at"`
}

func (s *Server) handlePostRoutes(w http.ResponseWriter, r *http.Request) {
	// /api/posts/{id}/comments
	// /api/posts/{id}/like
	rest := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")

	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	postID, err := strconv.Atoi(parts[0])
	if err != nil || postID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	switch parts[1] {
	case "comments":
		s.handlePostComments(w, r, postID)
		return
	case "like":
		s.handlePostLike(w, r, postID)
		return
	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
}

func (s *Server) handlePostComments(w http.ResponseWriter, r *http.Request, postID int) {
	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := s.DB.Query(`
			SELECT c.id, c.post_id, c.user_id, u.full_name, c.content, COALESCE(c.image_path, ''), c.created_at
			FROM post_comments c
			JOIN users u ON u.id = c.user_id
			WHERE c.post_id = ?
			ORDER BY c.created_at ASC
		`, postID)
		if err != nil {
			log.Println("list comments query error:", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch comments")
			return
		}
		defer rows.Close()

		out := []commentResponse{}
		for rows.Next() {
			var c commentResponse
			if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.AuthorName, &c.Content, &c.ImagePath, &c.CreatedAt); err != nil {
				log.Println("list comments scan error:", err)
				writeError(w, http.StatusInternalServerError, "failed to fetch comments")
				return
			}
			out = append(out, c)
		}
		writeJSON(w, http.StatusOK, out)
		return

	case http.MethodPost:
		var req commentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if strings.TrimSpace(req.Content) == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}

		res, err := s.DB.Exec(`
			INSERT INTO post_comments (post_id, user_id, content, image_path)
			VALUES (?, ?, ?, ?)
		`, postID, user.ID, req.Content, req.ImagePath)
		if err != nil {
			log.Println("create comment insert error:", err)
			writeError(w, http.StatusInternalServerError, "failed to create comment")
			return
		}

		newID, _ := res.LastInsertId()

		// Return the created comment
		var c commentResponse
		err = s.DB.QueryRow(`
			SELECT c.id, c.post_id, c.user_id, u.full_name, c.content, COALESCE(c.image_path, ''), c.created_at
			FROM post_comments c
			JOIN users u ON u.id = c.user_id
			WHERE c.id = ?
		`, newID).Scan(&c.ID, &c.PostID, &c.UserID, &c.AuthorName, &c.Content, &c.ImagePath, &c.CreatedAt)

		if err != nil {
			log.Println("create comment fetch error:", err)
			// still ok to return success without full record
			writeJSON(w, http.StatusCreated, map[string]any{
				"id": newID,
			})
			return
		}

		writeJSON(w, http.StatusCreated, c)
		return

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

func (s *Server) handlePostLike(w http.ResponseWriter, r *http.Request, postID int) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := CurrentUser(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// check if like exists
	var existing int
	if err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM post_likes
		WHERE post_id = ? AND user_id = ?
	`, postID, user.ID).Scan(&existing); err != nil {
		log.Println("like check error:", err)
		writeError(w, http.StatusInternalServerError, "failed to toggle like")
		return
	}

	liked := false
	if existing > 0 {
		_, err := s.DB.Exec(`
			DELETE FROM post_likes
			WHERE post_id = ? AND user_id = ?
		`, postID, user.ID)
		if err != nil {
			log.Println("unlike error:", err)
			writeError(w, http.StatusInternalServerError, "failed to toggle like")
			return
		}
		liked = false
	} else {
		_, err := s.DB.Exec(`
			INSERT INTO post_likes (post_id, user_id)
			VALUES (?, ?)
		`, postID, user.ID)
		if err != nil {
			log.Println("like insert error:", err)
			writeError(w, http.StatusInternalServerError, "failed to toggle like")
			return
		}
		liked = true
	}

	// return updated count
	var likeCount int
	if err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM post_likes
		WHERE post_id = ?
	`, postID).Scan(&likeCount); err != nil {
		log.Println("like count error:", err)
		writeError(w, http.StatusInternalServerError, "failed to toggle like")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"liked":      liked,
		"like_count": likeCount,
	})
}
