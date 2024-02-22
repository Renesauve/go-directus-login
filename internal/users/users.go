// Contents of /internal/users/users.go
package users

import (
	"navi/pkg/directus" // Update this import path to match your module's actual path
)

type UserService struct {
	DirectusClient *directus.Client
}

func NewUserService(directusURL, adminEmail, adminPassword string) *UserService {
	return &UserService{
		DirectusClient: directus.NewClient(directusURL, adminEmail, adminPassword),
	}
}

func (s *UserService) Authenticate() (string, error) {
	return s.DirectusClient.Authenticate()
}

func (s *UserService) CreateUser(token, email, password string) (string, error) {
	// Modified to pass email and password as parameters
	return s.DirectusClient.CreateUser(token, email, password)
}
