package main

import (
	"fmt"
	"log"
	"navi/internal/users" // Ensure this import path is correct based on your project structure
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	directusURL := os.Getenv("DIRECTUS_URL")

	if directusURL == "" || !(strings.HasPrefix(directusURL, "http://") || strings.HasPrefix(directusURL, "https://")) {
		log.Fatal("DIRECTUS_URL environment variable is not set or does not start with http:// or https://")
	}

	directusEmail := os.Getenv("DIRECTUS_EMAIL")
	if directusEmail == "" {
		log.Fatal("DIRECTUS_EMAIL environment variable is not set")
	}

	directusPassword := os.Getenv("DIRECTUS_PASSWORD")
	if directusPassword == "" {
		log.Fatal("DIRECTUS_PASSWORD environment variable is not set")
	}
	// Initialize userService with Directus credentials
	userService := users.NewUserService(directusURL, directusEmail, directusPassword)

	// Serve the HTML form
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/registration.html")
	})

	// Handle form submission
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate with Directus to get a token
		token, err := userService.Authenticate()
		if err != nil {
			fmt.Fprintf(w, "Failed to authenticate with Directus: %v", err)
			return
		}

		// Use the token to create a new user in Directus
		userID, err := userService.CreateUser(token, email, password)
		if err != nil {
			fmt.Fprintf(w, "Failed to create user: %v", err)
			return
		}

		// Registration was successful
		fmt.Fprintf(w, "Registration successful: User ID: %s", userID)
	})

	// Start the web server
	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
