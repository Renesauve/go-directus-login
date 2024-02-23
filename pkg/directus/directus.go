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

	// Check if the status code is not 200 OK
	if resp.StatusCode != http.StatusOK {
		// Read the response body to include in the error message
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract and return the access token
	if accessToken, ok := result["data"].(map[string]interface{})["access_token"].(string); ok {
		return accessToken, nil
	}

	return "", fmt.Errorf("failed to get access token from response")
}

func (c *Client) CreateUser(email, password, roleID string) (string, error) {
	userData := map[string]interface{}{
		"email":    email,
		"password": password,
		"role":     roleID, // Add the role ID here
	}
	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.URL+"/users", bytes.NewBuffer(userDataBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eGAEVUshdvle0MIfMbGaF0LqBuwOFqTF")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var createUserResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&createUserResp); err != nil {
			return "", err
		}
		return createUserResp.Data.ID, nil
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
}
