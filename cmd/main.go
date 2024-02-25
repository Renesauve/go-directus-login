package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"navi/internal/users" // Adjust the import path as needed

	"github.com/golang-jwt/jwt"
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
	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) { serveProfilePage(w, r, userService) })

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

// In main.go
func serveProfilePage(w http.ResponseWriter, r *http.Request, userService *users.UserService) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenString := sessionCookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTKey(), nil
	})

	if err != nil || !token.Valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	fmt.Println(claims)
	userEmail := claims.Email
	user, err := userService.GetUserDetails(userEmail)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("web/profile.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
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

		// Other registration logic...

		log.Printf("Checking existence for email: %s", email)
		exists, err := userService.UserExists(email)

		if err != nil {
			// Handle error, e.g., log it and return an internal server error response.
			log.Printf("Error checking if user exists: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			// Handle the case where the user already exists, e.g., return a specific error message.
			http.Error(w, "User already exists", http.StatusBadRequest) // Customize as needed
			return
		}

		password := r.FormValue("password")

		// Assuming you have a method to generate a secure token and hash the password
		verificationToken, err := generateSecureToken(16) // For a 32-character token
		if err != nil {
			log.Printf("Error generating verification token: %v", err)
			http.Error(w, "Failed to process registration", http.StatusInternalServerError)
			return
		}

		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Failed to process registration", http.StatusInternalServerError)
			return
		}

		// Use userService to store the user's data temporarily
		err = userService.CreateUser(email, password, verificationToken)

		if err != nil {
			log.Printf("Failed to store temporary user data: %v", err)
			http.Error(w, "Failed to process registration", http.StatusInternalServerError)
			return
		}

		// Send a verification email to the user

		fmt.Fprintf(w, "Registration successful. Please check your email to verify your account.")
	})
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
