package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"navi/internal/users" // Adjust the import path as needed

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	directusURL, directusEmail, directusPassword := loadEnv()
	userService := users.NewUserService(directusURL, directusEmail, directusPassword)

	setupRoutes(userService)
	startServer()
}

func loadEnv() (string, string, string) {
	directusURL := os.Getenv("DIRECTUS_URL")
	directusEmail := os.Getenv("DIRECTUS_EMAIL")
	directusPassword := os.Getenv("DIRECTUS_PASSWORD")
	return directusURL, directusEmail, directusPassword
}

func setupRoutes(userService *users.UserService) {
	http.HandleFunc("/", serveRegistrationPage)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { loginHandler(w, r, userService) })
	http.HandleFunc("/profile", serveProfilePage)
	http.HandleFunc("/logout", logoutHandler)
	setupRegisterHandler(userService)
}

func startServer() {
	srv := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Server starting on port 8080...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func serveRegistrationPage(w http.ResponseWriter, r *http.Request) {
	// Serve the registration HTML page
	http.ServeFile(w, r, "web/registration.html")
}

func serveProfilePage(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in by checking the session token
	_, err := r.Cookie("session_token")
	if err != nil {
		// If not logged in, redirect to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Serve the profile HTML page
	http.ServeFile(w, r, "web/profile.html")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie to log out the user
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	// Redirect to the login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func loginHandler(w http.ResponseWriter, r *http.Request, userService *users.UserService) {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate the user using userService. Assume it returns an error for now.
		isAuthenticated, err := userService.Authenticate(email, password)

		if err != nil || isAuthenticated == "" {
			fmt.Fprintf(w, "Login failed: %v", err)
			return
		}

		// If authentication is successful, generate a JWT for the session
		tokenString, err := createToken(email) // Directly use createToken here
		if err != nil {
			log.Printf("Error creating token: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the JWT as a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
		})

		// Redirect to the profile page after successful login
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	} else {
		// For any non-POST request, serve the login page
		http.ServeFile(w, r, "web/login.html")
	}
}

// setupRegisterHandler configures the registration endpoint.
func setupRegisterHandler(userService *users.UserService) {
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Use userService to create a new user.
		userID, err := userService.CreateUser(email, password)
		if err != nil {
			log.Printf("Failed to register user: %v", err)
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "User registered successfully: %s", userID)
	})
}
