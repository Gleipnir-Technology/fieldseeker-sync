package labelstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TasksListResponse represents the response from the /api/tasks endpoint
type TasksListResponse struct {
	Tasks            []Task `json:"tasks"`
	Total            int    `json:"total"`
	TotalAnnotations int    `json:"total_annotations"`
	TotalPredictions int    `json:"total_predictions"`
}

// Task represents a single task returned by the Label Studio API
type Task struct {
	Agreement                string                   `json:"agreement"`
	AgreementSelected        string                   `json:"agreement_selected"`
	Annotations              json.RawMessage          `json:"annotations"`
	AnnotationsIDs           json.RawMessage          `json:"annotations_ids"`
	AnnotationsResults       json.RawMessage          `json:"annotations_results"`
	Annotators               []int                    `json:"annotators"`
	AnnotatorsCount          int                      `json:"annotators_count"`
	AvgLeadTime              float64                  `json:"avg_lead_time"`
	CancelledAnnotations     int                      `json:"cancelled_annotations"`
	CommentAuthors           []map[string]interface{} `json:"comment_authors"`
	CommentAuthorsCount      int                      `json:"comment_authors_count"`
	CommentCount             int                      `json:"comment_count"`
	Comments                 json.RawMessage          `json:"comments"`
	CompletedAt              string                   `json:"completed_at"`
	CreatedAt                time.Time                `json:"created_at"`
	Data                     map[string]interface{}   `json:"data"`
	DraftExists              bool                     `json:"draft_exists"`
	Drafts                   []json.RawMessage        `json:"drafts"`
	FileUpload               string                   `json:"file_upload"`
	GroundTruth              bool                     `json:"ground_truth"`
	ID                       int                      `json:"id"`
	InnerID                  int                      `json:"inner_id"`
	IsLabeled                bool                     `json:"is_labeled"`
	LastCommentUpdatedAt     string                   `json:"last_comment_updated_at"`
	Meta                     map[string]interface{}   `json:"meta"`
	Overlap                  int                      `json:"overlap"`
	Predictions              []json.RawMessage        `json:"predictions"`
	PredictionsModelVersions json.RawMessage          `json:"predictions_model_versions"`
	PredictionsResults       json.RawMessage          `json:"predictions_results"`
	PredictionsScore         float64                  `json:"predictions_score"`
	Project                  int                      `json:"project"`
	ReviewTime               int                      `json:"review_time"`
	Reviewed                 bool                     `json:"reviewed"`
	Reviewers                []map[string]interface{} `json:"reviewers"`
	ReviewersCount           int                      `json:"reviewers_count"`
	ReviewsAccepted          int                      `json:"reviews_accepted"`
	ReviewsRejected          int                      `json:"reviews_rejected"`
	StorageFilename          string                   `json:"storage_filename"`
	TotalAnnotations         int                      `json:"total_annotations"`
	TotalPredictions         int                      `json:"total_predictions"`
	UnresolvedCommentCount   int                      `json:"unresolved_comment_count"`
	UpdatedAt                time.Time                `json:"updated_at"`
	UpdatedBy                []map[string]interface{} `json:"updated_by"`
}

// TasksListOptions represents query parameters that can be used to filter tasks
type TasksListOptions struct {
	ProjectID   int    // Filter by project ID
	Page        int    // Page number for pagination
	PageSize    int    // Number of items per page
	Ordering    string // Field to order by (e.g., "created_at", "-created_at" for descending)
	FilterQuery string // Search query for filtering tasks
	IsLabeled   *bool  // Filter by labeled status
	IsReviewed  *bool  // Filter by review status
	GroundTruth *bool  // Filter by ground truth status
}

// ListTasks fetches the list of tasks from the Label Studio API
func (c *Client) ListTasks(options *TasksListOptions) (*TasksListResponse, error) {
	// Check if we have an access token, if not try to get it
	if c.AccessToken == "" {
		if err := c.GetAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	// Build URL with query parameters
	endpoint := fmt.Sprintf("%s/api/tasks/", c.BaseURL)
	if options != nil {
		queryParams := url.Values{}

		// Add all the possible filter parameters
		if options.ProjectID > 0 {
			queryParams.Add("project", fmt.Sprintf("%d", options.ProjectID))
		}
		if options.Page > 0 {
			queryParams.Add("page", fmt.Sprintf("%d", options.Page))
		}
		if options.PageSize > 0 {
			queryParams.Add("page_size", fmt.Sprintf("%d", options.PageSize))
		}
		if options.Ordering != "" {
			queryParams.Add("ordering", options.Ordering)
		}
		if options.FilterQuery != "" {
			queryParams.Add("filter", options.FilterQuery)
		}
		if options.IsLabeled != nil {
			if *options.IsLabeled {
				queryParams.Add("is_labeled", "true")
			} else {
				queryParams.Add("is_labeled", "false")
			}
		}
		if options.IsReviewed != nil {
			if *options.IsReviewed {
				queryParams.Add("reviewed", "true")
			} else {
				queryParams.Add("reviewed", "false")
			}
		}
		if options.GroundTruth != nil {
			if *options.GroundTruth {
				queryParams.Add("ground_truth", "true")
			} else {
				queryParams.Add("ground_truth", "false")
			}
		}

		// Add query params to URL if we have any
		if len(queryParams) > 0 {
			endpoint = fmt.Sprintf("%s?%s", endpoint, queryParams.Encode())
		}
	}

	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
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
		return nil, fmt.Errorf("API returned error: %s: ", resp.Status, resp.Body)
	}

	// Parse response
	var tasksResponse TasksListResponse
	if err := json.NewDecoder(resp.Body).Decode(&tasksResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tasksResponse, nil
}
