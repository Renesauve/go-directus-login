package directus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func (c *Client) CreateUser(email, password, verificationToken string) error {
	tempUserData := TempUser{
		Email:             email,
		Password:          password, // Add the actual password here
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
