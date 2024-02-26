package directus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	URL      string
	Email    string
	Password string
}

type AuthResponse struct {
	Data struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type TempUser struct {
	Email             string `json:"email"`
	Password          string `json:"password"` // Add this line
	VerificationToken string `json:"verification_token"`
	ExpirationDate    string `json:"expiration_date"`
}

func NewClient(url, email, password string) *Client {
	return &Client{
		URL:      url,
		Email:    email,
		Password: password,
	}
}

func (c *Client) Authenticate(email, password string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(c.URL+"/auth/login", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.AccessToken, nil
}

// CreateTempUser creates a temporary user if one does not already exist with the given email.
func (c *Client) CreateTempUser(email, password, verificationToken string) error {
	// First, check if a temporary user already exists with the given email

	// Proceed with creating the temporary user if they don't exist

	exists, collection, err := c.UserExists(email)
	if err != nil {
		return fmt.Errorf("error checking if user exists: %v", err)
	}
	if exists {
		// If the user exists, return an error or handle as needed.
		return fmt.Errorf("a user with this email already exists in %s", strconv.FormatBool(collection))
	}

	tempUserData := TempUser{
		Email:             email,
		Password:          password,
		VerificationToken: verificationToken,
		ExpirationDate:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	userDataBytes, err := json.Marshal(tempUserData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.URL+"/items/temp_users", bytes.NewBuffer(userDataBytes))
	if err != nil {
		return err
	}

	token := os.Getenv("DIRECTUS_ADMIN_TOKEN")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) UserExists(email string) (exists bool, inTempUsers bool, err error) {
	// Check in 'users' collection first.
	exists, err = c.CheckEmailInCollection(email, "users")
	if err != nil || exists {
		return exists, false, err
	}

	// If not found in 'users', check in 'temp_users'.
	exists, err = c.CheckEmailInCollection(email, "temp_users")
	return exists, exists, err
}

// CheckEmailInCollection checks for an email's existence in a specified collection.
func (c *Client) CheckEmailInCollection(email, collection string) (bool, error) {
	// Adjust the endpoint based on the collection being queried
	var url string
	if collection == "users" {
		// Directly use the /users endpoint for querying users
		url = fmt.Sprintf("%s/users?filter[email][_eq]=%s", c.URL, email)
	} else {
		// Use the /items/{collection} format for other collections like temp_users
		url = fmt.Sprintf("%s/items/%s?filter[email][_eq]=%s", c.URL, collection, email)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	token := os.Getenv("DIRECTUS_ADMIN_TOKEN")
	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var response struct {
		Data []interface{} `json:"data"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}

	// If data is not empty, the email exists in the collection.
	return len(response.Data) > 0, nil
}
