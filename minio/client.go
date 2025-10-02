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

func NewClient(baseURL string, accessKeyID string, secretAccessKey string) *Client {
	// Initialize client
	minioClient, err := minio.New(baseURL, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return &Client{
		client: minioClient,
	}
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

func uploadFile(minioClient *minio.Client, bucketName string, filePath string) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Upload the file
	_, err = minioClient.FPutObject(context.Background(), bucketName, filePath, filePath, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("File uploaded successfully")
}
