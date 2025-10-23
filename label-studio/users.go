package labelstudio

import (
	"encoding/json"
	"fmt"
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
	resp, err := c.makeRequest("GET", "/api/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to GET /api/userls: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return users, nil
}

// GetUser fetches a specific user by ID
func (c *Client) GetUser(userID int) (*User, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("/api/users/%d", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to GET /api/users/%d: %w", userID, err)
	}
	defer resp.Body.Close()

	// Parse response
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}
