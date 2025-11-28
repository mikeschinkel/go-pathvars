// Package main demonstrates a complete REST API using PathVars.
//
// This example shows:
// - Full CRUD operations (Create, Read, Update, Delete)
// - Multiple HTTP methods (GET, POST, PUT, DELETE)
// - Complex validation constraints
// - UUID-based resource identification
// - Query parameter filtering and pagination
// - Proper HTTP status codes
// - JSON responses
//
// Run with: go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/mikeschinkel/go-pathvars"
)

// User represents a user resource
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserStore is a simple in-memory user database
type UserStore struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]User),
	}
}

func (s *UserStore) Create(user User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

func (s *UserStore) Get(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

func (s *UserStore) Update(user User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.ID]; !exists {
		return false
	}
	s.users[user.ID] = user
	return true
}

func (s *UserStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[id]; !exists {
		return false
	}
	delete(s.users, id)
	return true
}

func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// API routes
const (
	RouteListUsers = iota
	RouteCreateUser
	RouteGetUser
	RouteUpdateUser
	RouteDeleteUser
	RouteHealthCheck
)

func main() {
	// Create user store
	store := NewUserStore()

	// Seed with sample data
	store.Create(User{
		ID:    "550e8400-e29b-41d4-a716-446655440000",
		Name:  "Alice",
		Email: "alice@example.com",
	})
	store.Create(User{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Name:  "Bob",
		Email: "bob@example.com",
	})

	// Create and configure router
	router := pathvars.NewRouter()

	// Define routes
	routes := []struct {
		method string
		path   string
		index  int
	}{
		{"GET", "/users?{limit?10:int:range[1..100]}&{offset?0:int:range[0..1000]}", RouteListUsers},
		{"POST", "/users", RouteCreateUser},
		{"GET", "/users/{id:uuid}", RouteGetUser},
		{"PUT", "/users/{id:uuid}", RouteUpdateUser},
		{"DELETE", "/users/{id:uuid}", RouteDeleteUser},
		{"GET", "/health", RouteHealthCheck},
	}

	for _, route := range routes {
		if err := router.AddRoute(
			pathvars.HTTPMethod(route.method),
			pathvars.Template(route.path),
			nil,
		); err != nil {
			log.Fatalf("Failed to add route %s %s: %v", route.method, route.path, err)
		}
	}

	// Routes are compiled as they are added - ready to use!

	// HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Match the request
		result, err := router.Match(r)
		if err != nil {
			sendError(w, http.StatusNotFound, "Route not found", err)
			return
		}

		// Route to appropriate handler
		switch result.Index {
		case RouteListUsers:
			handleListUsers(w, r, store, result)
		case RouteCreateUser:
			handleCreateUser(w, r, store)
		case RouteGetUser:
			handleGetUser(w, r, store, result)
		case RouteUpdateUser:
			handleUpdateUser(w, r, store, result)
		case RouteDeleteUser:
			handleDeleteUser(w, r, store, result)
		case RouteHealthCheck:
			handleHealthCheck(w, r)
		default:
			sendError(w, http.StatusInternalServerError, "Unknown route", nil)
		}
	})

	// Start server
	fmt.Println("REST API server starting on :8080")
	fmt.Println()
	fmt.Println("Sample API calls:")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println("  curl http://localhost:8080/users")
	fmt.Println("  curl http://localhost:8080/users/550e8400-e29b-41d4-a716-446655440000")
	fmt.Println("  curl -X POST -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"id\":\"123e4567-e89b-12d3-a456-426614174000\",\"name\":\"Charlie\",\"email\":\"charlie@example.com\"}' \\")
	fmt.Println("    http://localhost:8080/users")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Handler functions

func handleListUsers(w http.ResponseWriter, r *http.Request, store *UserStore, result pathvars.MatchResult) {
	limit, _ := result.GetValue("limit")   // Default: 10
	offset, _ := result.GetValue("offset") // Default: 0

	users := store.List()

	// Simple pagination (in production, use proper database pagination)
	// For demo purposes only

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
		"total":  len(users),
	})
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, store *UserStore) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Validate required fields
	if user.ID == "" || user.Name == "" || user.Email == "" {
		sendError(w, http.StatusBadRequest, "Missing required fields (id, name, email)", nil)
		return
	}

	store.Create(user)
	sendJSON(w, http.StatusCreated, user)
}

func handleGetUser(w http.ResponseWriter, r *http.Request, store *UserStore, result pathvars.MatchResult) {
	id, _ := result.GetValue("id")

	user, ok := store.Get(id.(string))
	if !ok {
		sendError(w, http.StatusNotFound, "User not found", nil)
		return
	}

	sendJSON(w, http.StatusOK, user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request, store *UserStore, result pathvars.MatchResult) {
	id, _ := result.GetValue("id")

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	user.ID = id.(string) // Ensure ID matches URL

	if !store.Update(user) {
		sendError(w, http.StatusNotFound, "User not found", nil)
		return
	}

	sendJSON(w, http.StatusOK, user)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, store *UserStore, result pathvars.MatchResult) {
	id, _ := result.GetValue("id")

	if !store.Delete(id.(string)) {
		sendError(w, http.StatusNotFound, "User not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Helper functions

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	sendJSON(w, status, response)
}
