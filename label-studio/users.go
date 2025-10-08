package labelstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// User represents a user in Label Studio
type User struct {
	ID                     int                    `json:"id"`
	FirstName              string                 `json:"first_name"`
	LastName               string                 `json:"last_name"`
	Username               string                 `json:"username"`
	Email                  string                 `json:"email"`
	LastActivity           time.Time              `json:"last_activity"`
	CustomHotkeys          map[string]interface{} `json:"custom_hotkeys"`
	Avatar                 *string                `json:"avatar"`
	Initials               string                 `json:"initials"`
	Phone                  string                 `json:"phone"`
	ActiveOrganization     int                    `json:"active_organization"`
	ActiveOrganizationMeta struct {
		Title string `json:"title"`
		Email string `json:"email"`
	} `json:"active_organization_meta"`
	AllowNewsletters *bool     `json:"allow_newsletters"`
	DateJoined       time.Time `json:"date_joined"`
}

// ListUsers fetches the list of users from the Label Studio API
func (c *Client) ListUsers() ([]User, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/users", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		// Try to read error message
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API returned error %d: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	// Parse response
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return users, nil
}

// GetUser fetches a specific user by ID
func (c *Client) GetUser(userID int) (*User, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%d", c.BaseURL, userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		// Try to read error message
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API returned error %d: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	// Parse response
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}
