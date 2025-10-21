package fssync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	models "github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
	"github.com/Gleipnir-Technology/fieldseeker-sync/label-studio"
	"github.com/Gleipnir-Technology/fieldseeker-sync/minio"
	"github.com/google/uuid"
)

type LabelStudioJob struct {
	UUID uuid.UUID
}

var labelJobChannel chan LabelStudioJob

func EnqueueLabelStudioJob(job LabelStudioJob) {
	select {
	case labelJobChannel <- job:
		log.Printf("Enqueued label job for UUID: %s", job.UUID)
	default:
		log.Printf("Label job channel is full, dropping job for UUID: %s", job.UUID)
	}
}

func StartLabelStudioWorker(ctx context.Context) error {
	// Initialize the minio client
	minioBucket := os.Getenv("S3_BUCKET")

	labelStudioClient, err := createLabelStudioClient()
	if err != nil {
		return fmt.Errorf("Failed to create label studio client: %v", err)
	}
	minioClient, err := createMinioClient()
	if err != nil {
		return fmt.Errorf("Failed to create minio client: %v", err)
	}
	labelJobChannel = make(chan LabelStudioJob, 100) // Buffered channel to prevent blocking
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Audio worker shutting down.")
				return
			case job := <-labelJobChannel:
				log.Printf("Processing audio job for UUID: %s", job.UUID)
				err := processLabelTask(ctx, minioClient, minioBucket, labelStudioClient, job)
				if err != nil {
					log.Printf("Error processing audio file %s: %v", job.UUID, err)
				}
			}
		}
	}()
	return nil
}

func createMinioClient() (*minio.Client, error) {
	baseUrl := os.Getenv("S3_BASE_URL")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")

	client, err := minio.NewClient(baseUrl, accessKeyID, secretAccessKey)
	if err != nil {
		return nil, err
	}
	log.Println("Created minio client")
	return client, err
}

func createLabelStudioClient() (*labelstudio.Client, error) {
	// Initialize the client with your Label Studio base URL and API key
	labelStudioApiKey := os.Getenv("LABEL_STUDIO_API_KEY")
	labelStudioBaseUrl := os.Getenv("LABEL_STUDIO_BASE_URL")
	labelStudioClient := labelstudio.NewClient(labelStudioBaseUrl, labelStudioApiKey)
	log.Println("Created label studio client")

	// Get and store the access token
	err := labelStudioClient.GetAccessToken()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get access token: %v", err))
	}
	log.Println("Got label studio client access token")

	return labelStudioClient, nil
}

func processLabelTask(ctx context.Context, minioClient *minio.Client, minioBucket string, labelStudioClient *labelstudio.Client, job LabelStudioJob) error {
	customer := os.Getenv("CUSTOMER")
	if customer == "" {
		return errors.New("You must specify a CUSTOMER env var")
	}
	// Get the project we are going to upload to
	project, err := findLabelStudioProject(labelStudioClient, "Nidus Speech-to-Text Transcriptions")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to find the label studio project"))
	}
	note, err := database.NoteAudioGetLatest(ctx, job.UUID.String())
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get note %s", note.UUID))
	}

	if note.Version != 1 {
		return errors.New(fmt.Sprintf("Got version %d of %s", note.Version, note.UUID))
	}
	task, err := findMatchingTask(labelStudioClient, project, customer, note)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to search for a task: %v", err))
	}
	// We already have a task, nothing to do.
	if task != nil {
		return nil
	}

	err = createTask(labelStudioClient, project, minioClient, minioBucket, customer, note)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create a task: %v", err))
	}
	return nil
}

func createTask(client *labelstudio.Client, project *labelstudio.Project, minioClient *minio.Client, bucket string, customer string, note *models.NoteAudio) error {
	config := ReadConfig()
	audioRef := fmt.Sprintf("s3://%s/%s-normalized.m4a", bucket, note.UUID)
	audioFile := fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, note.UUID)
	uploadPath := fmt.Sprintf("%s-normalized.m4a", note.UUID)

	if !minioClient.ObjectExists(bucket, uploadPath) {
		err := minioClient.UploadFile(bucket, audioFile, uploadPath)
		if err != nil {
			return fmt.Errorf("Failed to upload audio: %v", err)
		}
	}
	var transcription string = ""
	if note.Transcription.IsValue() {
		transcription = note.Transcription.MustGet()
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
	_, err := client.ImportTasks(project.ID, simpleTasks)
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

func findMatchingTask(client *labelstudio.Client, project *labelstudio.Project, customer string, note *models.NoteAudio) (*labelstudio.Task, error) {
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
