package main

import (
	"errors"
	"fmt"
	"log"
	"os"

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
	config := fssync.ReadConfig()
	if len(config.Database.URL) == 0 {
		return errors.New("You must specify a database URL")
	}
	err = database.DoMigrations(config.Database.URL)
	if err != nil {
		fmt.Printf("Failed to migrate database: %v", err)
		os.Exit(1)
	}
	return nil
}
