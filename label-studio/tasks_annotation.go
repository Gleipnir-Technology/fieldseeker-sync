package labelstudio

import (
	"encoding/json"
	"fmt"
	"time"
)

// AnnotationRequest represents the request body for creating a draft
type AnnotationRequest struct {
	DraftID          int          `json:"draft_id"`
	LeadTime         float64      `json:"lead_time"`
	ParentAnnotation *int         `json:"parent_annotation,omitempty"`
	ParentPrediction *int         `json:"parent_prediction,omitempty"`
	Project          int          `json:"project"`
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
	UpdatedBy        int          `json:"updated_by"`
}

// NewAnnotation creates a new draft request builder
func NewAnnotationRequest(projectID int) *AnnotationRequest {
	return &AnnotationRequest{
		Project:   projectID,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// CreateAnnotation creates a new annotation on a task
func (c *Client) CreateAnnotation(taskID int, annotation *AnnotationRequest) (*Annotation, error) {
	// Marshal the annotation request to JSON
	annotationJSON, err := json.Marshal(annotation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal annotation request: %w", err)
	}

	// Create request URL with query parameter
	path := fmt.Sprintf("/api/tasks/%d/annotations", taskID)
	resp, err := c.makeRequest("POST", path, annotationJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var createdAnnotation Annotation
	if err := json.NewDecoder(resp.Body).Decode(&createdAnnotation); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createdAnnotation, nil
}
