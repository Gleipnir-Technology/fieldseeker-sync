package main

import (
	//"flag"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/label-studio"
	"github.com/Gleipnir-Technology/fieldseeker-sync/minio"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

func createMinioClient() (*minio.Client, error) {
	baseUrl := os.Getenv("S3_BASE_URL")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")

	client, err := minio.NewClient(baseUrl, accessKeyID, secretAccessKey)
	if err != nil {
		return nil, err
	}
	return client, err
}

func main() {
	err := fssync.InitDB()
	if err != nil {
		log.Println("Failed to initialize: ", err)
		os.Exit(1)
	}
	log.Println("Initialized database connection")

	// Initialize the minio client
	minioBucket := os.Getenv("S3_BUCKET")
	minioClient, err := createMinioClient()
	if err != nil {
		log.Printf("Failed to initialize minio: %v", err)
		os.Exit(2)
	}
	log.Println("Created minio client")

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

	customer := os.Getenv("CUSTOMER")
	if customer == "" {
		log.Fatalf("You must specify a CUSTOMER env var")
	}
	// Get all the note audios
	allNoteAudio, err := database.NoteAudioQuery()
	for _, note := range allNoteAudio {
		task, err := findMatchingTask(labelStudioClient, project, customer, note)
		if err != nil {
			log.Fatalf("Failed to search for a task: %v", err)
		}
		// We already have a task, nothing to do.
		if task != nil {
			continue
		}
		err = createTask(labelStudioClient, project, minioClient, minioBucket, customer, note)
		if err != nil {
			log.Printf("Failed to create a task: %v", err)
			continue
		}
	}
}

func createTask(client *labelstudio.Client, project *labelstudio.Project, minioClient *minio.Client, bucket string, customer string, note *shared.NoteAudio) error {
	config := fssync.ReadConfig()
	audioRef := fmt.Sprintf("s3://%s/%s-normalized.m4a", bucket, note.UUID)
	audioFile := fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, note.UUID)
	uploadPath := fmt.Sprintf("%s-normalized.m4a", note.UUID)
	err := minioClient.UploadFile(bucket, audioFile, uploadPath)
	if err != nil {
		return fmt.Errorf("Failed to upload audio: %v", err)
	}
	transcription := ""
	if note.Transcription != nil {
		transcription = *note.Transcription
	}
	simpleTasks := []map[string]interface{}{
		{
			"data": map[string]string{
				"audio":         audioRef,
				"note_uuid":     note.UUID,
				"transcription": transcription,
			},
			"meta": map[string]string{
				"customer":  customer,
				"note_uuid": note.UUID,
			},
		},
	}
	_, err = client.ImportTasks(project.ID, simpleTasks)
	if err != nil {
		log.Fatalf("Failed to import tasks: %v", err)
	}
	log.Printf("Created task for note audio %s", note.UUID)
	return nil
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

func findMatchingTask(client *labelstudio.Client, project *labelstudio.Project, customer string, note *shared.NoteAudio) (*labelstudio.Task, error) {
	/*meta := map[string]string{
		"customer": customer,
		"note_uuid": note.UUID,
	}*/
	items := []map[string]interface{}{
		{"filter": "filter:tasks:data.note_uuid", "operator": "equal", "type": "string", "value": note.UUID},
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
	// Specify bucket name
	//bucketNamePtr := flag.String("bucket", "label-studio", "The bucket to upload to")
	//filePathPtr := flag.String("file", "example.txt", "The file to upload")
	//flag.Parse()
}
