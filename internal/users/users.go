// Contents of /internal/users/users.go
package users

import (
	"navi/pkg/directus" // Update this import path to match your module's actual path
	"os"
)

type UserProfile struct {
	Email string
	Name  string
	// Add other fields as necessary
}
type UserService struct {
	DirectusClient *directus.Client
}

func NewUserService(directusURL, adminEmail, adminPassword string) *UserService {
	client := directus.NewClient(directusURL, adminEmail, adminPassword)
	return &UserService{
		DirectusClient: client,
	}
}

func (s *UserService) Authenticate(email, password string) (string, error) {
	// Directly return the call to DirectusClient's Authenticate method
	return s.DirectusClient.Authenticate(email, password)
}

func (s *UserService) CreateUser(email, password string) (string, error) {
	// Note: Removed the token parameter since it's likely not needed for user registration via Directus API,
	// but adjust according to your Directus setup and permissions.
	UUID := os.Getenv("DIRECTUS_USER_UUID")
	return s.DirectusClient.CreateUser(email, password, UUID)
}

func (s *UserService) GetUserDetails(email string) (UserProfile, error) {
	// Placeholder: Fetch user details from your storage.
	// This might involve querying a database or calling an external API.
	// Here, we'll return a dummy user profile.
	return UserProfile{
		Email: email,
		Name:  "John Doe",
	}, nil
}
