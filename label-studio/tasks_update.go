package labelstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdateTask updates specific fields of a task in Label Studio
// The updates parameter should contain only the fields that need to be changed
func (c *Client) UpdateTask(taskID int, updates interface{}) (*Task, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}
	
	// Marshal the updates to JSON
	updateJSON, err := json.Marshal(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updates: %w", err)
	}
	
	// Create request
	url := fmt.Sprintf("%s/api/tasks/%d", c.BaseURL, taskID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(updateJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	
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
	var updatedTask Task
	if err := json.NewDecoder(resp.Body).Decode(&updatedTask); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &updatedTask, nil
}
