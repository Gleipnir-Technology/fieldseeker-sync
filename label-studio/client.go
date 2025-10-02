package labelstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client represents a Label Studio API client
type Client struct {
	BaseURL     string
	APIKey      string
	AccessToken string
	HTTPClient  *http.Client
}

// NewClient creates a new Label Studio client
func NewClient(baseURL string, apiKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// GetAccessToken converts the API key into an access token
func (c *Client) GetAccessToken() error {
	// Create request body
	reqBody := map[string]string{
		"refresh": c.APIKey,
	}

	// Marshal to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/token/refresh", c.BaseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned error: %s", resp.Status)
	}

	// Parse response
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Get access token
	accessToken, ok := result["access"]
	if !ok {
		return fmt.Errorf("response did not contain access token")
	}

	// Store access token
	c.AccessToken = accessToken
	return nil
}
