package main

import (
	"fmt"
	"navi/internal/users"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Register")

	// Initialize userService with Directus credentials
	userService := users.NewUserService("http://navi.alberni.dev", "rdasauve@gmail.com", "bambi123")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	registerBtn := widget.NewButton("Register", func() {
		// Use userService to register the user
		email := emailEntry.Text
		password := passwordEntry.Text
		token, err := userService.Authenticate() // Authenticate to get a token for user creation
		if err != nil {
			fmt.Println("Authentication failed:", err)
			return
		}
		userID, err := userService.CreateUser(token, email, password) // Adjusted to pass email and password
		if err != nil {
			fmt.Println("Failed to create user:", err)
			return
		}
		fmt.Printf("User created successfully. User ID: %s\n", userID)
	})

	myWindow.SetContent(container.NewVBox(
		emailEntry,
		passwordEntry,
		registerBtn,
	))

	myWindow.ShowAndRun()
}
