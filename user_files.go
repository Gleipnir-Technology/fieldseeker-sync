package fssync

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
)

func AudioFileContentPathRaw(audioUUID string) string {
	config := ReadConfig()
	return fmt.Sprintf("%s/%s.m4a", config.UserFiles.Directory, audioUUID)
}
func AudioFileContentPathMp3(audioUUID string) string {
	config := ReadConfig()
	return fmt.Sprintf("%s/%s.mp3", config.UserFiles.Directory, audioUUID)
}
func AudioFileContentPathNormalized(audioUUID string) string {
	config := ReadConfig()
	return fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, audioUUID)
}
func AudioFileContentPathOgg(audioUUID string) string {
	config := ReadConfig()
	return fmt.Sprintf("%s/%s.ogg", config.UserFiles.Directory, audioUUID)
}
func AudioFileContentWrite(audioUUID uuid.UUID, body io.Reader) error {
	// Create file in configured directory
	filepath := AudioFileContentPathRaw(audioUUID.String())
	dst, err := os.Create(filepath)
	if err != nil {
		log.Printf("Failed to create audio file at %s: %v\n", filepath, err)
		return fmt.Errorf("Failed to create audio file at %s: %v", filepath, err)
	}
	defer dst.Close()

	// Copy rest of request body to file
	_, err = io.Copy(dst, body)
	if err != nil {
		return fmt.Errorf("Unable to save file to create audio file at %s: %v", filepath, err)
	}
	log.Printf("Saved audio content to %s\n", filepath)
	return nil
}
