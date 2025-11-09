package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Server represents an HTTP server with routing capabilities
type Server struct {
	addr    string
	router  *http.ServeMux
	logger  *log.Logger
	timeout time.Duration
}

// NewServer creates a new server instance
func NewServer(addr string, timeout time.Duration) *Server {
	return &Server{
		addr:    addr,
		router:  http.NewServeMux(),
		logger:  log.Default(),
		timeout: timeout,
	}
}

// RegisterRoutes sets up all HTTP routes
func (s *Server) RegisterRoutes() {
	s.router.HandleFunc("/health", s.handleHealth)
	s.router.HandleFunc("/api/users", s.handleUsers)
	s.router.HandleFunc("/api/users/", s.handleUserByID)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUsers handles listing all users
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users := []map[string]string{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)

	case http.MethodPost:
		var user map[string]string
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleUserByID handles operations on individual users
func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	user := map[string]string{
		"id":   id,
		"name": "User " + id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	s.logger.Printf("Starting server on %s", s.addr)
	server := &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  s.timeout,
		WriteTimeout: s.timeout,
	}
	return server.ListenAndServe()
}

func main() {
	server := NewServer(":8080", 30*time.Second)
	server.RegisterRoutes()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
