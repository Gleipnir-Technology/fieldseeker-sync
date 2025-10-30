package main

import (
	"errors"
	"fmt"
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
	statistics := map[string]int{
		"mp3":        0,
		"normalized": 0,
		"ogg":        0,
		"raw":        0,
	}
	raw_missing := NewSet()

	for _, note := range notesAudio {
		paths := map[string]string{
			"mp3":        fssync.AudioFileContentPathMp3(note.UUID),
			"normalized": fssync.AudioFileContentPathNormalized(note.UUID),
			"ogg":        fssync.AudioFileContentPathOgg(note.UUID),
			"raw":        fssync.AudioFileContentPathRaw(note.UUID),
		}
		for name, path := range paths {
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				statistics[name] = statistics[name] + 1
				if name == "raw" {
					raw_missing.Add(note.UUID)
				}
			}
		}
	}
	fmt.Printf("Checked %d audio notes. %d (%f) are missing raw files.\n", len(notesAudio), statistics["raw"], statistics["raw"]/len(notesAudio))
	if raw_missing.Size() < 30 {
		fmt.Println("Missing raw files from:")
		for _, uuid := range raw_missing.list {
			fmt.Printf("\t%s\n", uuid)
		}
	}
	transcoded_missing := statistics["mp3"] + statistics["normalized"] + statistics["ogg"]
	if transcoded_missing > 0 {
		fmt.Printf("Additionally there are %d derivative files missing", transcoded_missing)
	}
	return nil
}
