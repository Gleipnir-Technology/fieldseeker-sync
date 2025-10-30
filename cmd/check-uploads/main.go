package main

import (
	"errors"
	"log"
	"os"
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

	notesAudio, err := database.NoteAudioQuery()
	if err != nil {
		log.Printf("Failed to query database: %v", err)
		os.Exit(2)
	}
	for _, note := range notesAudio {
		paths := map[string]string{
			"mp3":        fssync.AudioFileContentPathMp3(note.UUID),
			"normalized": fssync.AudioFileContentPathNormalized(note.UUID),
			"ogg":        fssync.AudioFileContentPathOgg(note.UUID),
			"raw":        fssync.AudioFileContentPathRaw(note.UUID),
		}
		for name, path := range paths {
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				log.Printf("%s audio file %s does not exist", name, path)
			}
		}
	}
	return nil
}
