package app

import (
	"database/sql"
	"log"
	"net/http"
)

type Server struct {
	
		DB  *sql.DB
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

	// need to add more pages : mux.HandleFunc("/register", s.handleRegister) & login etc

	return s
}
//starting the server
func (s *Server) Listen(addr string) error {
	log.Println("Starting server on", addr)
	return http.ListenAndServe(addr, s.Mux)
}
