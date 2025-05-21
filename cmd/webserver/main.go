package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gleipnir.technology/fieldseeker-sync-bridge"
	"gleipnir.technology/fieldseeker-sync-bridge/html"
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
		fmt.Println("Failed to init fssync: %v", err)
	}

	//html.InitializeTemplates()
	r.Get("/", index)
	r.Get("/service-request", serviceRequestList)
	log.Println("Serving web requests on :3000")
	http.ListenAndServe(":3000", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	count, err := fssync.ServiceRequestCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := html.PageDataIndex{
		ServiceRequestCount: count,
		Title:               "Gateway Sync Test",
	}

	html.Index(w, data)
}

func serviceRequestList(w http.ResponseWriter, r *http.Request) {
	requests, err := fssync.ServiceRequests()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.ServiceRequests(w, requests)
}
