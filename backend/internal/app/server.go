package app

import (
	"database/sql"
	"log"
	"net/http"
)

type Server struct {
	DB *sql.DB
	// ServeMux is gos basic router it maps paths to handler functions
	Mux *http.ServeMux
}

// NewServer creates a new Server, sets up routes, and returns it.
func NewServer(db *sql.DB) *Server {
	//creates new http router
	mux := http.NewServeMux()

	s := &Server{
		DB:  db,
		Mux: mux,
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
mux.HandleFunc("/api/posts", s.CreatePost)
mux.HandleFunc("/api/feed", s.GetFeed)
mux.HandleFunc("/api/me/posts", s.GetMyPosts)

	// add posts, profiles, feed, bla bla

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

		// Only allow your frontend origin during dev
		if origin == "http://localhost:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
