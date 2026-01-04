package app

import (
	"encoding/json"
	"net/http"
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
