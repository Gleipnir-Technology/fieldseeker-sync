package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gleipnir.technology/fieldseeker-sync-bridge"
)

func main() {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	err := fssync.Initialize()
	if err != nil {
		fmt.Println("Failed to initialize fssync: %v", err)
	}

	fssync.InitializeTemplates()
	r.Get("/", fssync.HandleIndex)
	http.ListenAndServe(":3000", r)
}
