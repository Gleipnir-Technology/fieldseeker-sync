package labelstudio

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProjectsResponse represents the response from the /api/projects endpoint
type ProjectsResponse struct {
	Count    int       `json:"count"`
	Results  []Project `json:"results"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
}

// Project represents a single project returned by the Label Studio API
type Project struct {
	AllowStream                     bool               `json:"allow_stream"`
	AssignmentSettings              AssignmentSettings `json:"assignment_settings"`
	Blueprints                      []Blueprint        `json:"blueprints"`
	ConfigHasControlTags            bool               `json:"config_has_control_tags"`
	ConfigSuitableForBulkAnnotation bool               `json:"config_suitable_for_bulk_annotation"`
	CreatedAt                       time.Time          `json:"created_at"`
	DataTypes                       map[string]string  `json:"data_types"`
	DescriptionShort                string             `json:"description_short"`
	FinishedTaskNumber              int                `json:"finished_task_number"`
	GroundTruthNumber               int                `json:"ground_truth_number"`
	ID                              int                `json:"id"`
	Members                         string             `json:"members"`
	MembersCount                    int                `json:"members_count"`
	NumTasksWithAnnotations         int                `json:"num_tasks_with_annotations"`
	//ParsedLabelConfig               map[string]string    `json:"parsed_label_config"`
	Prompts   string `json:"prompts"`
	QueueDone int    `json:"queue_done"`
	QueueLeft int    `json:"queue_left"`
	//QueueTotal                      string               `json:"queue_total"`
	Ready              bool           `json:"ready"`
	Rejected           int            `json:"rejected"`
	ReviewSettings     ReviewSettings `json:"review_settings"`
	ReviewTotalTasks   int            `json:"review_total_tasks"`
	ReviewedNumber     int            `json:"reviewed_number"`
	ReviewerQueueTotal int            `json:"reviewer_queue_total"`
	//SkippedAnnotationsNumber        string               `json:"skipped_annotations_number"`
	StartTrainingOnAnnotationUpdate bool `json:"start_training_on_annotation_update"`
	TaskNumber                      int  `json:"task_number"`
	//TotalAnnotationsNumber          string               `json:"total_annotations_number"`
	TotalPredictionsNumber          int    `json:"total_predictions_number"`
	Workspace                       string `json:"workspace"`
	WorkspaceTitle                  string `json:"workspace_title"`
	AnnotationLimitCount            int    `json:"annotation_limit_count"`
	AnnotationLimitPercent          string `json:"annotation_limit_percent"`
	AnnotatorEvaluationMinimumScore string `json:"annotator_evaluation_minimum_score"`
	AnnotatorEvaluationMinimumTasks int    `json:"annotator_evaluation_minimum_tasks"`
	Color                           string `json:"color"`
	CommentClassificationConfig     string `json:"comment_classification_config"`
	//ControlWeights                  map[string]string    `json:"control_weights"`
	CreatedBy                         User   `json:"created_by"`
	CustomScript                      string `json:"custom_script"`
	CustomTaskLockTtl                 int    `json:"custom_task_lock_ttl"`
	Description                       string `json:"description"`
	DuplicationDone                   bool   `json:"duplication_done"`
	DuplicationStatus                 string `json:"duplication_status"`
	EnableEmptyAnnotation             bool   `json:"enable_empty_annotation"`
	EvaluatePredictionsAutomatically  bool   `json:"evaluate_predictions_automatically"`
	ExpertInstruction                 string `json:"expert_instruction"`
	IsDraft                           bool   `json:"is_draft"`
	IsPublished                       bool   `json:"is_published"`
	LabelConfig                       string `json:"label_config"`
	MaximumAnnotations                int    `json:"maximum_annotations"`
	MinAnnotationsToStartTraining     int    `json:"min_annotations_to_start_training"`
	ModelVersion                      string `json:"model_version"`
	Organization                      int    `json:"organization"`
	OverlapCohortPercentage           int    `json:"overlap_cohort_percentage"`
	PauseOnFailedAnnotatorEvaluation  bool   `json:"pause_on_failed_annotator_evaluation"`
	PinnedAt                          string `json:"pinned_at"`
	RequireCommentOnSkip              bool   `json:"require_comment_on_skip"`
	RevealPreannotationsInteractively bool   `json:"reveal_preannotations_interactively"`
	Sampling                          string `json:"sampling"`
	ShowAnnotationHistory             bool   `json:"show_annotation_history"`
	ShowCollabPredictions             bool   `json:"show_collab_predictions"`
	ShowGroundTruthFirst              bool   `json:"show_ground_truth_first"`
	ShowInstruction                   bool   `json:"show_instruction"`
	ShowOverlapFirst                  bool   `json:"show_overlap_first"`
	ShowSkipButton                    bool   `json:"show_skip_button"`
	ShowUnusedDataColumnsToAnnotators bool   `json:"show_unused_data_columns_to_annotators"`
	SkipQueue                         string `json:"skip_queue"`
	Title                             string `json:"title"`
	UsefulAnnotationNumber            int    `json:"useful_annotation_number"`
}

// Blueprint represents a blueprint in a project
type Blueprint struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int       `json:"id"`
	ShareID   string    `json:"share_id"`
	ShortURL  string    `json:"short_url"`
}

// AssignmentSettings represents the assignment settings of a project
type AssignmentSettings struct {
	ID int `json:"id"`
}

// ReviewSettings represents the review settings of a project
type ReviewSettings struct {
	ID                              int  `json:"id"`
	RequeueRejectedTasksToAnnotator bool `json:"requeue_rejected_tasks_to_annotator"`
}

// Projects fetches the list of projects from the Label Studio API
func (c *Client) Projects() (*ProjectsResponse, error) {
	resp, err := c.makeRequest("GET", "/api/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to GET /api/projects: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var projects ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &projects, nil
}
