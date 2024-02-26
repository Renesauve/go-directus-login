package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	startServer()
	directusURL, directusEmail, directusPassword := loadEnv()
	userService := users.NewUserService(directusURL, directusEmail, directusPassword)

	setupRoutes(userService)
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

	// Handling shutdown in a separate goroutine
	go func() {
		<-waitForShutdownSignal()
		log.Println("Shutdown signal received, shutting down server...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Server Shutdown: %v", err)
		}
	}()

	log.Println("Server starting on port 8080...")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
	log.Println("Server stopped")
}

func waitForShutdownSignal() <-chan struct{} {
	quit := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("Server is stopping due to %+v", sig)
		close(done)
	}()

	return done
}

func serveRegistrationPage(w http.ResponseWriter, r *http.Request) {
	// Serve the registration HTML page
	http.ServeFile(w, r, "web/registration.html")
}

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
		// Ensure method is POST
		if r.Method != "POST" {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Call centralized registration logic
		err := userService.RegisterUser(email, password)
		if err != nil {
			// You can refine the error handling here to provide more specific messages based on the error returned
			log.Printf("Registration error: %v", err)
			if strings.Contains(err.Error(), "awaiting verification") {
				http.Error(w, "Please check your email to verify your account.", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "already exists") {
				http.Error(w, "User already exists. Please login.", http.StatusBadRequest)
			} else {
				http.Error(w, "Failed to process registration", http.StatusInternalServerError)
			}
			return
		}

		// Registration successful, send a verification email
		// Note: Assuming the actual sending of the verification email happens within userService.RegisterUser
		fmt.Fprintf(w, "Registration successful. Please check your email to verify your account.")
	})
}
