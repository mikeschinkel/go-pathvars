// Package main demonstrates basic routing with PathVars.
//
// This example shows:
// - Creating a router
// - Adding routes with typed parameters
// - Compiling routes
// - Matching incoming requests
// - Extracting and using parameter values
//
// Run with: go run main.go
// Then test with:
//
//	curl http://localhost:8080/users/123
//	curl http://localhost:8080/users/abc        (validation error)
//	curl http://localhost:8080/posts/hello-world
//	curl http://localhost:8080/notfound         (404)
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mikeschinkel/go-pathvars"
)

func main() {
	// Create a new router
	router := pathvars.NewRouter()

	// Add routes with typed parameters
	// Route 0: GET /users/{id:int}
	if err := router.AddRoute("GET", "/users/{id:int}", nil); err != nil {
		log.Fatalf("Failed to add route: %v", err)
	}

	// Route 1: GET /posts/{slug:slug:length[5..50]}
	if err := router.AddRoute("GET", "/posts/{slug:slug:length[5..50]}", nil); err != nil {
		log.Fatalf("Failed to add route: %v", err)
	}

	// Route 2: GET /products?{category:string}&{limit?20:int:range[1..100]}
	if err := router.AddRoute("GET", "/products?{category:string}&{limit?20:int:range[1..100]}", nil); err != nil {
		log.Fatalf("Failed to add route: %v", err)
	}

	// Routes are compiled as they are added - ready to use!

	// HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Match the request
		result, err := router.Match(r)
		if err != nil {
			// No route matched or validation failed
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusNotFound)
			return
		}

		// Handle based on route index
		switch result.Index {
		case 0:
			// GET /users/{id:int}
			userID, _ := result.GetValue("id")
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "User ID: %s\n", userID)
			fmt.Fprintf(w, "Matched route: %s %s\n", result.Route.Method, result.Route.ParsedTemplate.String())

		case 1:
			// GET /posts/{slug:slug:length[5..50]}
			slug, _ := result.GetValue("slug")
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Post slug: %s\n", slug)
			fmt.Fprintf(w, "Matched route: %s %s\n", result.Route.Method, result.Route.ParsedTemplate.String())

		case 2:
			// GET /products?{category:string}&{limit?20:int:range[1..100]}
			category, _ := result.GetValue("category")
			limit, _ := result.GetValue("limit")
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Product category: %s\n", category)
			fmt.Fprintf(w, "Limit: %s (default: 20)\n", limit)
			fmt.Fprintf(w, "Matched route: %s %s\n", result.Route.Method, result.Route.ParsedTemplate.String())

		default:
			http.Error(w, "Unknown route", http.StatusInternalServerError)
		}
	})

	// Start server
	fmt.Println("Server starting on :8080")
	fmt.Println("Try these URLs:")
	fmt.Println("  http://localhost:8080/users/123")
	fmt.Println("  http://localhost:8080/users/abc         (validation error)")
	fmt.Println("  http://localhost:8080/posts/hello-world")
	fmt.Println("  http://localhost:8080/posts/hi          (too short)")
	fmt.Println("  http://localhost:8080/products?category=electronics")
	fmt.Println("  http://localhost:8080/products?category=books&limit=50")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
