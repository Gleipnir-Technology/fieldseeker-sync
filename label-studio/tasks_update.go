package labelstudio

import (
	"encoding/json"
	"fmt"
)

type TaskResultValue struct {
	Text []string `json:"text"`
}

type TaskResult struct {
	ID       string          `json:"id"`
	FromName string          `json:"from_name"`
	Origin   string          `json:"origin"`
	ToName   string          `json:"to_name"`
	Type     string          `json:"type"`
	Value    TaskResultValue `json:"value"`
}

// TaskUpdate defines fields that can be updated in a task
type TaskUpdate struct {
	// Fields that can be updated
	Annotations json.RawMessage         `json:"annotations,omitempty"`
	Data        *map[string]interface{} `json:"data,omitempty"`
	DraftExists *bool                   `json:"draft_exists,omitempty"`
	Drafts      json.RawMessage         `json:"drafts,omitempty"`
	GroundTruth *bool                   `json:"ground_truth,omitempty"`
	IsLabeled   *bool                   `json:"is_labeled,omitempty"`
	Meta        *map[string]interface{} `json:"meta,omitempty"`
	Predictions json.RawMessage         `json:"predictions,omitempty"`
	Reviewed    *bool                   `json:"reviewed"`

	// Internal tracking
	fieldsToUpdate map[string]bool
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
	// Marshal the updates to JSON
	updateJSON, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updates: %w", err)
	}

	// Create request
	path := fmt.Sprintf("/api/tasks/%d", taskID)
	resp, err := c.makeRequest("PATCH", path, updateJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to PATCH %s: %w", path, err)
	}
	defer resp.Body.Close()

	// Parse response
	var updatedTask Task
	if err := json.NewDecoder(resp.Body).Decode(&updatedTask); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &updatedTask, nil
}
