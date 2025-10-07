package labelstudio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// TaskUpdate defines fields that can be updated in a task
type TaskUpdate struct {
	// Fields that can be updated
	Annotations             json.RawMessage           `json:"annotations,omitempty"`
	Data                    *map[string]interface{}   `json:"data,omitempty"`
	DraftExists             *bool                     `json:"draft_exists,omitempty"`
	Drafts                  json.RawMessage           `json:"drafts,omitempty"`
	GroundTruth             *bool                     `json:"ground_truth,omitempty"`
	IsLabeled               *bool                     `json:"is_labeled,omitempty"`
	Meta                    *map[string]interface{}   `json:"meta,omitempty"`
	Predictions             json.RawMessage           `json:"predictions,omitempty"`
	Reviewed                *bool                     `json:"reviewed"`

	// Internal tracking
	fieldsToUpdate          map[string]bool
}

// NewTaskUpdate creates a new TaskUpdate builder
func NewTaskUpdate() *TaskUpdate {
	return &TaskUpdate{
		fieldsToUpdate: make(map[string]bool),
	}
}

func (t *TaskUpdate) MarshalJSON() ([]byte, error) {
	// Only include fields that are explicitly set
	updateMap := make(map[string]interface{})
	
	if t.fieldsToUpdate["annotations"] {
		// Parse raw JSON back to interface{} to include in the map
		var annotations interface{}
		if err := json.Unmarshal(t.Annotations, &annotations); err != nil {
			return nil, err
		}
		updateMap["annotations"] = annotations
	}
	if t.fieldsToUpdate["data"] {
		updateMap["data"] = t.Data
	}
	if t.fieldsToUpdate["draft_exists"] {
		updateMap["draft_exists"] = t.DraftExists
	}
	if t.fieldsToUpdate["drafts"] {
		var drafts interface{}
		if err := json.Unmarshal(t.Drafts, &drafts); err != nil {
			return nil, err
		}
		updateMap["drafts"] = drafts
	}
	if t.fieldsToUpdate["ground_truth"] {
		updateMap["ground_truth"] = t.GroundTruth
	}
	if t.fieldsToUpdate["is_labeled"] {
		updateMap["is_labeled"] = t.IsLabeled
	}
	if t.fieldsToUpdate["meta"] {
		updateMap["meta"] = t.Meta
	}
	if t.fieldsToUpdate["predictions"] {
		var predictions interface{}
		if err := json.Unmarshal(t.Predictions, &predictions); err != nil {
			return nil, err
		}
		updateMap["predictions"] = predictions
	}
	if t.fieldsToUpdate["reviewed"] {
		updateMap["reviewed"] = t.Reviewed
	}
	
	return json.Marshal(updateMap)
}

func (t *TaskUpdate) SetAnnotations(annotations interface{}) *TaskUpdate {
	annotationsJSON, err := json.Marshal(annotations)
	if err != nil {
		// Handle error gracefully in a builder pattern
		// Could store the error and check it later
		return t
	}
	t.Annotations = annotationsJSON
	t.fieldsToUpdate["annotations"] = true
	return t
}

func (t *TaskUpdate) SetData(data map[string]interface{}) *TaskUpdate {
	t.Data = &data
	t.fieldsToUpdate["data"] = true
	return t
}

func (t *TaskUpdate) SetDraftExists(draftExists bool) *TaskUpdate {
	t.DraftExists = &draftExists
	t.fieldsToUpdate["draft_exists"] = true
	return t
}

func (t *TaskUpdate) SetDrafts(drafts interface{}) *TaskUpdate {
	draftsJSON, err := json.Marshal(drafts)
	if err != nil {
		return t
	}
	t.Drafts = draftsJSON
	t.fieldsToUpdate["drafts"] = true
	return t
}

func (t *TaskUpdate) SetGroundTruth(groundTruth bool) *TaskUpdate {
	t.GroundTruth = &groundTruth
	t.fieldsToUpdate["ground_truth"] = true
	return t
}

func (t *TaskUpdate) SetIsLabeled(isLabeled bool) *TaskUpdate {
	t.IsLabeled = &isLabeled
	t.fieldsToUpdate["is_labeled"] = true
	return t
}

func (t *TaskUpdate) SetMeta(meta map[string]interface{}) *TaskUpdate {
	t.Meta = &meta
	t.fieldsToUpdate["meta"] = true
	return t
}


func (t *TaskUpdate) SetPredictions(predictions interface{}) *TaskUpdate {
	predictionsJSON, err := json.Marshal(predictions)
	if err != nil {
		return t
	}
	t.Predictions = predictionsJSON
	t.fieldsToUpdate["predictions"] = true
	return t
}

func (t *TaskUpdate) SetReviewed(isReviewed bool) *TaskUpdate {
	t.Reviewed = &isReviewed
	t.fieldsToUpdate["reviewed"] = true
	return t
}

func (c *Client) TaskUpdate(taskID int, update *TaskUpdate) (*Task, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}
	
	// Marshal the updates to JSON
	updateJSON, err := json.Marshal(update)
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
