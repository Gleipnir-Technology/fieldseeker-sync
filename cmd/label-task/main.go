package main

import (
	//"flag"
	"fmt"
	"log"
	"os"

	"github.com/Gleipnir-Technology/fieldseeker-sync/label-studio"
	"github.com/Gleipnir-Technology/fieldseeker-sync/minio"
)

func createMinioClient() *minio.Client {
	baseUrl := os.Getenv("S3_BASE_URL")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")

	client := minio.NewClient(baseUrl, accessKeyID, secretAccessKey)
	return client
}

func main() {

	// Initialize the client with your Label Studio base URL and API key
	labelStudioApiKey := os.Getenv("LABEL_STUDIO_API_KEY")
	labelStudioBaseUrl := os.Getenv("LABEL_STUDIO_BASE_URL")
	client := labelstudio.NewClient(labelStudioBaseUrl, labelStudioApiKey)

	// Get and store the access token
	err := client.GetAccessToken()
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	// Attempt to get live projects
	projects, err := client.Projects()
	if err != nil {
		log.Fatalf("Failed to get projects: %v", err)
	}
	fmt.Printf("Found %d projects:\n", projects.Count)
	var project labelstudio.Project
	for i, p := range projects.Results {
		fmt.Printf("%d. %s (ID: %d) - Tasks: %d\n",
			i+1,
			p.Title,
			p.ID,
			p.TaskNumber)
		project = p
	}

	simpleTasks := []map[string]interface{}{
		{
			"data": map[string]string{
				"audio":         "s3://label-studio-nidus-audio/ffda05fd-a999-4a1d-b043-0089d3241280-normalized.m4a",
				"transcription": "This is a fake transcription I just wrote.",
			},
		},
	}
	response, err := client.ImportTasks(project.ID, simpleTasks)
	if err != nil {
		log.Fatalf("Failed to import tasks: %v", err)
	}
	fmt.Printf("Successfully imported %d tasks\n", response.TaskCount)

	// Specify bucket name
	//bucketNamePtr := flag.String("bucket", "label-studio", "The bucket to upload to")
	//filePathPtr := flag.String("file", "example.txt", "The file to upload")
	//flag.Parse()

}
