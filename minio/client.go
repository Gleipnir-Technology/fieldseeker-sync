package minio

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
}

func NewClient(baseURL string, accessKeyID string, secretAccessKey string) (*Client, error) {
	log.Printf("Connecting to S3 at %s", baseURL)
	minioClient, err := minio.New(baseURL, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to minio: %v", err)
	}
	return &Client{
		client: minioClient,
	}, nil
}

func signUrl(minioClient *minio.Client, bucketName string, filePath string) {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+filePath+"\"")

	// Generates a presigned url which expires in a day.
	presignedURL, err := minioClient.PresignedGetObject(context.Background(), bucketName, filePath, time.Second*24*60*60, reqParams)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully generated presigned URL", presignedURL)
}
func (minioClient *Client) ObjectExists(bucket string, path string) bool {
	ctx := context.Background()
	opts := minio.ListObjectsOptions{
		UseV1:     false,
		Prefix:    path,
		Recursive: false,
	}
	for object := range minioClient.client.ListObjects(ctx, bucket, opts) {
		if object.Err == nil {
			return true
		}
		log.Printf("Error getting object %s/%s: %v", bucket, path, object.Err)
	}
	return false
}

func (minioClient *Client) UploadFile(bucketName string, filePath string, uploadPath string) error {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file %s to upload: %v", filePath, err)
	}
	defer file.Close()

	// Upload the file
	_, err = minioClient.client.FPutObject(context.Background(), bucketName, uploadPath, filePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("Failed to put object to bucket %s: %v", bucketName, err)
	}
	return nil
}
