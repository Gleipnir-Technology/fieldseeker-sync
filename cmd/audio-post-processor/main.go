package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/getsentry/sentry-go"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return err
	}
	defer sentry.Flush(2 * time.Second)

	err = fssync.InitDB()
	if err != nil {
		log.Printf("Failed to init database: %v", err)
		os.Exit(1)
	}

	audioToNormalize, err := database.NoteAudioToNormalize()
	if err != nil {
		log.Printf("Failed to get notes to normalize: %v", err)
		os.Exit(2)
	}

	for _, a := range audioToNormalize {
		err = normalizeAudio(a.UUID)
		if err != nil {
			log.Printf("Failed to normalize %s: %v", a.UUID, err)
		}
	}

	audioToTranscodeToOgg, err := database.NoteAudioToTranscodeToOgg()
	if err != nil {
		log.Printf("Failed to get notes to normalize: %v", err)
		os.Exit(2)
	}
	for _, a := range audioToTranscodeToOgg {
		err = transcodeToOgg(a.UUID)
		if err != nil {
			log.Printf("Failed to transcode %s to OGG: %v", a.UUID, err)
		}
	}
	return nil
}

func normalizeAudio(uuid string) error {
	config := fssync.ReadConfig()
	source := fmt.Sprintf("%s/%s.m4a", config.UserFiles.Directory, uuid)
	_, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s doesn't exist, skipping", source)
		return nil
	}
	log.Printf("Normalizing %s", source)
	destination := fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, uuid)
	cmd := exec.Command("/run/current-system/sw/bin/ffmpeg", "-i", source, "-filter:a", "loudnorm", destination)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Out: %s", out)
		return err
	}
	err = database.NoteAudioNormalized(uuid)
	if err != nil {
		return err
	}
	log.Printf("Normalized to %s", destination)
	return nil
}

func transcodeToOgg(uuid string) error {
	config := fssync.ReadConfig()
	source := fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, uuid)
	_, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("%s doesn't exist, skipping", source)
		return nil
	}
	log.Printf("Transcoding %s to ogg", source)
	destination := fmt.Sprintf("%s/%s.ogg", config.UserFiles.Directory, uuid)
	cmd := exec.Command("/run/current-system/sw/bin/ffmpeg", "-i", source, "-vn", "-acodec", "libvorbis", destination)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Out: %s", out)
		return err
	}
	err = database.NoteAudioTranscodedToOgg(uuid)
	if err != nil {
		return err
	}
	log.Printf("Transcoded to %s", destination)
	return nil
	return nil
}
