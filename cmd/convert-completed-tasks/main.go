package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database/sql"
	"github.com/Gleipnir-Technology/fieldseeker-sync/label-studio"
)

func main() {
	count := 0
	flag.IntVar(&count, "count", 0, "Set the max number of tasks")
	username := flag.String("username", "", "The username of the user to convert")
	flag.Parse()
	if username == nil || *username == "" {
		log.Println("You must specify -username")
		os.Exit(2)
	}

	customer := os.Getenv("CUSTOMER")
	if customer == "" {
		log.Fatalf("You must specify a CUSTOMER env var")
	}

	err := fssync.InitDB()
	if err != nil {
		log.Println("Failed to initialize: ", err)
		os.Exit(1)
	}
	log.Println("Initialized database connection")

	// Initialize the client with your Label Studio base URL and API key
	labelStudioApiKey := os.Getenv("LABEL_STUDIO_API_KEY")
	labelStudioBaseUrl := os.Getenv("LABEL_STUDIO_BASE_URL")
	labelStudioClient := labelstudio.NewClient(labelStudioBaseUrl, labelStudioApiKey)
	log.Println("Created label studio client")

	// Get and store the access token
	err = labelStudioClient.GetAccessToken()
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}
	log.Println("Got label studio client access token")

	// Get the project we are going to upload to
	project, err := findLabelStudioProject(labelStudioClient, "Nidus Speech-to-Text Transcriptions")
	if err != nil {
		log.Fatalf("Failed to find the label studio project")
	}
	log.Printf("Using project %d", project.ID)

	// Get the users because we have to reference the user IDs later
	labelStudioUser, err := findLabelStudioUser(labelStudioClient, *username)
	if err != nil {
		log.Fatalf("Failed to find label studio user: %v", err)
	}
	if labelStudioUser == nil {
		log.Fatalf("Couldn't find a matching Label Studio user for '%s'", username)
	}

	log.Printf("Using customer %s", customer)

	// Get all the completed tasks for my user
	ctx := context.Background()
	log.Printf("Searching for completed tasks for '%s'", *username)
	completedTasks, err := sql.TaskAudioReviewCompletedBy(*username).All(ctx, database.PGInstance.BobDB)
	log.Printf("Found %d completed tasks", len(completedTasks))
	for i, reviewTask := range completedTasks {
		if count != 0 && i >= count {
			log.Printf("Stopping after %d tasks", count)
			return
		}
		if reviewTask.Transcription.IsNull() {
			log.Printf("Review task %d has no transcription", reviewTask.TaskID)
			continue
		}
		transcription := reviewTask.Transcription.MustGet()
		// Find the task for the given note_audio
		labelTask, err := findMatchingTask(labelStudioClient, project, customer, reviewTask)
		if err != nil {
			log.Fatalf("Failed to search for a task: %v", err)
		}
		// If there's no such task, there's not much we can do.
		if labelTask == nil {
			log.Printf("Cannot find a task for note %s, moving on", reviewTask.NoteAudioUUID)
			continue
		}
		// If it's already correctly updated, move on
		if labelTask.Data["transcription"] == transcription {
			log.Printf("Already updated the transcription for label studio task %d note_audio %s to '%s'", labelTask.ID, reviewTask.NoteAudioUUID, transcription)
			continue
		}
		// Otherwise, update the task with the updated information
		//update := labelstudio.NewTaskUpdate().SetData(map[string]interface{}{
		//"note_uuid": reviewTask.NoteAudioUUID,
		//"transcription": reviewTask.Transcription,
		//}).SetReviewed(true)
		//_, err = labelStudioClient.TaskUpdate(labelTask.ID, update)
		resultID := randomID()
		taskResultValue := labelstudio.TaskResultValue{}
		taskResultValue.Text = []string{transcription}
		taskResult := labelstudio.TaskResult{}
		taskResult.FromName = "transcription"
		taskResult.ID = resultID
		taskResult.Origin = "manual"
		taskResult.ToName = "audio"
		taskResult.Type = "textarea"
		taskResult.Value = taskResultValue
		draftRequest := labelstudio.NewDraft(project.ID)
		draftRequest.CreatedUsername = fmt.Sprintf(" %s, %d", labelStudioUser.Email, labelStudioUser.ID)
		draftRequest.CreatedAgo = "2 minutes"
		draftRequest.ImportID = nil
		// Fake value
		draftRequest.Task = labelTask.ID
		draftRequest.User = "janie@gleipnir.technology"

		draftRequest.DraftID = 0
		draftRequest.LeadTime = 20.123
		draftRequest.ParentAnnotation = nil
		draftRequest.ParentPrediction = nil
		draftRequest.Project = string(project.ID)
		draftRequest.Result = []labelstudio.TaskResult{taskResult}
		draftRequest.StartedAt = time.Now().Format("2006-01-02T15:04:05.000Z")
		draft, err := labelStudioClient.CreateDraft(labelTask.ID, draftRequest)

		if err != nil {
			log.Fatalf("Failed to create draft: %v", err)
		}
		log.Printf("Created draft %d", draft.ID)

		annotationRequest := labelstudio.NewAnnotationRequest(project.ID)
		annotationRequest.DraftID = draft.ID
		annotationRequest.LeadTime = 20.123
		annotationRequest.ParentAnnotation = nil
		annotationRequest.ParentPrediction = nil
		annotationRequest.Result = []labelstudio.TaskResult{taskResult}
		annotation, err := labelStudioClient.CreateAnnotation(labelTask.ID, annotationRequest)
		if err != nil {
			log.Fatalf("Failed to create annotation: %v", err)
		}
		log.Printf("Created annotation %d", annotation.ID)
		log.Printf("Finished port of review task %d of note_audio %s", reviewTask.TaskID, reviewTask.NoteAudioUUID)
	}
}

func findLabelStudioProject(client *labelstudio.Client, title string) (*labelstudio.Project, error) {
	// Attempt to get live projects
	projects, err := client.Projects()
	if err != nil {
		log.Fatalf("Failed to get projects: %v", err)
	}
	fmt.Printf("Found %d projects:\n", projects.Count)
	for i, p := range projects.Results {
		fmt.Printf("%d. %s (ID: %d) - Tasks: %d\n",
			i+1,
			p.Title,
			p.ID,
			p.TaskNumber)
		if p.Title == title {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("No such project '%s'", title)
}

func findMatchingTask(client *labelstudio.Client, project *labelstudio.Project, customer string, task sql.TaskAudioReviewCompletedByRow) (*labelstudio.Task, error) {
	/*meta := map[string]string{
		"customer": customer,
		"note_uuid": note.UUID,
	}*/
	items := []map[string]interface{}{
		{"filter": "filter:tasks:data.note_uuid", "operator": "equal", "type": "string", "value": task.NoteAudioUUID},
	}
	filters := map[string]interface{}{
		"conjunction": "and",
		"items":       items,
	}
	query := map[string]interface{}{
		"filters": filters,
	}
	queryStr, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal query JSON: %v", err)
	}
	// Get all tasks
	options := &labelstudio.TasksListOptions{
		ProjectID: project.ID,
		Query:     string(queryStr),
	}
	tasksResponse, err := client.ListTasks(options)
	if err != nil {
		return nil, fmt.Errorf("Failed to get tasks: %v", err)
	}
	if len(tasksResponse.Tasks) == 0 {
		return nil, nil
	} else if len(tasksResponse.Tasks) == 1 {
		return &tasksResponse.Tasks[0], nil
	} else {
		return nil, fmt.Errorf("Got too many tasks: %d", len(tasksResponse.Tasks))
	}
}

func findLabelStudioUser(labelStudioClient *labelstudio.Client, username string) (*labelstudio.User, error) {
	users, err := labelStudioClient.ListUsers()
	if err != nil {
		return nil, err
	}
	emailMap := map[string]string{
		"benjaminsperry":  "ben@gleipnir.technology",
		"eliribble":       "eli@gleipnir.technology",
		"janiesperry":     "janie@gleipnir.technology",
		"tzipporahribble": "tzipporah@gleipnir.technology",
	}
	email := emailMap[username]
	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, errors.New("No such user")
}

func randomID() string {
	// Define the character set
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create a byte slice to store the result
	result := make([]byte, 10)

	// Fill the result with random characters
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
