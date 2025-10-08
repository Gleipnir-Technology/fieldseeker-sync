package labelstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AnnotationRequest represents the request body for creating a draft
type AnnotationRequest struct {
	DraftID          int          `json:"draft_id"`
	LeadTime         float64      `json:"lead_time"`
	ParentAnnotation *int         `json:"parent_annotation,omitempty"`
	ParentPrediction *int         `json:"parent_prediction,omitempty"`
	Project          string       `json:"project"`
	Result           []TaskResult `json:"result"`
	StartedAt        string       `json:"started_at"`
}

// Annotation represents a draft annotation returned by the API
type Annotation struct {
	BulkCreated      bool         `json:"bulk_created"`
	CompletedBy      int          `json:"completed_by"`
	CreatedAgo       string       `json:"created_ago"`
	CreatedAt        string       `json:"created_at"`
	CreatedUsername  string       `json:"created_username"`
	DraftCreatedAt   string       `json:"draft_created_at"`
	GroundTruth      bool         `json:"ground_truth"`
	ID               int          `json:"id"`
	ImportID         *string      `json:"import_id"`
	LastAction       *string      `json:"last_action"`
	LastCreatedBy    *string      `json:"last_created_by"`
	LeadTime         float64      `json:"lead_time"`
	ParentAnnotation *int         `json:"parent_annotation,omitempty"`
	ParentPrediction *int         `json:"parent_prediction,omitempty"`
	Project          int          `json:"project"`
	Result           []TaskResult `json:"result"`
	Task             int          `json:"task"`
	WasCancelled     bool         `json:"was_cancelled"`
	UpdatedAt        string       `json:"updated_at"`
	UpdatedBy        string       `json:"updated_by"`
}

// NewAnnotation creates a new draft request builder
func NewAnnotationRequest(projectID int) *AnnotationRequest {
	return &AnnotationRequest{
		Project:   string(projectID),
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// CreateAnnotation creates a new draft for a task
func (c *Client) CreateAnnotation(taskID int, draft *AnnotationRequest) (*Annotation, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// Marshal the draft request to JSON
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal draft request: %w", err)
	}

	// Create request URL with query parameter
	//url := fmt.Sprintf("%s/api/tasks/%d/drafts?project=%s", c.BaseURL, taskID, draft.Project)
	url := fmt.Sprintf("%s/api/tasks/%d/drafts", c.BaseURL, taskID)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(draftJSON))
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
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Try to read error message
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API returned error %d: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	// Parse response
	var createdAnnotation Annotation
	if err := json.NewDecoder(resp.Body).Decode(&createdAnnotation); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createdAnnotation, nil
}
