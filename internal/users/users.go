// Contents of /internal/users/users.go
package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"navi/pkg/directus" // Update this import path to match your module's actual path
	"net/http"
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

func (s *UserService) GetUserDetails(email string) (UserProfile, error) {

	return UserProfile{
		Email: email,
		Name:  "John Doe",
	}, nil
}

func (s *UserService) CreateUser(email, password, verificationToken string) error {
	// Example logic to create a temporary user, potentially calling DirectusClient
	err := s.DirectusClient.CreateUser(email, password, verificationToken) // Adjust based on actual implementation
	if err != nil {
		return err // Use '=' here since 'err' is already declared in the scope by the short variable declaration
	}
	return nil
}

func (s *UserService) UserExists(email string) (bool, error) {
	// Construct the URL for querying the users collection by email.
	// Adjust the URL based on your Directus version and setup.
	url := fmt.Sprintf("%s/users?filter[email][_eq]=%s", s.DirectusClient.URL, email)

	// Prepare the HTTP request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}

	// Include the authorization header with your Directus API token.
	// Make sure to replace "YOUR_DIRECTUS_API_TOKEN" with your actual token.
	req.Header.Add("Authorization", "Bearer "+os.Getenv("DIRECTUS_ADMIN_TOKEN"))

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("reading response body: %w", err)
	}

	var result struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("parsing response body: %w", err)
	}

	// If the data array is not empty, a user with the email exists.
	return len(result.Data) > 0, nil
}
