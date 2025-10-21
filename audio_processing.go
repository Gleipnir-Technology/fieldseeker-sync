package fssync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/google/uuid"
)

// AudioJob represents a job to process an audio file.
type AudioJob struct {
	AudioUUID uuid.UUID
}

// audioJobChannel is the channel used to send audio processing jobs to the worker.
var audioJobChannel chan AudioJob

// StartAudioWorker initializes the audio job channel and starts the worker goroutine.
func StartAudioWorker(ctx context.Context) {
	audioJobChannel = make(chan AudioJob, 100) // Buffered channel to prevent blocking
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Audio worker shutting down.")
				return
			case job := <-audioJobChannel:
				log.Printf("Processing audio job for UUID: %s", job.AudioUUID)
				err := processAudioFile(job.AudioUUID)
				if err != nil {
					log.Printf("Error processing audio file %s: %v", job.AudioUUID, err)
				}
			}
		}
	}()
}

// EnqueueAudioJob sends an audio processing job to the worker.
func EnqueueAudioJob(job AudioJob) {
	select {
	case audioJobChannel <- job:
		log.Printf("Enqueued audio job for UUID: %s", job.AudioUUID)
	default:
		log.Printf("Audio job channel is full, dropping job for UUID: %s", job.AudioUUID)
	}
}

func processAudioFile(audioUUID uuid.UUID) error {
	// Normalize audio
	err := normalizeAudio(audioUUID)
	if err != nil {
		return fmt.Errorf("failed to normalize audio %s: %v", audioUUID, err)
	}

	// Transcode to OGG
	err = transcodeToOgg(audioUUID)
	if err != nil {
		return fmt.Errorf("failed to transcode audio %s to OGG: %v", audioUUID, err)
	}

	EnqueueLabelStudioJob(LabelStudioJob{
		UUID: audioUUID,
	})
	return nil
}

func normalizeAudio(audioUUID uuid.UUID) error {
	source := AudioFileContentPathRaw(audioUUID)
	_, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s doesn't exist, skipping normalization", source)
		return nil
	}
	log.Printf("Normalizing %s", source)
	destination := AudioFileContentPathNormalized(audioUUID)
	// Use "ffmpeg" directly, assuming it's in the system PATH
	cmd := exec.Command("ffmpeg", "-i", source, "-filter:a", "loudnorm", destination)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFmpeg output for normalization: %s", out)
		return fmt.Errorf("ffmpeg normalization failed: %v", err)
	}
	err = database.NoteAudioNormalized(audioUUID.String())
	if err != nil {
		return fmt.Errorf("failed to update database for normalized audio %s: %v", audioUUID, err)
	}
	log.Printf("Normalized audio to %s", destination)
	return nil
}

func transcodeToOgg(audioUUID uuid.UUID) error {
	source := AudioFileContentPathNormalized(audioUUID)
	_, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s doesn't exist, skipping OGG transcoding", source)
		return nil
	}
	log.Printf("Transcoding %s to ogg", source)
	destination := AudioFileContentPathOgg(audioUUID)
	// Use "ffmpeg" directly, assuming it's in the system PATH
	cmd := exec.Command("ffmpeg", "-i", source, "-vn", "-acodec", "libvorbis", destination)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFmpeg output for OGG transcoding: %s", out)
		return fmt.Errorf("ffmpeg OGG transcoding failed: %v", err)
	}
	err = database.NoteAudioTranscodedToOgg(audioUUID.String())
	if err != nil {
		return fmt.Errorf("failed to update database for OGG transcoded audio %s: %v", audioUUID, err)
	}
	log.Printf("Transcoded audio to %s", destination)
	return nil
}
