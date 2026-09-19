package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
)

// allowedOrigins is the set of browser origins permitted to call the API with
// credentials and to open a WebSocket. It defaults to the Vite dev server and
// can be overridden with ALLOWED_ORIGINS (comma separated) for other
// deployments — under docker-compose, for example, the app is served from
// http://localhost, not http://localhost:5173.
var allowedOrigins = loadAllowedOrigins()

func loadAllowedOrigins() map[string]bool {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:5173,http://127.0.0.1:5173"
	}

	origins := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins[o] = true
		}
	}
	return origins
}

// isAllowedOrigin reports whether origin may talk to this API.
func isAllowedOrigin(origin string) bool {
	return origin != "" && allowedOrigins[origin]
}

type Server struct {
	DB *sql.DB
	// ServeMux is gos basic router it maps paths to handler functions
	Mux *http.ServeMux
	// Hub manages WebSocket connections for real-time messaging
	Hub *Hub
}

// NewServer creates a new Server, sets up routes, and returns it.
func NewServer(db *sql.DB) *Server {
	//creates new http router
	mux := http.NewServeMux()

	// Create WebSocket hub for real-time messaging
	hub := NewHub(db)
	go hub.Run()

	s := &Server{
		DB:  db,
		Mux: mux,
		Hub: hub,
	}

	// checking to see if server is running smoothly
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/me", s.handleMe)
	mux.HandleFunc("/api/me/privacy", s.handleTogglePrivacy)
	mux.HandleFunc("/api/posts", s.CreatePost)
	mux.HandleFunc("/api/posts/", s.handlePostRoutes)
	mux.HandleFunc("/api/feed", s.GetFeed)
	mux.HandleFunc("/api/me/posts", s.GetMyPosts)
	mux.HandleFunc("/api/users/suggestions", s.GetSuggestedUsers)
	mux.HandleFunc("/api/users/search", s.SearchUsers) // User search - must be before /api/users/
	mux.HandleFunc("/api/me/followers", s.handleGetFollowers)
	mux.HandleFunc("/api/me/following", s.handleGetFollowing)
	mux.HandleFunc("/api/me/follow-requests", s.handleGetFollowRequests)
	mux.HandleFunc("/api/me/follow-counts", s.handleGetFollowCounts)
	mux.HandleFunc("/api/users/", s.handleUserFollowRoutes)
	mux.HandleFunc("/api/follow-requests/", s.handleFollowRequestRoutes)

	// Stage 6: Messages & Notifications
	mux.HandleFunc("/api/messages", s.handleMessageRoutes)
	mux.HandleFunc("/api/messages/", s.handleMessageRoutes)
	mux.HandleFunc("/api/notifications", s.handleNotificationRoutes)
	mux.HandleFunc("/api/notifications/", s.handleNotificationRoutes)
	mux.HandleFunc("/ws", s.HandleWebSocket(hub))
	mux.HandleFunc("/api/me/friends", s.handleGetFriends)
	mux.HandleFunc("/api/check-mutual", s.handleCheckMutual)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Stage 7: Groups
	mux.HandleFunc("/api/groups", s.handleGroups)
	mux.HandleFunc("/api/groups/invitations", s.handleGroupInvitations)
	mux.HandleFunc("/api/groups/invitations/", s.handleGroupInvitationRoutes)
	mux.HandleFunc("/api/groups/", s.handleGroupRoutes)

	return s
}

// starting the server
func (s *Server) Listen(addr string) error {
	log.Println("Starting server on", addr)

	// First apply auth then wrap everything in CORS
	handler := s.CORSMiddleware(s.AuthMiddleware(s.Mux))

	return http.ListenAndServe(addr, handler)
}

// CORSMiddleware adds CORS headers so the React dev server (localhost:5173) can talk to the Go API on localhost:8080 using cookies (added this when test failed)
func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only reflect origins that are explicitly allowed; never echo back an
		// arbitrary Origin, which would hand any site credentialed access.
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")

		}

		// Handle preflight requests directly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// For normal requests, continue down the chain
		next.ServeHTTP(w, r)
	})
}
