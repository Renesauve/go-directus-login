package main

import (
	"fmt"
	"navi/internal/users" // Ensure this import path is correct based on your project structure
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// It's good practice to separate app initialization and window creation into their own functions for better readability and maintenance.
	myApp := app.New()
	myWindow := createRegistrationWindow(myApp)

	myWindow.ShowAndRun()
}

func createRegistrationWindow(app fyne.App) fyne.Window {
	myWindow := app.NewWindow("Register")

	// Consider moving user service initialization outside of your main or UI logic if it's used in multiple places or requires initial setup.
	directusURL := os.Getenv("DIRECTUS_URL")
	directusEmail := os.Getenv("DIRECTUS_EMAIL")
	directusPassword := os.Getenv("DIRECTUS_PASSWORD")

	userService := users.NewUserService(directusURL, directusEmail, directusPassword)

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	registerBtn := widget.NewButton("Register", func() {
		registerUser(userService, emailEntry.Text, passwordEntry.Text, myWindow)
	})

	myWindow.SetContent(container.NewVBox(
		emailEntry,
		passwordEntry,
		registerBtn,
	))

	return myWindow
}

// Separating the user registration logic into its own function improves readability and makes the code easier to manage.
func registerUser(userService *users.UserService, email, password string, window fyne.Window) {
	token, err := userService.Authenticate()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Authentication failed: %v", err), window)
		return
	}

	userID, err := userService.CreateUser(token, email, password)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to create user: %v", err), window)
		return
	}

	dialog.ShowInformation("Success", fmt.Sprintf("User created successfully. User ID: %s", userID), window)
}
