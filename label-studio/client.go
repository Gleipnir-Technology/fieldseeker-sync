package labelstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Label Studio API client
type Client struct {
	BaseURL     string
	APIKey      string
	AccessToken string
	AccessTokenExpires time.Time
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

// According to https://github.com/HumanSignal/label-studio/blob/develop/docs/source/guide/access_tokens.md
// the access tokens expire "in about 5 minutes". We'll do 4 minutes to give us a bit of margin.
var ACCESS_TOKEN_DURATION_SECONDS time.Duration = 240 * time.Second

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
	c.AccessTokenExpires = time.Now().Add(ACCESS_TOKEN_DURATION_SECONDS)
	return nil
}

func (c *Client) makeRequest(method string, path string, payload []byte) (*http.Response, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" || time.Now().After(c.AccessTokenExpires) {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}
	// Create request
	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check for successful response
	if resp.StatusCode > http.StatusBadRequest {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("Got status code %d and failed to read response body: %v", resp.StatusCode, err)
		}
		bodyString := string(bodyBytes)
		// Try to read error message
		var errorResp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			return nil, fmt.Errorf("API returned JSON error %d: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("API returned error status %d: %s: ", resp.Status, bodyString)
	}


	return resp, nil
}
