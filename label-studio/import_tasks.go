package labelstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TaskImportResponse represents the response from the import tasks endpoint
type TaskImportResponse struct {
	// Common fields that might be returned
	TaskCount  int                    `json:"task_count,omitempty"`
	Annotation map[string]interface{} `json:"annotation,omitempty"`
	Task       map[string]interface{} `json:"task,omitempty"`

	// For handling any other fields in the response
	AdditionalProperties map[string]interface{} `json:"-"`
}

// UnmarshalJSON custom unmarshaler for TaskImportResponse to capture all fields
func (r *TaskImportResponse) UnmarshalJSON(data []byte) error {
	// First unmarshal the known fields
	type Alias TaskImportResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Then capture any additional fields
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	r.AdditionalProperties = make(map[string]interface{})
	for k, v := range rawMap {
		// Skip fields we already processed
		if k != "task_count" && k != "annotation" && k != "task" {
			r.AdditionalProperties[k] = v
		}
	}

	return nil
}

// ImportTasks imports tasks into a Label Studio project
// tasks parameter can be any data structure that can be marshalled to JSON
func (c *Client) ImportTasks(projectID int, tasks interface{}) (*TaskImportResponse, error) {
	// Marshal the tasks to JSON
	taskJSON, err := json.Marshal(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	path := fmt.Sprintf("/api/projects/%d/import", projectID)
	resp, err := c.makeRequest("POST", path, taskJSON)
	if err != nil {
		return nil, fmt.Errorf("Failed to POST %s: %v", path, err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
	}

	// Parse response
	var response TaskImportResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
