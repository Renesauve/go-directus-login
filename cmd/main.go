package main

import (
	"context"
	"fmt"
	"log"
	"navi/internal/users" // Adjust the import path as needed
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Environment variable checks...
	directusURL, directusEmail, directusPassword := loadEnv()

	// Initialize userService with Directus credentials
	userService := users.NewUserService(directusURL, directusEmail, directusPassword)

	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/registration.html")
	})

	// Handle form submission
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "web/login.html")
		} else if r.Method == "POST" {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Error parsing form", http.StatusBadRequest)
				return
			}

			email := r.FormValue("email")
			password := r.FormValue("password")

			// Attempt to login via Directus
			token, err := userService.Login(email, password)
			if err != nil {
				// Handle failed login
				fmt.Fprintf(w, "Login failed: %v", err)
				return
			}

			// Handle successful login, e.g., by setting a session cookie
			fmt.Fprintf(w, "Login successful, token: %s", token) // Adjust according to your needs
		}
	})
	setupRegisterHandler(userService)

	go func() {
		fmt.Println("Server starting on port 8080...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server exiting")
}

func loadEnv() (string, string, string) {
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

	return directusURL, directusEmail, directusPassword
}

func setupRegisterHandler(userService *users.UserService) {
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Use the userService to create a new user in Directus
		userID, err := userService.CreateUser(email, password) // Assuming CreateUser doesn't need a token for registration
		if err != nil {
			// Log the error and return a user-friendly message
			log.Printf("Failed to register user: %v", err)
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}

		// Respond to the client
		fmt.Fprintf(w, "User registered successfully: %s", userID)
	})
}
