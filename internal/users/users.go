// Contents of /internal/users/users.go
package users

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"navi/pkg/directus" // Update this import path to match your module's actual path
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

func (s *UserService) GetUserDetails(email string) (UserProfile, error) {

	return UserProfile{
		Email: email,
		Name:  "John Doe",
	}, nil
}

func (s *UserService) CreateTempUser(email, password, verificationToken string) error {
	// Example logic to create a temporary user, potentially calling DirectusClient
	err := s.DirectusClient.CreateTempUser(email, password, verificationToken) // Adjust based on actual implementation
	if err != nil {
		return err // Use '=' here since 'err' is already declared in the scope by the short variable declaration
	}
	return nil
}

func (s *UserService) UserExists(email string) (exists bool, inTempUsers bool, err error) {
	// Use the DirectusClient to check the 'users' collection.
	exists, err = s.DirectusClient.CheckEmailInCollection(email, "users")
	if err != nil || exists {
		return exists, false, err
	}

	// If not found in 'users', check in 'temp_users'.
	exists, err = s.DirectusClient.CheckEmailInCollection(email, "temp_users")

	return exists, exists, err
}

func (s *UserService) RegisterUser(email, password string) error {
	exists, inTempUsers, err := s.UserExists(email)

	if err != nil {
		return fmt.Errorf("error checking user existence: %v", err)
	}
	if exists {
		if inTempUsers {
			return fmt.Errorf("user already exists and is awaiting verification")
		}
		return fmt.Errorf("user already exists")
	}

	verificationToken, err := generateSecureToken(16) // Assuming this is implemented correctly
	if err != nil {
		return fmt.Errorf("error generating verification token: %v", err)
	}

	return s.CreateTempUser(email, password, verificationToken)
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
