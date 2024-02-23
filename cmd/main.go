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
		// Form handling logic...
	})
}
