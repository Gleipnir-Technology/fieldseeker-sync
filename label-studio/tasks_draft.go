package labelstudio

import (
	"encoding/json"
	"fmt"
	"time"
)

// DraftRequest represents the request body for creating a draft
type DraftRequest struct {
	Annotation       *string      `json:"annotation"`
	CreatedAgo       string       `json:"created_ago"`
	CreatedAt        string       `json:"created_at"`
	CreatedUsername  string       `json:"created_username"`
	DraftID          int          `json:"draft_id"`
	ID               int          `json:"id"`
	ImportID         *string      `json:"import_id"`
	LeadTime         float64      `json:"lead_time"`
	ParentAnnotation *int         `json:"parent_annotation,omitempty"`
	ParentPrediction *int         `json:"parent_prediction,omitempty"`
	Project          string       `json:"project"`
	Result           []TaskResult `json:"result"`
	StartedAt        string       `json:"started_at"`
	Task             int          `json:"task"`
	User             string       `json:"user"`
	WasPostponed     bool         `json:"was_postponed"`
}

// Draft represents a draft annotation returned by the API
type Draft struct {
	ID         int                      `json:"id"`
	TaskID     int                      `json:"task"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	LeadTime   float64                  `json:"lead_time"`
	Result     []map[string]interface{} `json:"result"`
	Annotation *int                     `json:"annotation,omitempty"`
	User       string                   `json:"user"`
}

// NewDraft creates a new draft request builder
func NewDraft(projectID int) *DraftRequest {
	return &DraftRequest{
		DraftID:   0,
		Project:   string(projectID),
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// SetLeadTime sets the time spent on the draft
func (d *DraftRequest) SetLeadTime(leadTime float64) *DraftRequest {
	d.LeadTime = leadTime
	return d
}

// SetResult sets the annotation result
func (d *DraftRequest) SetResult(result []TaskResult) *DraftRequest {
	d.Result = result
	return d
}

// SetParentPrediction sets the parent prediction ID if the draft is based on a prediction
func (d *DraftRequest) SetParentPrediction(predictionID int) *DraftRequest {
	d.ParentPrediction = &predictionID
	return d
}

// SetParentAnnotation sets the parent annotation ID if the draft is based on an annotation
func (d *DraftRequest) SetParentAnnotation(annotationID int) *DraftRequest {
	d.ParentAnnotation = &annotationID
	return d
}

// SetStartedAt sets the time when work on the draft started
func (d *DraftRequest) SetStartedAt(startedAt time.Time) *DraftRequest {
	d.StartedAt = startedAt.UTC().Format(time.RFC3339Nano)
	return d
}

// CreateDraft creates a new draft for a task
func (c *Client) CreateDraft(taskID int, draft *DraftRequest) (*Draft, error) {
	// Marshal the draft request to JSON
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal draft request: %w", err)
	}

	// Create request URL with query parameter
	//url := fmt.Sprintf("%s/api/tasks/%d/drafts?project=%s", c.BaseURL, taskID, draft.Project)
	path := fmt.Sprintf("/api/tasks/%d/drafts", taskID)

	// Create request
	resp, err := c.makeRequest("POST", path, draftJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	// Parse response
	var createdDraft Draft
	if err := json.NewDecoder(resp.Body).Decode(&createdDraft); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createdDraft, nil
}
