package directus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io" // Use io package instead of ioutil
	"net/http"
	// Potentially used for other ioutil replacements
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

type CreateUserResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func NewClient(url, email, password string) *Client {
	return &Client{
		URL:      url,
		Email:    email,
		Password: password,
	}
}

func (c *Client) Authenticate() (string, error) {
	authData := map[string]string{
		"email":    c.Email,
		"password": c.Password,
	}
	authDataBytes, _ := json.Marshal(authData)
	resp, err := http.Post(c.URL+"/auth/login", "application/json", bytes.NewBuffer(authDataBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	return authResp.Data.AccessToken, nil
}

func (c *Client) CreateUser(token, email, password string) (string, error) {
	userData := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	userDataBytes, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", c.URL+"/users", bytes.NewBuffer(userDataBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var createUserResp CreateUserResponse
		if err := json.NewDecoder(resp.Body).Decode(&createUserResp); err != nil {
			return "", err
		}
		return createUserResp.Data.ID, nil
	} else {
		bodyBytes, err := io.ReadAll(resp.Body) // Replace ioutil.ReadAll with io.ReadAll
		if err != nil {
			return "", err // Handle potential errors from io.ReadAll
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
}
